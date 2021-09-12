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

type App struct {
	DB          *models.Client
	FormDecoder *form.Decoder
}

func (a *App) List() rl.Data {
	return func(w http.ResponseWriter, r *http.Request) (rl.D, error) {
		todos, err := a.DB.Todo.Query().All(r.Context())
		if err != nil {
			return nil, fmt.Errorf("%w", err)
		}

		return rl.D{
			"todos": todos,
		}, nil
	}
}

func (a *App) Create() rl.Data {
	type req struct {
		Text string `json:"text"`
	}

	return func(w http.ResponseWriter, r *http.Request) (rl.D, error) {
		req := new(req)
		err := r.ParseForm()
		if err != nil {
			return nil, fmt.Errorf("%w", err)
		}

		err = a.FormDecoder.Decode(req, r.Form)
		if err != nil {
			return nil, fmt.Errorf("%w", err)
		}

		if req.Text == "" {
			return nil, fmt.Errorf("%w", fmt.Errorf("empty task"))
		}

		_, err = a.DB.Todo.Create().
			SetStatus(todo.StatusInprogress).
			SetText(req.Text).
			Save(r.Context())
		if err != nil {
			return nil, fmt.Errorf("%w", err)
		}

		return nil, nil
	}
}

func (a *App) CreateMulti() rl.Data {
	type req struct {
		Text string `json:"text"`
	}

	return func(w http.ResponseWriter, r *http.Request) (rl.D, error) {
		req := new(req)
		err := r.ParseForm()
		if err != nil {
			return nil, fmt.Errorf("%w", err)
		}

		err = a.FormDecoder.Decode(req, r.Form)
		if err != nil {
			return nil, fmt.Errorf("%w", err)
		}

		if req.Text == "" {
			return nil, fmt.Errorf("%w", fmt.Errorf("empty task"))
		}

		t, err := a.DB.Todo.Create().
			SetStatus(todo.StatusInprogress).
			SetText(req.Text).
			Save(r.Context())
		if err != nil {
			return nil, fmt.Errorf("%w", err)
		}

		http.Redirect(w, r, "/samples/todos_multi/"+t.ID.String(), http.StatusSeeOther)

		return nil, nil
	}
}

func (a *App) View() rl.Data {
	return func(w http.ResponseWriter, r *http.Request) (rl.D, error) {
		id := chi.URLParam(r, "id")
		uid, err := uuid.Parse(id)
		if err != nil {
			return nil, fmt.Errorf("%w", err)
		}
		t, err := a.DB.Todo.Get(r.Context(), uid)
		if err != nil {
			return nil, fmt.Errorf("%w", err)
		}
		return rl.D{
				"todo":          t,
				"action_delete": fmt.Sprintf("/samples/todos_multi/%s/delete", t.ID.String()),
				"action_edit":   fmt.Sprintf("/samples/todos_multi/%s", t.ID.String()),
			},
			nil
	}
}

func (a *App) Edit() rl.Data {
	type req struct {
		Text string `json:"text"`
	}
	return func(w http.ResponseWriter, r *http.Request) (rl.D, error) {
		req := new(req)
		err := r.ParseForm()
		if err != nil {
			return nil, fmt.Errorf("%w", err)
		}

		err = a.FormDecoder.Decode(req, r.Form)
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

		err = a.DB.Todo.UpdateOneID(uid).SetText(req.Text).Exec(r.Context())
		if err != nil {
			return nil, fmt.Errorf("%w", err)
		}

		return nil, nil
	}
}

func (a *App) Delete() rl.Data {
	return func(w http.ResponseWriter, r *http.Request) (rl.D, error) {
		id := chi.URLParam(r, "id")
		uid, err := uuid.Parse(id)
		if err != nil {
			return nil, fmt.Errorf("%w", err)
		}
		err = a.DB.Todo.DeleteOneID(uid).Exec(r.Context())
		if err != nil {
			return nil, fmt.Errorf("%w", err)
		}

		return nil, nil
	}
}

func (a *App) DeleteMulti() rl.Data {
	return func(w http.ResponseWriter, r *http.Request) (rl.D, error) {
		id := chi.URLParam(r, "id")
		uid, err := uuid.Parse(id)
		if err != nil {
			return nil, fmt.Errorf("%w", err)
		}
		err = a.DB.Todo.DeleteOneID(uid).Exec(r.Context())
		if err != nil {
			return nil, fmt.Errorf("%w", err)
		}

		http.Redirect(w, r, "/samples/todos_multi", http.StatusSeeOther)
		return nil, nil
	}
}
