package websocketjsonrpc2

import (
	"context"
	"log"
	"net/http"

	"github.com/lithammer/shortuuid/v3"

	websocketjsonrpc2Sg "github.com/sourcegraph/jsonrpc2/websocket"

	"github.com/gorilla/websocket"
	"github.com/sourcegraph/jsonrpc2"
)

type opt struct {
	requestContextFunc func(r *http.Request) context.Context
	sessionKeyFunc     func(r *http.Request) *string
	upgrader           websocket.Upgrader
}

type Option func(*opt)

func WithRequestContext(f func(r *http.Request) context.Context) Option {
	return func(o *opt) {
		o.requestContextFunc = f
	}
}

func WithSessionKey(f func(r *http.Request) *string) Option {
	return func(o *opt) {
		o.sessionKeyFunc = f
	}
}

func WithUpgrader(upgrader websocket.Upgrader) Option {
	return func(o *opt) {
		o.upgrader = upgrader
	}
}

type Method func(ctx context.Context, params []byte) (interface{}, error)

type connHandler struct {
	requestContext     context.Context
	methods            map[string]Method
	sessionKey         string
	sessionConnections map[string]map[string]*jsonrpc2.Conn
}

func (h *connHandler) Handle(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) {
	method, ok := h.methods[req.Method]

	if !ok {
		err := conn.ReplyWithError(ctx, req.ID, &jsonrpc2.Error{
			Code:    jsonrpc2.CodeMethodNotFound,
			Message: "method not found",
			Data:    nil,
		})
		if err != nil {
			log.Println("ReplyWithError err: ", err)
		}
		return
	}
	var params []byte
	if req.Params != nil {
		params = *req.Params
	}
	result, err := method(ctx, params)
	if err != nil {
		err = conn.ReplyWithError(ctx, req.ID, &jsonrpc2.Error{
			Code:    jsonrpc2.CodeInternalError,
			Message: err.Error(),
			Data:    nil,
		})
		if err != nil {
			log.Println("ReplyWithError err: ", err)
		}
		return
	}

	if err := conn.Reply(ctx, req.ID, result); err != nil {
		log.Println("reply err: ", err)
		return
	}
}

type Router interface {
	HandlerFunc(methods map[string]Method, options ...Option) http.HandlerFunc
	RegisterSession(sessionKey string) error
	UnregisterSession(sessionKey string) error
}

func NewRouter() Router {
	return &router{
		sessionConnections: make(map[string]map[string]*jsonrpc2.Conn),
	}
}

type router struct {
	sessionConnections map[string]map[string]*jsonrpc2.Conn
}

func (ro *router) RegisterSession(sessionKey string) error {
	_, ok := ro.sessionConnections[sessionKey]
	if !ok {
		ro.sessionConnections[sessionKey] = make(map[string]*jsonrpc2.Conn)
	}
	return nil
}

func (ro *router) UnregisterSession(sessionKey string) error {
	delete(ro.sessionConnections, sessionKey)
	return nil
}

func (ro *router) HandlerFunc(methods map[string]Method, options ...Option) http.HandlerFunc {
	m := &connHandler{methods: methods, sessionConnections: ro.sessionConnections}
	o := &opt{
		requestContextFunc: nil,
		upgrader:           websocket.Upgrader{ReadBufferSize: 1024, WriteBufferSize: 1024},
	}

	for _, option := range options {
		option(o)
	}

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		if o.requestContextFunc != nil {
			ctx = o.requestContextFunc(r)
		}
		var sessionKey *string
		if o.sessionKeyFunc != nil {
			sessionKey = o.sessionKeyFunc(r)
			if sessionKey != nil {
				m.sessionKey = *sessionKey
			}
		}

		c, err := o.upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer c.Close()
		jc := jsonrpc2.NewConn(ctx, websocketjsonrpc2Sg.NewObjectStream(c), m)
		connID := shortuuid.New()
		if sessionKey != nil {
			// check if session already exists
			connMap, ok := ro.sessionConnections[*sessionKey]
			if ok {
				connMap[connID] = jc
				//ro.sessionConnections[*sessionKey] = connMap
			}

		}
		<-jc.DisconnectNotify()
		if sessionKey != nil {
			// check if session already exists
			connMap, ok := ro.sessionConnections[*sessionKey]
			if ok {
				delete(connMap, connID)
				//ro.sessionConnections[*sessionKey] = connMap
			}
		}
	}
}
