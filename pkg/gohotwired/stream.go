package gohotwired

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

type Event struct {
	ID     string          `json:"id"`
	Target string          `json:"target"`
	Params json.RawMessage `json:"params"`
}

type EventHandler func(ctx context.Context, stream Stream)

type Stream interface {
	Event() Event
	UnsetError()
	Error(userMessage string, errs ...error)
	Append(target, content string, data M)
	Prepend(target, content string, data M)
	Replace(target, content string, data M)
	Update(target, content string, data M)
	Remove(target string)
	Before(target, content string, data M)
	After(target, content string, data M)
}
type WebsocketStream struct {
	event        Event
	conn         *websocket.Conn
	rootTemplate *template.Template
	errs         []error
	messageType  int
}

func (w *WebsocketStream) getStreamResponse(action, target, content string, data M) (string, error) {
	if target == "" {
		return "", fmt.Errorf("err target empty for action %s, content %s", action, content)
	}
	var buf bytes.Buffer
	if content != "" {
		err := w.rootTemplate.ExecuteTemplate(&buf, content, data)
		if err != nil {
			return "", fmt.Errorf("err %v,while executing content %s with data %v", err, content, data)
		}
	}
	var streamResponse string
	if strings.HasPrefix(target, ".") {
		streamResponse = fmt.Sprintf(turboTargetsWrapper, action, target, buf.String())
		return streamResponse, nil
	}

	return fmt.Sprintf(turboTargetWrapper, action, target, buf.String()), nil
}

func (w *WebsocketStream) write(action, target, content string, data M) {
	msg, err := w.getStreamResponse(action, target, content, data)
	if err != nil {
		log.Printf("warning: err creating stream response %v\n", err)
		return
	}
	err = w.conn.WriteMessage(w.messageType, []byte(msg))
	if err != nil {
		w.errs = append(w.errs, err)
	}
}

func (w *WebsocketStream) Event() Event {
	return w.event
}

func (w *WebsocketStream) UnsetError() {
	w.write("replace", "gh-error", "gh-error", nil)
}

func (w *WebsocketStream) Error(userMessage string, errs ...error) {
	if len(errs) != 0 {
		var errstrs []string
		for _, err := range errs {
			errstrs = append(errstrs, err.Error())
		}
		log.Printf("err: %v, errors: %v\n", userMessage, strings.Join(errstrs, ","))
	}
	w.write("replace", "gh-error", "gh-error", M{
		"error": userMessage,
	})
}

func (w *WebsocketStream) Append(target, content string, data M) {
	w.write("append", target, content, data)
}

func (w *WebsocketStream) Prepend(target, content string, data M) {
	w.write("prepend", target, content, data)
}

func (w *WebsocketStream) Replace(target, content string, data M) {
	w.write("replace", target, content, data)
}

func (w *WebsocketStream) Update(target, content string, data M) {
	w.write("update", target, content, data)
}

func (w *WebsocketStream) Remove(target string) {
	w.write("remove", target, "", M{})
}

func (w *WebsocketStream) Before(target, content string, data M) {
	w.write("before", target, content, data)
}

func (w *WebsocketStream) After(target, content string, data M) {
	w.write("after", target, content, data)
}

type StreamResponse struct {
	Action   string
	Target   string
	Targets  string
	Root     string
	Template string
	Data     map[string]interface{}
}

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
