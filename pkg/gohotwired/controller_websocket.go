package gohotwired

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"sync"

	"github.com/lithammer/shortuuid/v3"

	"github.com/Masterminds/sprig"
	"github.com/gorilla/websocket"
)

type Controller interface {
	NewView(page string, options ...ViewOption) http.HandlerFunc
}

type controlOpt struct {
	requestContextFunc func(r *http.Request) context.Context
	subscribeTopicFunc func(r *http.Request) *string
	upgrader           websocket.Upgrader
}
type ControllerOption func(*controlOpt)

func WithRequestContext(f func(r *http.Request) context.Context) ControllerOption {
	return func(o *controlOpt) {
		o.requestContextFunc = f
	}
}

func WithSubscribeTopic(f func(r *http.Request) *string) ControllerOption {
	return func(o *controlOpt) {
		o.subscribeTopicFunc = f
	}
}

func WithUpgrader(upgrader websocket.Upgrader) ControllerOption {
	return func(o *controlOpt) {
		o.upgrader = upgrader
	}
}

func WebsocketController(options ...ControllerOption) Controller {
	o := &controlOpt{
		requestContextFunc: nil,
		subscribeTopicFunc: func(r *http.Request) *string {
			challengeKey := r.Header.Get("Sec-Websocket-Key")
			topic := fmt.Sprintf("%s_%s",
				strings.Replace(r.URL.Path, "/", "_", -1), challengeKey)
			log.Println("new client subscribed to topic", topic)
			return &topic
		},
		upgrader: websocket.Upgrader{},
	}

	for _, option := range options {
		option(o)
	}
	return &websocketController{
		topicConnections: make(map[string]map[string]*websocket.Conn),
		controlOpt:       *o,
	}
}

type websocketController struct {
	controlOpt
	topicConnections map[string]map[string]*websocket.Conn
	sync.RWMutex
}

func (wc *websocketController) addConnection(topic, connID string, conn *websocket.Conn) {
	wc.Lock()
	defer wc.Unlock()
	_, ok := wc.topicConnections[topic]
	if !ok {
		// topic doesn't exit. create
		wc.topicConnections[topic] = make(map[string]*websocket.Conn)
	}
	wc.topicConnections[topic][connID] = conn
	log.Println("addConnection", topic, connID, len(wc.topicConnections[topic]))
}

func (wc *websocketController) removeConnection(topic, connID string) {
	wc.Lock()
	defer wc.Unlock()
	connMap, ok := wc.topicConnections[topic]
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
		delete(wc.topicConnections, topic)
	}

	log.Println("removeConnection", topic, connID, len(wc.topicConnections[topic]))
}

func (wc *websocketController) getTopicConnections(topic string) ([]*websocket.Conn, error) {
	wc.Lock()
	defer wc.Unlock()
	connMap, ok := wc.topicConnections[topic]
	if !ok {
		return nil, fmt.Errorf("topic doesn't exist")
	}
	var conns []*websocket.Conn
	for _, conn := range connMap {
		conns = append(conns, conn)
	}
	return conns, nil
}

func (wc *websocketController) NewView(page string, options ...ViewOption) http.HandlerFunc {
	o := &viewOpt{
		layout:            "./templates/layouts/index.html",
		layoutContentName: "content",
		partials:          []string{"./templates/partials"},
		extensions:        []string{".html", ".tmpl"},
		funcMap:           sprig.FuncMap(),
	}
	for _, option := range options {
		option(o)
	}

	// layout
	files := []string{o.layout}
	// global partials
	for _, p := range o.partials {
		files = append(files, find(p, o.extensions)...)
	}

	// page and its partials
	files = append(files, find(page, o.extensions)...)
	// contains: 1. layout 2. page  3. partials
	pageTemplate, err := template.New("").Funcs(o.funcMap).ParseFiles(files...)
	if err != nil {
		panic(fmt.Errorf("error parsing files err %v", err))
	}

	if ct := pageTemplate.Lookup(o.layoutContentName); ct == nil {
		panic(fmt.Errorf("err looking up layoutContent: the layout %s expects a template named %s",
			o.layout, o.layoutContentName))
	}

	if err != nil {
		panic(err)
	}
	var errorTemplate *template.Template
	if o.errorPage != "" {
		// layout
		errorFiles := []string{o.layout}
		// global partials
		for _, p := range o.partials {
			errorFiles = append(errorFiles, find(p, o.extensions)...)
		}
		// error page and its partials
		errorFiles = append(errorFiles, find(page, o.extensions)...)
		// contains: 1. layout 2. page  3. partials
		errorTemplate, err = template.New("").Funcs(o.funcMap).ParseFiles(errorFiles...)
		if err != nil {
			panic(fmt.Errorf("error parsing error page template err %v", err))
		}

		if ct := errorTemplate.Lookup(o.layoutContentName); ct == nil {
			panic(fmt.Errorf("err looking up layoutContent: the layout %s expects a template named %s",
				o.layout, o.layoutContentName))
		}
	}

	renderPage := func(w http.ResponseWriter, r *http.Request) {
		data := make(M)
		status := 200

		if o.onMountFunc != nil {
			status, data = o.onMountFunc(r)
		}

		w.WriteHeader(status)
		err = pageTemplate.ExecuteTemplate(w, filepath.Base(o.layout), data)
		if err != nil {
			if errorTemplate != nil {
				err = errorTemplate.ExecuteTemplate(w, filepath.Base(o.layout), nil)
				if err != nil {
					w.Write([]byte("something went wrong"))
				}
			} else {
				w.Write([]byte("something went wrong"))
			}
		}
	}

	handleSocket := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		if wc.requestContextFunc != nil {
			ctx = wc.requestContextFunc(r)
		}
		var topic *string
		if wc.subscribeTopicFunc != nil {
			topic = wc.subscribeTopicFunc(r)
		}

		c, err := wc.upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer c.Close()

		connID := shortuuid.New()
		if topic != nil {
			wc.addConnection(*topic, connID, c)
		}
	loop:
		for {
			mt, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				break loop
			}

			event := new(Event)
			err = json.NewDecoder(bytes.NewReader(message)).Decode(event)
			if err != nil {
				log.Printf("err: parsing event, msg %s \n", string(message))
				continue
			}

			if event.ID == "" {
				log.Printf("err: event %v, field event.id is required\n", event)
				continue
			}

			eventHandler, ok := o.eventHandlers[event.ID]
			if !ok {
				log.Printf("err: no handler found for event %s\n", event.ID)
				continue
			}

			stream := &WebsocketStream{
				event:        *event,
				conn:         c,
				rootTemplate: pageTemplate,
				messageType:  mt,
			}
			// unset any previously set errors
			stream.UnsetError()
			// handle event and write response
			eventHandler(ctx, stream)

			if len(stream.errs) != 0 {
				var errs []string
				for _, err := range stream.errs {
					errs = append(errs, err.Error())
				}
				log.Printf("err writing to connection %v\n", err)
				break loop
			}
		}

		if topic != nil {
			wc.removeConnection(*topic, connID)
		}
	}

	return func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Connection") == "Upgrade" && r.Header.Get("Upgrade") == "websocket" {
			handleSocket(w, r)
		} else {
			renderPage(w, r)
		}
	}
}
