package todos

import (
	"bytes"
	"context"
	"encoding/json"
	"gomodest-template/samples/todos/gen/models"
	"gomodest-template/samples/todos/gen/models/todo"
	"time"

	"github.com/google/uuid"
)

type TodosJsonRpc2 struct {
	DB *models.Client
}

type Query struct {
	Offset int `json:"offset"`
	Limit  int `json:"limit"`
}

func (t *TodosJsonRpc2) List(ctx context.Context, params []byte) (interface{}, error) {
	query := &Query{
		Offset: 0,
		Limit:  3,
	}
	err := json.NewDecoder(bytes.NewReader(params)).Decode(query)
	todos, err := t.DB.Todo.Query().Offset(query.Offset).Limit(query.Limit).All(ctx)
	if err != nil {
		return nil, err
	}
	return todos, nil
}

func (t *TodosJsonRpc2) Create(ctx context.Context, params []byte) (interface{}, error) {
	req := new(TodoRequest)
	err := json.NewDecoder(bytes.NewReader(params)).Decode(req)
	if err != nil {
		return nil, err
	}
	todo, err := t.DB.Todo.Create().
		SetStatus(todo.StatusInprogress).
		SetText(req.Text).
		Save(ctx)
	if err != nil {
		return nil, err
	}
	return todo, nil
}
func (t *TodosJsonRpc2) Update(ctx context.Context, params []byte) (interface{}, error) {
	req := new(TodoRequest)
	err := json.NewDecoder(bytes.NewReader(params)).Decode(req)
	if err != nil {
		return nil, err
	}
	uid, err := uuid.Parse(req.ID)
	if err != nil {
		return nil, err
	}

	todo, err := t.DB.Todo.
		UpdateOneID(uid).
		SetUpdatedAt(time.Now()).
		SetText(req.Text).
		Save(ctx)
	if err != nil {
		return nil, err
	}

	return todo, nil
}
func (t *TodosJsonRpc2) Delete(ctx context.Context, params []byte) (interface{}, error) {
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

	req.Text = ""
	return req, nil
}

func (t *TodosJsonRpc2) Get(ctx context.Context, params []byte) (interface{}, error) {
	req := new(TodoRequest)
	err := json.NewDecoder(bytes.NewReader(params)).Decode(req)
	if err != nil {
		return nil, err
	}

	uid, err := uuid.Parse(req.ID)
	if err != nil {
		return nil, err
	}
	todo, err := t.DB.Todo.Get(ctx, uid)
	if err != nil {
		return nil, err
	}

	return todo, nil
}
