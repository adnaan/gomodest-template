package todos

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	gh "gomodest-template/pkg/gohotwired"
	"gomodest-template/samples/todos/gen/models"
	"gomodest-template/samples/todos/gen/models/todo"
	"time"

	"github.com/google/uuid"
)

type Stream struct {
	DB *models.Client
}

func (t *Stream) List(ctx context.Context, e gh.StreamEvent) (*gh.StreamResponse, error) {
	query := &Query{
		Offset: 0,
		Limit:  3,
	}
	err := json.NewDecoder(bytes.NewReader(e.Params)).Decode(query)
	todos, err := t.DB.Todo.
		Query().
		Offset(query.Offset).
		Limit(query.Limit).
		Order(models.Desc(todo.FieldUpdatedAt)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	return &gh.StreamResponse{
		Action:   "update",
		Target:   e.Target,
		Root:     "templates/partials",
		Template: "todos",
		Data: map[string]interface{}{
			"todos": todos,
		},
	}, nil
}

func (t *Stream) Create(ctx context.Context, e gh.StreamEvent) (*gh.StreamResponse, error) {
	req := new(TodoRequest)
	err := json.NewDecoder(bytes.NewReader(e.Params)).Decode(req)
	if err != nil {
		return nil, err
	}
	if len(req.Text) < 3 {
		return nil, fmt.Errorf("minimum text size is 4")
	}
	todo, err := t.DB.Todo.Create().
		SetStatus(todo.StatusInprogress).
		SetText(req.Text).
		Save(ctx)
	if err != nil {
		return nil, err
	}
	return &gh.StreamResponse{
		Action:   "append",
		Target:   e.Target,
		Root:     "templates/partials",
		Template: "todo",
		Data: map[string]interface{}{
			"ID":   todo.ID,
			"Text": todo.Text,
		},
	}, nil
}
func (t *Stream) Update(ctx context.Context, params []byte) (interface{}, error) {
	req := new(TodoRequest)
	err := json.NewDecoder(bytes.NewReader(params)).Decode(req)
	if err != nil {
		return nil, err
	}
	uid, err := uuid.Parse(req.ID)
	if err != nil {
		return nil, err
	}

	if len(req.Text) < 3 {
		return nil, fmt.Errorf("minimum text size is 4")
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
func (t *Stream) Delete(ctx context.Context, e gh.StreamEvent) (*gh.StreamResponse, error) {
	req := new(TodoRequest)
	err := json.NewDecoder(bytes.NewReader(e.Params)).Decode(req)
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

	return &gh.StreamResponse{
		Action: "remove",
		Target: e.Target,
	}, nil
}

func (t *Stream) Get(ctx context.Context, params []byte) (interface{}, error) {
	time.Sleep(1 * time.Second)
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
