package todos

import (
	"context"
	"errors"
	"fmt"
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

type M map[string]interface{}

var (
	errParseParams = errors.New("error parsing params")
	errQueryDB     = errors.New("error fetching data from db")
	errUpdateDB    = errors.New("error updating data")
)

func (t *EventHandler) OnMount(r *http.Request) (int, interface{}) {
	todos, err := t.DB.Todo.Query().Order(models.Desc(todo.FieldUpdatedAt)).All(r.Context())
	if err != nil {
		log.Printf("err: query.all todos %v", err)
		return 200, nil
	}
	return 200, M{
		"todos": todos,
	}
}

func (t *EventHandler) Map() map[string]gh.EventHandler {
	return map[string]gh.EventHandler{
		"todos/list":   t.List,
		"todos/insert": t.Create,
		"todos/update": t.Update,
		"todos/delete": t.Delete,
		"todos/get":    t.Get,
	}
}

func (t *EventHandler) List(ctx context.Context, s gh.Stream) error {
	query := &Query{
		Offset: 0,
		Limit:  3,
	}
	err := s.DecodeParams(query)
	if err != nil {
		return fmt.Errorf(
			"err decode params: %v, %w",
			s.Event().Params,
			errParseParams)
	}

	todos, err := t.DB.Todo.
		Query().
		Offset(query.Offset).
		Limit(query.Limit).
		Order(models.Desc(todo.FieldUpdatedAt)).
		All(ctx)
	if err != nil {
		return fmt.Errorf("err db %v, %w", err, errQueryDB)
	}

	s.Echo(M{"todos": todos})
	return nil
}

func loadingCreateTodo(enable bool) gh.Event {
	e := gh.Event{
		Action:  gh.Update,
		Target:  "new_todo",
		Content: "new_todo",
	}
	if enable {
		e.Data = M{"loading": 1}
	}
	return e
}

func (t *EventHandler) Create(ctx context.Context, s gh.Stream) error {
	// reply a turbo-stream partial by sending the event
	// set loading
	s.Send(loadingCreateTodo(true))
	defer func() {
		// unset loading
		s.Send(loadingCreateTodo(false))
	}()

	// fake sleep a bit to show the loading state.
	time.Sleep(1 * time.Second)

	// decode incoming params
	req := new(TodoRequest)
	err := s.DecodeParams(req)
	if err != nil {
		return fmt.Errorf("err decode params: %v, %w", err, errParseParams)
	}

	// validate
	if len(req.Text) < 3 {
		// wrap the error you want to show the user in the UI with %w
		return fmt.Errorf("err %w", errors.New("minimum text size is 3"))
	}

	// create todo
	todo, err := t.DB.Todo.Create().
		SetStatus(todo.StatusInprogress).
		SetText(req.Text).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("err create todo %v, %w", err, errUpdateDB)
	}

	// reply with the received action, target, content + new data(todo)
	s.Echo(todo)
	return nil
}

func (t *EventHandler) Update(ctx context.Context, s gh.Stream) error {
	req := new(TodoRequest)
	err := s.DecodeParams(req)
	if err != nil {
		return fmt.Errorf("err decode params: %v, %w", err, errParseParams)
	}

	uid, err := uuid.Parse(req.ID)
	if err != nil {
		return fmt.Errorf("err %v, %w", err, errors.New("invalid todo id"))
	}

	if len(req.Text) < 3 {
		return fmt.Errorf("err %w", errors.New("minimum text size is 3"))
	}

	todo, err := t.DB.Todo.
		UpdateOneID(uid).
		SetUpdatedAt(time.Now()).
		SetText(req.Text).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("err update todo %v, %w", err, errUpdateDB)
	}

	s.Echo(todo)
	return nil
}
func (t *EventHandler) Delete(ctx context.Context, s gh.Stream) error {
	req := new(TodoRequest)
	err := s.DecodeParams(req)
	if err != nil {
		return fmt.Errorf("err decode params: %v, %w", err, errParseParams)
	}

	uid, err := uuid.Parse(req.ID)
	if err != nil {
		return fmt.Errorf("err %v, %w", err, errors.New("invalid todo id"))
	}

	err = t.DB.Todo.DeleteOneID(uid).Exec(ctx)
	if err != nil {
		return fmt.Errorf("err %v, %w", err, errors.New("error deleting todo"))
	}

	s.Echo(nil)
	return nil
}

func (t *EventHandler) Get(ctx context.Context, s gh.Stream) error {
	req := new(TodoRequest)
	err := s.DecodeParams(req)
	if err != nil {
		return fmt.Errorf("err decode params: %v, %w", err, errParseParams)
	}

	uid, err := uuid.Parse(req.ID)
	if err != nil {
		return fmt.Errorf("err %v, %w", err, errors.New("invalid todo id"))
	}
	todo, err := t.DB.Todo.Get(ctx, uid)
	if err != nil {
		return err
	}

	s.Echo(todo)
	return nil
}
