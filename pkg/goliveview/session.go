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

	"github.com/lithammer/shortuuid/v3"

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

type M map[string]interface{}

type ChangeRequest struct {
	ID       string          `json:"id"`
	Params   json.RawMessage `json:"params"`
	Action   ActionType      `json:"action"`
	Target   string          `json:"target,omitempty"`
	Targets  string          `json:"targets,omitempty"`
	Template string          `json:"template"`
}

func (c ChangeRequest) DecodeParams(v interface{}) error {
	return json.NewDecoder(bytes.NewReader(c.Params)).Decode(v)
}

func ChangeTarget(action ActionType, target, template string) M {
	return M{
		"action":   action,
		"target":   target,
		"template": template,
	}
}

func ChangeTargets(action ActionType, targets, template string) M {
	return M{
		"action":   action,
		"targets":  targets,
		"template": template,
	}
}

func changeTargetFromReq(c ChangeRequest) M {
	return M{
		"action":   c.Action,
		"target":   c.Target,
		"template": c.Template,
	}
}

func changeTargetsFromReq(c ChangeRequest) M {
	return M{
		"action":   c.Action,
		"targets":  c.Targets,
		"template": c.Template,
	}
}

type SessionStore interface {
	Set(m M) error
	Get(key string) (interface{}, bool)
}

type Session interface {
	Change(changeset M)
	Flash(duration time.Duration, changeset M)
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

	s.write(Replace, "glv-error", "", "glv-error",
		M{
			"error": userMessage,
		})
}

func (s session) unsetError() {
	s.write(Replace, "glv-error", "", "glv-error", nil)
}

func (s session) write(action ActionType, target, targets, template string, data M) {
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
	if template != "" && action != Remove {
		err := s.rootTemplate.ExecuteTemplate(&buf, template, data)
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

func (s session) change(changeset M) {
	// calculate change
	var changeTargets M
	if s.changeRequest.Targets != "" {
		changeTargets = changeTargetsFromReq(s.changeRequest)
	} else {
		changeTargets = changeTargetFromReq(s.changeRequest)
	}

	mergedChangeset := make(M)

	// from request
	for k, v := range changeTargets {
		mergedChangeset[k] = v
	}

	// from handler
	for k, v := range changeset {
		mergedChangeset[k] = v
	}

	var action ActionType
	var target, targets, template string
	data := make(M)

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
		if k == "template" {
			template = v.(string)
			continue
		}
		data[k] = v
	}

	s.write(action, target, targets, template, data)

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

func (s session) Flash(duration time.Duration, changeset M) {
	nilDataChangeSet := ChangeTarget(Append, "glv-flash", "glv-flash-message")
	if _, ok := changeset["action"]; !ok {
		changeset["action"] = nilDataChangeSet["action"]
	}
	if _, ok := changeset["target"]; !ok {
		changeset["target"] = nilDataChangeSet["target"]
	}
	if _, ok := changeset["template"]; !ok {
		changeset["template"] = nilDataChangeSet["template"]
	}

	flashID := shortuuid.New()
	changeset["flash_id"] = flashID

	s.change(changeset)
	go func(target string, changeset M) {
		time.Sleep(duration)
		nilDataChangeSet["action"] = Remove
		nilDataChangeSet["target"] = flashID
		s.change(nilDataChangeSet)
	}(flashID, nilDataChangeSet)
}

func (s session) Change(changeset M) {
	s.change(changeset)
}

func (s session) Set(m M) error {
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
