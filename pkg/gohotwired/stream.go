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

type Action string

const (
	Append  Action = "append"
	Prepend        = "prepend"
	Replace        = "replace"
	Update         = "update"
	Before         = "before"
	After          = "after"
)

type Event struct {
	ID      string          `json:"id"`
	Action  Action          `json:"action"`
	Target  string          `json:"target"`
	Content string          `json:"content"`
	Params  json.RawMessage `json:"params"` // incoming parameters from the client. unused when data is sent from the server
	Data    interface{}     `json:"-"`      // outgoing data from the server
}

type EventHandler func(ctx context.Context, stream Stream) error

type Stream interface {
	// Event contains the incoming event details
	Event() Event
	// DecodeParams decodes Event.Params in the provided pointer type
	DecodeParams(v interface{}) error
	// Echo sends a turbo-stream html partial using the incoming event's action, target and content values
	Echo(data interface{})
	// Send a turbo-stream html partial
	Send(event Event)
}
type WebsocketStream struct {
	event        Event
	conn         *websocket.Conn
	rootTemplate *template.Template
	errs         []error
	messageType  int
}

func (w *WebsocketStream) getStreamResponse(e Event) (string, error) {
	if e.Target == "" {
		return "", fmt.Errorf("err target empty for event %+v\n", e)
	}
	var buf bytes.Buffer
	if e.Content != "" {
		err := w.rootTemplate.ExecuteTemplate(&buf, e.Content, e.Data)
		if err != nil {
			return "", fmt.Errorf("err %v,while executing template for event %+v\n", err, e)
		}
	}
	var streamResponse string
	if strings.HasPrefix(e.Target, ".") {
		streamResponse = fmt.Sprintf(turboTargetsWrapper, e.Action, e.Target, buf.String())
		return streamResponse, nil
	}

	return fmt.Sprintf(turboTargetWrapper, e.Action, e.Target, buf.String()), nil
}

func (w *WebsocketStream) write(e Event) {
	msg, err := w.getStreamResponse(e)
	if err != nil {
		log.Printf("warning: err creating stream response %v\n", err)
		return
	}
	err = w.conn.WriteMessage(w.messageType, []byte(msg))
	if err != nil {
		w.errs = append(w.errs, err)
	}
}

func (w *WebsocketStream) unsetError() {
	w.write(Event{
		Action:  Replace,
		Target:  "gh-error",
		Content: "gh-error",
	})
}

func (w *WebsocketStream) error(userMessage string, errs ...error) {
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

	w.write(Event{
		Action:  Replace,
		Target:  "gh-error",
		Content: "gh-error",
		Data: map[string]interface{}{
			"error": userMessage,
		},
	})
}

func (w *WebsocketStream) Event() Event {
	return w.event
}

func (w *WebsocketStream) DecodeParams(v interface{}) error {
	return json.NewDecoder(bytes.NewReader(w.event.Params)).Decode(v)
}

func (w *WebsocketStream) Echo(data interface{}) {
	w.write(Event{
		Action:  w.event.Action,
		Target:  w.event.Target,
		Content: w.event.Content,
		Data:    data,
	})
}

func (w *WebsocketStream) Send(e Event) {
	w.write(e)
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
