package todos

import (
	"bytes"
	"context"
	"encoding/json"
	"gomodest-template/samples/todos/gen/models"
	"gomodest-template/samples/todos/gen/models/todo"
	"log"
	"net/http"

	"github.com/google/uuid"

	"github.com/gorilla/websocket"
	"github.com/sourcegraph/jsonrpc2"
	websocketjsonrpc2 "github.com/sourcegraph/jsonrpc2/websocket"
)

type MethodHandler func(ctx context.Context, params []byte) (interface{}, error)

type JSONRPC2Handler struct {
	methods map[string]MethodHandler
}

func (h *JSONRPC2Handler) Register(method string, handler MethodHandler) {
	h.methods[method] = handler
}

func (h *JSONRPC2Handler) Handle(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) {
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

type Todos2 struct {
	DB *models.Client
}

func (t *Todos2) List(ctx context.Context, params []byte) (interface{}, error) {
	todos, err := t.DB.Todo.Query().All(ctx)
	if err != nil {
		return nil, err
	}
	return todos, nil
}

func (t *Todos2) Add(ctx context.Context, params []byte) (interface{}, error) {
	req := new(TodoRequest)
	err := json.NewDecoder(bytes.NewReader(params)).Decode(req)
	if err != nil {
		return nil, err
	}
	_, err = t.DB.Todo.Create().
		SetStatus(todo.StatusInprogress).
		SetText(req.Text).
		Save(ctx)
	if err != nil {
		return nil, err
	}

	todos, err := t.DB.Todo.Query().All(ctx)
	if err != nil {
		return nil, err
	}
	return todos, nil
}
func (t *Todos2) Delete(ctx context.Context, params []byte) (interface{}, error) {
	req := new(TodoRequest)
	err := json.NewDecoder(bytes.NewReader(params)).Decode(req)
	if err != nil {
		return nil, err
	}

	uid, err := uuid.Parse(req.ID)
	if err != nil {
		return nil, err
	}
	err = t.DB.Todo.DeleteOneID(uid).Exec(ctx)
	if err != nil {
		return nil, err
	}

	todos, err := t.DB.Todo.Query().All(ctx)
	if err != nil {
		return nil, err
	}
	return todos, nil
}

func JSONRPC2HandlerFunc(db *models.Client) http.HandlerFunc {
	todos := Todos2{DB: db}
	ha := JSONRPC2Handler{methods: map[string]MethodHandler{}}
	ha.Register("list", todos.List)
	ha.Register("add", todos.Add)
	ha.Register("delete", todos.Delete)

	upgrader := websocket.Upgrader{ReadBufferSize: 1024, WriteBufferSize: 1024}
	return func(w http.ResponseWriter, r *http.Request) {
		done := make(chan struct{})
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer c.Close()
		jc := jsonrpc2.NewConn(r.Context(), websocketjsonrpc2.NewObjectStream(c), &ha)
		<-jc.DisconnectNotify()
		close(done)
	}
}
