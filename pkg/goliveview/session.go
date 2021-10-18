package goliveview

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"strings"

	"github.com/gorilla/websocket"
)

type ActionType string

const (
	Append  ActionType = "append"
	Prepend            = "prepend"
	Replace            = "replace"
	Update             = "update"
	Before             = "before"
	After              = "after"
	Remove             = "remove"
)

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
	Change(m map[string]interface{})
	Temporary(keys ...string)
	SessionStore
}

type session struct {
	changeRequest ChangeRequest
	conn          *websocket.Conn
	rootTemplate  *template.Template
	errs          []error
	messageType   int
	store         SessionStore
	temporaryKeys []string
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

	s.write(Replace, "gh-error", "", "gh-error",
		map[string]interface{}{
			"error": userMessage,
		})
}

func (s session) unsetError() {
	s.write(Replace, "gh-error", "", "gh-error", nil)
}

func (s session) write(action, target, targets, contentTemplate string, data map[string]interface{}) {
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
	var message string
	if targets != "" {
		message = fmt.Sprintf(turboTargetsWrapper, action, targets, buf.String())
	} else {
		message = fmt.Sprintf(turboTargetWrapper, action, target, buf.String())
	}

	err := s.conn.WriteMessage(s.messageType, []byte(message))
	if err != nil {
		s.errs = append(s.errs, err)
		return
	}
}

func (s session) Temporary(keys ...string) {
	s.temporaryKeys = append(s.temporaryKeys, keys...)
}

func (s session) Change(m map[string]interface{}) {
	// calculate change
	var changeTargets map[string]interface{}
	if s.changeRequest.Targets != "" {
		changeTargets = changeTargetsFromReq(s.changeRequest)
	} else {
		changeTargets = changeTargetFromReq(s.changeRequest)
	}

	changes := make(map[string]interface{})
	// from request
	for k, v := range changeTargets {
		changes[k] = v
	}
	// from handler
	for k, v := range m {
		changes[k] = v
	}

	var action, target, targets, contentTemplate string
	data := make(map[string]interface{})

	for k, v := range changes {
		if k == "action" {
			a, ok := v.(ActionType)
			if ok {
				action = string(a)
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
		delete(m, t)
	}
	// update store
	err := s.store.Set(m)
	if err != nil {
		log.Printf("error store.set %v\n", err)
	}
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
