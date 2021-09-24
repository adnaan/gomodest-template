package websocketjsonrpc2

import (
	"context"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/sourcegraph/jsonrpc2"
	websocketjsonrpc2Sg "github.com/sourcegraph/jsonrpc2/websocket"
)

type opt struct {
	requestContextFunc func(r *http.Request) context.Context
	upgrader           websocket.Upgrader
}

type Option func(*opt)

func WithRequestContext(f func(r *http.Request) context.Context) Option {
	return func(o *opt) {
		o.requestContextFunc = f
	}
}

func WithUpgrader(upgrader websocket.Upgrader) Option {
	return func(o *opt) {
		o.upgrader = upgrader
	}
}

type Method func(ctx context.Context, params []byte) (interface{}, error)

type mux struct {
	requestContext context.Context
	methods        map[string]Method
}

func (h *mux) Handle(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) {
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

func HandlerFunc(methods map[string]Method, options ...Option) http.HandlerFunc {
	m := &mux{methods: methods}
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
		done := make(chan struct{})
		c, err := o.upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer c.Close()
		jc := jsonrpc2.NewConn(ctx, websocketjsonrpc2Sg.NewObjectStream(c), m)
		<-jc.DisconnectNotify()
		close(done)
	}
}
