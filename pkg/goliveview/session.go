package goliveview

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"strings"
	"time"

	"github.com/yosssi/gohtml"

	"github.com/gorilla/websocket"
)

type ActionType string

const (
	Append  ActionType = "append"
	Prepend ActionType = "prepend"
	Replace ActionType = "replace"
	Update  ActionType = "update"
	Before  ActionType = "before"
	After   ActionType = "after"
	Remove  ActionType = "remove"
)

var actions = map[string]int{
	"append":  0,
	"prepend": 0,
	"replace": 0,
	"update":  0,
	"before":  0,
	"after":   0,
	"remove":  0,
}

type ChangeRequest struct {
	ID              string          `json:"id"`
	Params          json.RawMessage `json:"params"`
	Action          ActionType      `json:"action"`
	Target          string          `json:"target,omitempty"`
	Targets         string          `json:"targets,omitempty"`
	ContentTemplate string          `json:"content_template"`
}

func (c ChangeRequest) DecodeParams(v interface{}) error {
	return json.NewDecoder(bytes.NewReader(c.Params)).Decode(v)
}

func ChangeTarget(action ActionType, target, contentTemplate string) map[string]interface{} {
	return map[string]interface{}{
		"action":           action,
		"target":           target,
		"content_template": contentTemplate,
	}
}

func ChangeTargets(action ActionType, targets, contentTemplate string) map[string]interface{} {
	return map[string]interface{}{
		"action":           action,
		"targets":          targets,
		"content_template": contentTemplate,
	}
}

func changeTargetFromReq(c ChangeRequest) map[string]interface{} {
	return map[string]interface{}{
		"action":           c.Action,
		"target":           c.Target,
		"content_template": c.ContentTemplate,
	}
}

func changeTargetsFromReq(c ChangeRequest) map[string]interface{} {
	return map[string]interface{}{
		"action":           c.Action,
		"targets":          c.Targets,
		"content_template": c.ContentTemplate,
	}
}

type SessionStore interface {
	Set(m map[string]interface{}) error
	Get(key string) (interface{}, bool)
}

type Session interface {
	Change(changeset map[string]interface{})
	Flash(duration time.Duration, changeset map[string]interface{})
	Temporary(keys ...string)
	SessionStore
}

type session struct {
	rootTemplate         *template.Template
	topic                string
	changeRequest        ChangeRequest
	conns                map[string]*websocket.Conn
	messageType          int
	store                SessionStore
	temporaryKeys        []string
	enableHTMLFormatting bool
}

func (s session) setError(userMessage string, errs ...error) {
	if len(errs) != 0 {
		var errstrs []string
		for _, err := range errs {
			if err == nil {
				continue
			}
			errstrs = append(errstrs, err.Error())
		}
		log.Printf("err: %v, errors: %v\n", userMessage, strings.Join(errstrs, ","))
	}

	s.write(Replace, "glw-error", "", "glw-error",
		map[string]interface{}{
			"error": userMessage,
		})
}

func (s session) unsetError() {
	s.write(Replace, "glw-error", "", "glw-error", nil)
}

func (s session) write(action ActionType, target, targets, contentTemplate string, data map[string]interface{}) {
	if action == "" {
		log.Printf("err action is empty\n")
		return
	}
	// stream response
	if target == "" && targets == "" {
		log.Printf("err target/targets %s/%s empty for changeRequest %+v\n", target, targets, s.changeRequest)
		return
	}
	var buf bytes.Buffer
	if contentTemplate != "" && action != Remove {
		err := s.rootTemplate.ExecuteTemplate(&buf, contentTemplate, data)
		if err != nil {
			log.Printf("err %v,while executing template for changeRequest %+v\n", err, s.changeRequest)
			return
		}
	}
	html := buf.String()
	var message string
	if targets != "" {
		message = fmt.Sprintf(turboTargetsWrapper, action, targets, html)
	} else {
		message = fmt.Sprintf(turboTargetWrapper, action, target, html)
	}

	if s.enableHTMLFormatting {
		message = gohtml.Format(message)
	}

	for topic, conn := range s.conns {
		err := conn.WriteMessage(s.messageType, []byte(message))
		if err != nil {
			log.Printf("err writing message for topic:%v, %v, closing conn", topic, err)
			conn.Close()
			return
		}
	}
}

func (s session) Temporary(keys ...string) {
	s.temporaryKeys = append(s.temporaryKeys, keys...)
}

func (s session) change(changeset map[string]interface{}) {
	// calculate change
	var changeTargets map[string]interface{}
	if s.changeRequest.Targets != "" {
		changeTargets = changeTargetsFromReq(s.changeRequest)
	} else {
		changeTargets = changeTargetFromReq(s.changeRequest)
	}

	mergedChangeset := make(map[string]interface{})

	// from request
	for k, v := range changeTargets {
		mergedChangeset[k] = v
	}

	// from handler
	for k, v := range changeset {
		mergedChangeset[k] = v
	}

	var action ActionType
	var target, targets, contentTemplate string
	data := make(map[string]interface{})

	for k, v := range mergedChangeset {

		if k == "action" {
			if a, ok := v.(ActionType); ok {
				action = a
				continue
			}

			if a, ok := v.(string); ok {
				if _, ok := actions[a]; ok {
					action = ActionType(a)
				}
				continue
			}

		}
		if k == "target" {
			target = v.(string)
			continue
		}
		if k == "targets" {
			targets = v.(string)
			continue
		}
		if k == "content_template" {
			contentTemplate = v.(string)
			continue
		}
		data[k] = v
	}

	s.write(action, target, targets, contentTemplate, data)

	// delete keys which are marked temporary
	for _, t := range s.temporaryKeys {
		delete(changeset, t)
	}
	// update store
	err := s.store.Set(changeset)
	if err != nil {
		log.Printf("error store.set %v\n", err)
	}
}

func (s session) Flash(duration time.Duration, changeset map[string]interface{}) {
	nilDataChangeSet := ChangeTarget(Replace, "glw-flash", "glw-flash")
	if _, ok := changeset["action"]; !ok {
		changeset["action"] = nilDataChangeSet["action"]
	}
	if _, ok := changeset["target"]; !ok {
		changeset["target"] = nilDataChangeSet["target"]
	}
	if _, ok := changeset["content_template"]; !ok {
		changeset["content_template"] = nilDataChangeSet["content_template"]
	}

	s.change(changeset)
	go func() {
		time.Sleep(duration)
		s.change(nilDataChangeSet)
	}()
}

func (s session) Change(changeset map[string]interface{}) {
	s.change(changeset)
}

func (s session) Set(m map[string]interface{}) error {
	return s.store.Set(m)
}

func (s session) Get(key string) (interface{}, bool) {
	return s.store.Get(key)
}

type ChangeRequestHandler func(ctx context.Context, req ChangeRequest, session Session) error

var turboTargetWrapper = `{
							"message":
							  "<turbo-stream action="%s" target="%s">
								<template>
									%s
								</template>
							   </turbo-stream>"
						  }`

var turboTargetsWrapper = `{
							"message":
							  "<turbo-stream action="%s" targets="%s">
								<template>
									%s
								</template>
							   </turbo-stream>"
						  }`
