package todos

import (
	"bytes"
	"context"
	"encoding/json"
	gh "gomodest-template/pkg/gohotwired"
	"gomodest-template/samples/todos/gen/models"
	"gomodest-template/samples/todos/gen/models/todo"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type EventHandler struct {
	DB *models.Client
}

func (t *EventHandler) OnMount(r *http.Request) (int, gh.M) {
	todos, err := t.DB.Todo.Query().All(r.Context())
	if err != nil {
		log.Printf("err: query.all todos %v", err)
		return 200, nil
	}
	return 200, gh.M{
		"todos": todos,
	}
}

func (t *EventHandler) Map() map[string]gh.EventHandler {
	return map[string]gh.EventHandler{
		"todos/list":   t.List,
		"todos/insert": t.Create,
		"todos/update": t.Update,
		"todos/delete": t.Delete,
	}
}

func (t *EventHandler) List(ctx context.Context, s gh.Stream) {
	query := &Query{
		Offset: 0,
		Limit:  3,
	}
	err := json.NewDecoder(bytes.NewReader(s.Event().Params)).Decode(query)
	if err != nil {
		s.Error("error parsing params", err)
		return
	}
	todos, err := t.DB.Todo.
		Query().
		Offset(query.Offset).
		Limit(query.Limit).
		Order(models.Desc(todo.FieldUpdatedAt)).
		All(ctx)
	if err != nil {
		s.Error("error fetching todos from db", err)
		return
	}

	s.Update(s.Event().Target, "todos", gh.M{"todos": todos})
}

func (t *EventHandler) Create(ctx context.Context, s gh.Stream) {
	s.Update("new_todo", "new_todo", gh.M{"loading": 1})
	defer func() {
		s.Update("new_todo", "new_todo", nil)
	}()
	time.Sleep(1 * time.Second)
	req := new(TodoRequest)
	err := json.NewDecoder(bytes.NewReader(s.Event().Params)).Decode(req)
	if err != nil {
		s.Error("error parsing params", err)
		return
	}
	if len(req.Text) < 3 {
		s.Error("minimum text-size is 3")
		return
	}
	todo, err := t.DB.Todo.Create().
		SetStatus(todo.StatusInprogress).
		SetText(req.Text).
		Save(ctx)
	if err != nil {
		s.Error("error saving todo", err)
		return
	}

	s.Append(s.Event().Target, "todo", gh.M{"ID": todo.ID, "Text": todo.Text})

}
func (t *EventHandler) Update(ctx context.Context, s gh.Stream) {
	req := new(TodoRequest)
	err := json.NewDecoder(bytes.NewReader(s.Event().Params)).Decode(req)
	if err != nil {
		s.Error("error parsing params", err)
		return
	}

	uid, err := uuid.Parse(req.ID)
	if err != nil {
		s.Error("error parsing todo id", err)
		return
	}

	if len(req.Text) < 3 {
		s.Error("minimum text-size is 3")
		return
	}

	todo, err := t.DB.Todo.
		UpdateOneID(uid).
		SetUpdatedAt(time.Now()).
		SetText(req.Text).
		Save(ctx)
	if err != nil {
		s.Error("error updating todo", err)
		return
	}
	s.Update(s.Event().Target, "todo", gh.M{"ID": todo.ID, "Text": todo.Text})
}
func (t *EventHandler) Delete(ctx context.Context, s gh.Stream) {
	req := new(TodoRequest)
	err := json.NewDecoder(bytes.NewReader(s.Event().Params)).Decode(req)
	if err != nil {
		s.Error("error parsing params", err)
		return
	}

	uid, err := uuid.Parse(req.ID)
	if err != nil {
		s.Error("error parsing todo id", err)
		return
	}
	err = t.DB.Todo.DeleteOneID(uid).Exec(ctx)
	if err != nil {
		s.Error("error deleting todo", err)
		return
	}

	s.Remove(s.Event().Target)
}

func (t *EventHandler) Get(ctx context.Context, params []byte) (interface{}, error) {
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
