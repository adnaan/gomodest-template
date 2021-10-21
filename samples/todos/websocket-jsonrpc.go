package todos

import (
	"context"
	"encoding/json"
	"gomodest-template/samples/todos/gen/models"
	"gomodest-template/samples/todos/gen/models/todo"
	"log"
	"net/http"
	"net/rpc"
	"net/rpc/jsonrpc"

	"golang.org/x/net/websocket"

	"github.com/google/uuid"
)

type Params struct {
	Payload json.RawMessage `json:"payload"`
}

type TodoRequest struct {
	ID       string `json:"id"`
	Text     string `json:"text"`
	Redirect bool   `json:"redirect,omitempty"`
}

type Todos struct {
	DB  *models.Client
	Ctx context.Context
}

func (t *Todos) list() ([]byte, error) {
	todos, err := t.DB.Todo.Query().All(t.Ctx)
	if err != nil {
		return nil, err
	}
	data, err := json.Marshal(todos)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (t *Todos) List(_ []struct{}, reply *Params) error {
	log.Println("Todos.List called")
	data, err := t.list()
	if err != nil {
		return err
	}
	reply.Payload = data
	return nil
}

func (t *Todos) Add(req TodoRequest, reply *Params) error {
	log.Println("Todos.Add called")
	t.DB.Todo.Create().
		SetStatus(todo.StatusInprogress).
		SetText(req.Text).
		Save(t.Ctx)

	data, err := t.list()
	if err != nil {
		return err
	}
	reply.Payload = data
	return nil
}

func (t *Todos) Delete(req TodoRequest, reply *Params) error {
	log.Println("Todos.Delete called")
	uid, err := uuid.Parse(req.ID)
	if err != nil {
		return err
	}

	err = t.DB.Todo.DeleteOneID(uid).Exec(t.Ctx)
	if err != nil {
		return err
	}

	data, err := t.list()
	if err != nil {
		return err
	}
	reply.Payload = data
	return nil
}

func StartRPCServer(db *models.Client, ctx context.Context) {
	rpc.Register(&Todos{DB: db, Ctx: ctx})
	http.Handle("/", websocket.Handler(func(conn *websocket.Conn) {
		jsonrpc.ServeConn(conn)
	}))
	go http.ListenAndServe("localhost:3001", nil)
}
