package todos

import (
	"fmt"
	"gomodest-template/samples/todos/gen/models"
	"gomodest-template/samples/todos/gen/models/todo"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/google/uuid"

	rl "github.com/adnaan/renderlayout"
	"github.com/go-playground/form"
)

type Deps struct {
	DB          *models.Client
	FormDecoder *form.Decoder
}

func List(dp Deps) rl.Data {
	return func(w http.ResponseWriter, r *http.Request) (rl.D, error) {
		todos, err := dp.DB.Todo.Query().All(r.Context())
		if err != nil {
			return nil, fmt.Errorf("%w", err)
		}

		return rl.D{
			"todos": todos,
		}, nil
	}
}

func Create(dp Deps) rl.Data {
	type req struct {
		Text string `json:"text"`
	}

	return func(w http.ResponseWriter, r *http.Request) (rl.D, error) {
		req := new(req)
		err := r.ParseForm()
		if err != nil {
			return nil, fmt.Errorf("%w", err)
		}

		err = dp.FormDecoder.Decode(req, r.Form)
		if err != nil {
			return nil, fmt.Errorf("%w", err)
		}

		if req.Text == "" {
			return nil, fmt.Errorf("%w", fmt.Errorf("empty task"))
		}

		_, err = dp.DB.Todo.Create().
			SetStatus(todo.StatusInprogress).
			SetText(req.Text).
			Save(r.Context())
		if err != nil {
			return nil, fmt.Errorf("%w", err)
		}

		return nil, nil
	}
}

func Edit(dp Deps) rl.Data {
	type req struct {
		Text string `json:"text"`
	}
	return func(w http.ResponseWriter, r *http.Request) (rl.D, error) {
		req := new(req)
		err := r.ParseForm()
		if err != nil {
			return nil, fmt.Errorf("%w", err)
		}

		err = dp.FormDecoder.Decode(req, r.Form)
		if err != nil {
			return nil, fmt.Errorf("%w", err)
		}

		if req.Text == "" {
			return nil, fmt.Errorf("%w", fmt.Errorf("empty task"))
		}

		id := chi.URLParam(r, "id")
		uid, err := uuid.Parse(id)
		if err != nil {
			return nil, fmt.Errorf("%w", err)
		}

		err = dp.DB.Todo.UpdateOneID(uid).SetText(req.Text).Exec(r.Context())
		if err != nil {
			return nil, fmt.Errorf("%w", err)
		}

		return nil, nil
	}
}

func Delete(dp Deps) rl.Data {
	return func(w http.ResponseWriter, r *http.Request) (rl.D, error) {
		id := chi.URLParam(r, "id")
		uid, err := uuid.Parse(id)
		if err != nil {
			return nil, fmt.Errorf("%w", err)
		}
		err = dp.DB.Todo.DeleteOneID(uid).Exec(r.Context())
		if err != nil {
			return nil, fmt.Errorf("%w", err)
		}

		return nil, nil
	}
}
