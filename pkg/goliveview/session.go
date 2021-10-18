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

type KV struct {
	K    string      `json:"k"`
	V    interface{} `json:"v"`
	Temp bool        `json:"temp"` // do not store in session
}

func Action(v string) KV {
	return KV{
		K:    "action",
		V:    v,
		Temp: true,
	}
}

func Target(v string) KV {
	return KV{
		K:    "target",
		V:    v,
		Temp: true,
	}
}

func Targets(v string) KV {
	return KV{
		K:    "targets",
		V:    v,
		Temp: true,
	}
}

func ContentTemplate(v string) KV {
	return KV{
		K:    "content_template",
		V:    v,
		Temp: true,
	}
}

func ChangeTarget(action, target, contentTemplate string) []KV {
	return []KV{
		Action(action),
		Target(target),
		ContentTemplate(contentTemplate),
	}
}

func ChangeTargets(action, targets, contentTemplate string) []KV {
	return []KV{
		Action(action),
		Targets(targets),
		ContentTemplate(contentTemplate),
	}
}

func changeTargetFromReq(c ChangeRequest) []KV {
	return []KV{
		Action(string(c.Action)),
		Target(c.Target),
		ContentTemplate(c.ContentTemplate),
	}
}

func changeTargetsFromReq(c ChangeRequest) []KV {
	return []KV{
		Action(string(c.Action)),
		Targets(c.Targets),
		ContentTemplate(c.ContentTemplate),
	}
}

type SessionStore interface {
	Set(kvs ...KV) error
	Get(key string) (interface{}, bool)
}

type Session interface {
	Change(kvs ...KV)
	SessionStore
}

type session struct {
	changeRequest ChangeRequest
	conn          *websocket.Conn
	rootTemplate  *template.Template
	errs          []error
	messageType   int
	store         SessionStore
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

func (s session) Change(kvs ...KV) {
	// calculate change
	var changeTargets []KV
	if s.changeRequest.Targets != "" {
		changeTargets = changeTargetsFromReq(s.changeRequest)
	} else {
		changeTargets = changeTargetFromReq(s.changeRequest)
	}

	changes := make(map[string]interface{})
	// from request
	for _, kv := range changeTargets {
		changes[kv.K] = kv.V
	}
	// from handler
	for _, kv := range kvs {
		changes[kv.K] = kv.V
	}

	var action, target, targets, contentTemplate string
	data := make(map[string]interface{})

	for k, v := range changes {
		if k == "action" {
			action = v.(string)
			continue
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

	// update
	var persistKVs []KV
	for _, kv := range kvs {
		if kv.Temp {
			continue
		}
		persistKVs = append(persistKVs, kv)
	}
	err := s.store.Set(persistKVs...)
	if err != nil {
		log.Printf("error store.set %v\n", err)
	}
}

func (s session) Set(kvs ...KV) error {
	return s.store.Set(kvs...)
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
