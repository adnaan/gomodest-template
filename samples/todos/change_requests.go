package todos

import (
	"context"
	"errors"
	"fmt"
	gw "gomodest-template/pkg/goliveview"
	"gomodest-template/samples/todos/gen/models"
	"gomodest-template/samples/todos/gen/models/todo"
	"log"
	"net/http"
	"time"

	"github.com/fatih/structs"
	"github.com/google/uuid"
)

type ChangeRequestHandlers struct {
	DB *models.Client
}

type M map[string]interface{}

var (
	errParseParams = errors.New("error parsing params")
	errQueryDB     = errors.New("error fetching data from db")
	errUpdateDB    = errors.New("error updating data")
	offset         = 0
	limit          = 3
)

func (t *ChangeRequestHandlers) todosPageData(ctx context.Context, query Query) (map[string]interface{}, error) {
	todos, err := t.DB.Todo.
		Query().
		Offset(query.Offset).
		Limit(query.Limit).
		Order(models.Desc(todo.FieldUpdatedAt)).
		All(ctx)
	if err != nil {
		log.Printf("err: query.all todos %v", err)
		return nil, err
	}

	pageData := M{"todos": todos}
	if len(todos) > 0 {
		pageData["next"] = query.Offset + query.Limit
	}
	if query.Offset > query.Limit {
		pageData["prev"] = query.Offset - query.Limit
	}

	return pageData, nil
}

func (t *ChangeRequestHandlers) OnMount(r *http.Request) (int, map[string]interface{}) {
	query := Query{
		Offset: offset,
		Limit:  limit,
	}
	pageData, err := t.todosPageData(r.Context(), query)
	if err != nil {
		return 200, nil
	}
	return 200, pageData
}

func (t *ChangeRequestHandlers) Map() map[string]gw.ChangeRequestHandler {
	return map[string]gw.ChangeRequestHandler{
		"todos/list":   t.List,
		"todos/insert": t.Create,
		"todos/update": t.Update,
		"todos/delete": t.Delete,
		"todos/get":    t.Get,
	}
}

func (t *ChangeRequestHandlers) List(ctx context.Context, r gw.ChangeRequest, s gw.Session) error {
	query := &Query{
		Offset: offset,
		Limit:  limit,
	}
	err := r.DecodeParams(query)
	if err != nil {
		return fmt.Errorf(
			"err decode changeRequest params: %v, %w",
			r,
			errParseParams)
	}

	pageData, err := t.todosPageData(ctx, *query)
	if err != nil {
		return fmt.Errorf("err db %v, %w", err, errQueryDB)
	}

	s.Change(pageData)
	return nil
}

func loading(enable bool) map[string]interface{} {
	target := gw.ChangeTarget(gw.Update, "new_todo", "new_todo")
	if enable {
		target["loading"] = 1
	}
	return target
}

func (t *ChangeRequestHandlers) Create(ctx context.Context, r gw.ChangeRequest, s gw.Session) error {
	s.Change(loading(true))
	defer func() { s.Change(loading(false)) }()

	// fake sleep a bit to show the loading state.
	time.Sleep(1 * time.Second)

	// decode incoming params
	req := new(TodoRequest)
	err := r.DecodeParams(req)
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

	next, _ := s.Get("next")
	log.Println("next", next)

	s.Change(structs.Map(todo))
	s.Flash(3*time.Second, map[string]interface{}{
		"message": "created todo",
	})
	return nil
}

func (t *ChangeRequestHandlers) Update(ctx context.Context, r gw.ChangeRequest, s gw.Session) error {
	req := new(TodoRequest)
	err := r.DecodeParams(req)
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

	s.Change(structs.Map(todo))
	return nil
}
func (t *ChangeRequestHandlers) Delete(ctx context.Context, r gw.ChangeRequest, s gw.Session) error {
	req := new(TodoRequest)
	err := r.DecodeParams(req)
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

	s.Change(nil)
	return nil
}

func (t *ChangeRequestHandlers) Get(ctx context.Context, r gw.ChangeRequest, s gw.Session) error {
	req := new(TodoRequest)
	err := r.DecodeParams(req)
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

	s.Change(structs.Map(todo))
	return nil
}
