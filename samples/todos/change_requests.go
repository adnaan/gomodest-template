package todos

import (
	"context"
	"errors"
	"fmt"
	glv "gomodest-template/pkg/goliveview"
	"gomodest-template/samples/todos/gen/models"
	"gomodest-template/samples/todos/gen/models/todo"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi"

	"github.com/fatih/structs"
	"github.com/google/uuid"
)

var (
	errParseParams = errors.New("error parsing params")
	errQueryDB     = errors.New("error fetching data from db")
	errUpdateDB    = errors.New("error updating data")
	offset         = 0
	limit          = 3
)

type ChangeRequestHandlers struct {
	DB *models.Client
}

func (t *ChangeRequestHandlers) Map() map[string]glv.ChangeRequestHandler {
	return map[string]glv.ChangeRequestHandler{
		"list":           t.List,
		"insert":         t.Create,
		"update":         t.Update,
		"delete":         t.Delete,
		"get":            t.Get,
		"validate_input": t.ValidateInput,
	}
}

func (t *ChangeRequestHandlers) todosPageData(ctx context.Context, query Query) (glv.M, error) {

	todosQuery := t.DB.Todo.
		Query().
		Offset(query.Offset).
		Limit(query.Limit)

	if query.Order == "asc" {
		todosQuery = todosQuery.Order(models.Asc(todo.FieldUpdatedAt))
	} else {
		todosQuery = todosQuery.Order(models.Desc(todo.FieldUpdatedAt))
	}

	todos, err := todosQuery.All(ctx)
	if err != nil {
		log.Printf("err: query.all todos %v", err)
		return nil, err
	}

	count := t.DB.Todo.Query().CountX(ctx)
	pageData := glv.M{"todos": todos}

	if count-query.Offset > query.Limit {
		pageData["next"] = query.Offset + query.Limit
	}
	if (query.Offset - query.Limit) >= 0 {
		pageData["prev"] = query.Offset - query.Limit
		pageData["prev_enabled"] = 1
	}

	pageData["limit"] = query.Limit
	pageData["offset"] = query.Offset
	pageData["order"] = query.Order
	pageData["query"] = query

	return pageData, nil
}

func (t *ChangeRequestHandlers) OnListMount(r *http.Request) (int, glv.M) {
	query := Query{
		Offset: offset,
		Limit:  limit,
		Order:  "asc",
	}
	pageData, err := t.todosPageData(r.Context(), query)
	if err != nil {
		return 200, nil
	}

	return 200, pageData
}

func (t *ChangeRequestHandlers) OnEditMount(r *http.Request) (int, glv.M) {
	id := chi.URLParam(r, "id")
	uid, err := uuid.Parse(id)
	if err != nil {
		return 404, nil
	}
	todo, err := t.DB.Todo.Get(r.Context(), uid)
	if err != nil {
		return 500, nil
	}
	return 200, structs.Map(todo)
}

func (t *ChangeRequestHandlers) List(ctx context.Context, r glv.ChangeRequest, s glv.Session) error {
	var query Query
	err := r.DecodeParams(&query)
	if err != nil {
		return fmt.Errorf(
			"err decode changeRequest params: %v, %w",
			r,
			errParseParams)
	}

	pageData, err := t.todosPageData(ctx, query)
	if err != nil {
		return fmt.Errorf("err db %v, %w", err, errQueryDB)
	}

	s.Change(pageData)
	return nil
}

func loading(enable bool) glv.M {
	target := glv.ChangeTarget(glv.Update, "new_todo", "new_todo")
	if enable {
		target["loading"] = 1
	}
	return target
}

func (t *ChangeRequestHandlers) Create(ctx context.Context, r glv.ChangeRequest, s glv.Session) error {
	s.Change(loading(true))
	defer func() { s.Change(loading(false)) }()

	//// fake sleep a bit to show the loading state.
	//time.Sleep(1 * time.Second)

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
	_, err = t.DB.Todo.Create().
		SetStatus(todo.StatusInprogress).
		SetText(req.Text).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("err create todo %v, %w", err, errUpdateDB)
	}

	if req.Redirect {
		s.Change(glv.M{
			"redirect": "/samples/live/multi/todos",
		})
		return nil
	}

	var query Query
	if v, ok := s.Get("query"); ok {
		query = v.(Query)
	}

	pageData, err := t.todosPageData(ctx, query)
	if err != nil {
		return fmt.Errorf("err db %v, %w", err, errQueryDB)
	}

	s.Change(pageData)
	return nil
}

func (t *ChangeRequestHandlers) Update(ctx context.Context, r glv.ChangeRequest, s glv.Session) error {
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
	s.Flash(2*time.Second, glv.M{
		"message": "saved",
	})
	s.Change(structs.Map(todo))
	return nil
}

func (t *ChangeRequestHandlers) Delete(ctx context.Context, r glv.ChangeRequest, s glv.Session) error {
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

	if req.Redirect {
		s.Change(glv.M{
			"redirect": "/samples/live/multi/todos",
		})
		return nil
	}

	var query Query
	if v, ok := s.Get("query"); ok {
		query = v.(Query)
	}

	pageData, err := t.todosPageData(ctx, query)
	if err != nil {
		return fmt.Errorf("err db %v, %w", err, errQueryDB)
	}

	s.Change(pageData)
	return nil
}

func (t *ChangeRequestHandlers) ValidateInput(ctx context.Context, r glv.ChangeRequest, s glv.Session) error {
	req := new(TodoRequest)
	err := r.DecodeParams(req)
	if err != nil {
		return fmt.Errorf("err decode params: %v, %w", err, errParseParams)
	}

	if len(req.Text) < 3 {
		s.Change(glv.M{
			"new_todo_error": "minimum text length is 3",
		})
	} else {
		s.Change(nil)
	}

	return nil
}

func (t *ChangeRequestHandlers) Get(ctx context.Context, r glv.ChangeRequest, s glv.Session) error {
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
