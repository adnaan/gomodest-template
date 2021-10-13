package gohotwired

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
	"sync"

	"github.com/Masterminds/sprig"

	"github.com/lithammer/shortuuid/v3"

	"github.com/gorilla/websocket"
)

type Handler func(ctx context.Context, event StreamEvent) (*StreamResponse, error)
type StreamEvent struct {
	Event  string          `json:"event"`
	Target string          `json:"target"`
	Params json.RawMessage `json:"params"`
}

type StreamResponse struct {
	Action   string
	Target   string
	Targets  string
	Root     string
	Template string
	Data     map[string]interface{}
}

type opt struct {
	requestContextFunc func(r *http.Request) context.Context
	subscribeTopicFunc func(r *http.Request) *string
	upgrader           websocket.Upgrader
	errorTarget        string
	errorTargets       string
}

type Option func(*opt)

func WithRequestContext(f func(r *http.Request) context.Context) Option {
	return func(o *opt) {
		o.requestContextFunc = f
	}
}

func WithSubscribeTopic(f func(r *http.Request) *string) Option {
	return func(o *opt) {
		o.subscribeTopicFunc = f
	}
}

func WithUpgrader(upgrader websocket.Upgrader) Option {
	return func(o *opt) {
		o.upgrader = upgrader
	}
}

func WithErrorTarget(errorTarget string) Option {
	return func(o *opt) {
		o.errorTarget = errorTarget
	}
}

func WithErrorTargets(errorTargets string) Option {
	return func(o *opt) {
		o.errorTargets = errorTargets
	}
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

type connHandler struct {
	requestContext context.Context
	handlers       map[string]Handler
	topic          string
	router         *router
	templates      map[string]*template.Template
}

func (h *connHandler) Handle(ctx context.Context, msg []byte) ([]byte, error) {
	streamEvent := new(StreamEvent)
	err := json.NewDecoder(bytes.NewReader(msg)).Decode(streamEvent)
	if err != nil {
		return nil, err
	}
	if streamEvent.Event == "" {
		return nil, fmt.Errorf("field event is required")
	}
	if streamEvent.Event == "" {
		return nil, fmt.Errorf("field target is required")
	}
	handler, ok := h.handlers[streamEvent.Event]
	if !ok {
		return nil, fmt.Errorf("no handler found for event %s", streamEvent.Event)
	}

	response, err := handler(ctx, *streamEvent)
	if err != nil {
		return nil, err
	}
	if response == nil {
		return nil, fmt.Errorf("no response from event handler")
	}

	var t *template.Template
	var buf bytes.Buffer
	if response.Root != "" {
		if response.Template == "" {
			return nil, fmt.Errorf("response has no templatePath")
		}

		fileInfo, err := ioutil.ReadDir(response.Root)
		if err != nil {
			return nil, err
		}
		var partials []string
		for _, file := range fileInfo {
			partials = append(partials, fmt.Sprintf("%s/%s", response.Root, file.Name()))
		}
		baseName := filepath.Base(response.Root)
		t, ok = h.templates[response.Root]
		if !ok {
			t, err = template.New(baseName).Funcs(sprig.FuncMap()).ParseFiles(partials...)
			if err != nil {
				return nil, err
			}
		}
		err = t.ExecuteTemplate(&buf, response.Template, response.Data)
		if err != nil {
			return nil, err
		}
	}

	streamResponse := fmt.Sprintf(turboTargetWrapper, response.Action, response.Target, buf.String())
	if response.Targets != "" {
		streamResponse = fmt.Sprintf(turboTargetsWrapper, response.Action, response.Targets, buf.String())
	}

	return []byte(streamResponse), nil
}

type Router interface {
	HandlerFunc(methods map[string]Handler, options ...Option) http.HandlerFunc
}

func NewRouter() Router {
	return &router{
		topicConnections: make(map[string]map[string]*websocket.Conn),
	}
}

type router struct {
	topicConnections map[string]map[string]*websocket.Conn
	sync.RWMutex
}

func (ro *router) addConnection(topic, connID string, conn *websocket.Conn) {
	ro.Lock()
	defer ro.Unlock()
	_, ok := ro.topicConnections[topic]
	if !ok {
		// topic doesn't exit. create
		ro.topicConnections[topic] = make(map[string]*websocket.Conn)
	}
	ro.topicConnections[topic][connID] = conn
	log.Println("addConnection", topic, connID, len(ro.topicConnections[topic]))
}

func (ro *router) removeConnection(topic, connID string) {
	ro.Lock()
	defer ro.Unlock()
	connMap, ok := ro.topicConnections[topic]
	if !ok {
		return
	}
	// delete connection from topic
	_, ok = connMap[connID]
	if ok {
		delete(connMap, connID)
	}
	// no connections for the topic, remove it
	if len(connMap) == 0 {
		delete(ro.topicConnections, topic)
	}

	log.Println("removeConnection", topic, connID, len(ro.topicConnections[topic]))
}

func (ro *router) getTopicConnections(topic string) ([]*websocket.Conn, error) {
	ro.Lock()
	defer ro.Unlock()
	connMap, ok := ro.topicConnections[topic]
	if !ok {
		return nil, fmt.Errorf("topic doesn't exist")
	}
	var conns []*websocket.Conn
	for _, conn := range connMap {
		conns = append(conns, conn)
	}
	return conns, nil
}

func (ro *router) HandlerFunc(handlers map[string]Handler, options ...Option) http.HandlerFunc {
	m := &connHandler{handlers: handlers, router: ro}
	o := &opt{
		requestContextFunc: nil,
		upgrader:           websocket.Upgrader{},
	}

	for _, option := range options {
		option(o)
	}

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		if o.requestContextFunc != nil {
			ctx = o.requestContextFunc(r)
		}
		var topic *string
		if o.subscribeTopicFunc != nil {
			topic = o.subscribeTopicFunc(r)
			if topic != nil {
				m.topic = *topic
			}
		}

		c, err := o.upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer c.Close()

		connID := shortuuid.New()
		if topic != nil {
			ro.addConnection(*topic, connID, c)
		}
	loop:
		for {
			mt, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				break loop
			}
			resp, err := m.Handle(ctx, message)
			if err != nil {
				log.Println("read:", err)
				continue
			}

			err = c.WriteMessage(mt, resp)
			if err != nil {
				log.Println("write:", err)
				break loop
			}
		}

		if topic != nil {
			ro.removeConnection(*topic, connID)
		}
	}
}
