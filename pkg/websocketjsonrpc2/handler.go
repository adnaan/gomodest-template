package websocketjsonrpc2

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/lithammer/shortuuid/v3"

	websocketjsonrpc2Sg "github.com/sourcegraph/jsonrpc2/websocket"

	"github.com/gorilla/websocket"
	"github.com/sourcegraph/jsonrpc2"
)

type opt struct {
	requestContextFunc func(r *http.Request) context.Context
	subscribeTopicFunc func(r *http.Request) *string
	upgrader           websocket.Upgrader
	resultHook         func(method string, result interface{}) interface{}
	onConnectMethod    string
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

func WithResultHook(resultHook func(method string, result interface{}) interface{}) Option {
	return func(o *opt) {
		o.resultHook = resultHook
	}
}

func WithOnConnectMethod(method string) Option {
	return func(o *opt) {
		o.onConnectMethod = method
	}
}

type Method func(ctx context.Context, params []byte) (interface{}, error)

type connHandler struct {
	requestContext context.Context
	methods        map[string]Method
	topic          string
	router         *router
	resultHook     func(method string, result interface{}) interface{}
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

	if h.resultHook != nil {
		result = h.resultHook(req.Method, result)
	}

	// also broadcast to other connections for the session
	connections, err := h.router.getTopicConnections(h.topic)
	if err != nil {
		return
	}

	for _, topicConn := range connections {
		topicConn := topicConn
		go func(conn *jsonrpc2.Conn) {
			if err := conn.Reply(ctx, req.ID, result); err != nil {
				log.Printf("conn for topic %s, reply err: %v\n", h.topic, err)
				return
			}
		}(topicConn)
	}
}

type Router interface {
	HandlerFunc(methods map[string]Method, options ...Option) http.HandlerFunc
}

func NewRouter() Router {
	return &router{
		topicConnections: make(map[string]map[string]*jsonrpc2.Conn),
	}
}

type router struct {
	topicConnections map[string]map[string]*jsonrpc2.Conn
	sync.RWMutex
}

func (ro *router) addConnection(topic, connID string, conn *jsonrpc2.Conn) {
	ro.Lock()
	defer ro.Unlock()
	_, ok := ro.topicConnections[topic]
	if !ok {
		// topic doesn't exit. create
		ro.topicConnections[topic] = make(map[string]*jsonrpc2.Conn)
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

func (ro *router) getTopicConnections(topic string) ([]*jsonrpc2.Conn, error) {
	ro.Lock()
	defer ro.Unlock()
	connMap, ok := ro.topicConnections[topic]
	if !ok {
		return nil, fmt.Errorf("topic doesn't exist")
	}
	var conns []*jsonrpc2.Conn
	for _, conn := range connMap {
		conns = append(conns, conn)
	}
	return conns, nil
}

func (ro *router) HandlerFunc(methods map[string]Method, options ...Option) http.HandlerFunc {
	m := &connHandler{methods: methods, router: ro}
	o := &opt{
		requestContextFunc: nil,
		upgrader:           websocket.Upgrader{},
	}

	for _, option := range options {
		option(o)
	}

	m.resultHook = o.resultHook

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
		jc := jsonrpc2.NewConn(ctx, websocketjsonrpc2Sg.NewObjectStream(c), m)
		connID := shortuuid.New()
		if topic != nil {
			ro.addConnection(*topic, connID, jc)
		}
		// onConnect
		if onConnectMethod, ok := methods[o.onConnectMethod]; ok {
			id := jsonrpc2.ID{
				Str:      o.onConnectMethod,
				IsString: true,
			}
			result, err := onConnectMethod(ctx, nil)
			if err != nil {
				err = jc.ReplyWithError(ctx, id, &jsonrpc2.Error{
					Code:    jsonrpc2.CodeInternalError,
					Message: err.Error(),
					Data:    nil,
				})
				if err != nil {
					log.Println("onConnectMethod, ReplyWithError err: ", err)
				}
				return
			}

			if o.resultHook != nil {
				result = o.resultHook(o.onConnectMethod, result)
			}

			if err := jc.Reply(ctx, id, result); err != nil {
				log.Printf("onConnectMethod %v, reply err: %v\n", o.onConnectMethod, err)
				return
			}
		}
		<-jc.DisconnectNotify()
		if topic != nil {
			ro.removeConnection(*topic, connID)
		}
	}
}
