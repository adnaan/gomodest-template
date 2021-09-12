package todos

import (
	"gomodest-template/samples/todos/gen/models"
	"gomodest-template/samples/todos/gen/models/todo"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/go-chi/chi"

	"github.com/go-chi/render"
)

func List(c *models.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tasks, err := c.Todo.Query().All(r.Context())
		if err != nil {
			render.Render(w, r, ErrInternal(err))
			return
		}

		render.JSON(w, r, tasks)
	}
}

func Create(c *models.Client) http.HandlerFunc {
	type req struct {
		Text string `json:"text"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		req := new(req)
		err := render.DecodeJSON(r.Body, &req)
		if err != nil {
			render.Render(w, r, ErrInternal(err))
			return
		}

		newTask, err := c.Todo.Create().
			SetStatus(todo.StatusInprogress).
			SetText(req.Text).
			Save(r.Context())
		if err != nil {
			render.Render(w, r, ErrInternal(err))
			return
		}
		render.JSON(w, r, newTask)
	}
}

func UpdateStatus(c *models.Client) http.HandlerFunc {
	type req struct {
		Status string `json:"status"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		req := new(req)
		id := chi.URLParam(r, "id")

		err := render.DecodeJSON(r.Body, &req)
		if err != nil {
			render.Render(w, r, ErrInternal(err))
			return
		}

		uid, err := uuid.Parse(id)
		if err != nil {
			render.Render(w, r, ErrInternal(err))
			return
		}

		updatedTask, err := c.Todo.
			UpdateOneID(uid).
			SetUpdatedAt(time.Now()).
			SetStatus(todo.Status(req.Status)).
			Save(r.Context())
		if err != nil {
			render.Render(w, r, ErrInternal(err))
			return
		}
		render.JSON(w, r, updatedTask)
	}
}

func UpdateText(c *models.Client) http.HandlerFunc {
	type req struct {
		Text string `json:"text"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		req := new(req)
		id := chi.URLParam(r, "id")

		err := render.DecodeJSON(r.Body, &req)
		if err != nil {
			render.Render(w, r, ErrInternal(err))
			return
		}

		uid, err := uuid.Parse(id)
		if err != nil {
			render.Render(w, r, ErrInternal(err))
			return
		}

		updatedTask, err := c.Todo.
			UpdateOneID(uid).
			SetUpdatedAt(time.Now()).
			SetText(req.Text).
			Save(r.Context())
		if err != nil {
			render.Render(w, r, ErrInternal(err))
			return
		}
		render.JSON(w, r, updatedTask)
	}
}

func Delete(c *models.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		uid, err := uuid.Parse(id)
		if err != nil {
			render.Render(w, r, ErrInternal(err))
			return
		}
		err = c.Todo.DeleteOneID(uid).Exec(r.Context())
		if err != nil {
			render.Render(w, r, ErrInternal(err))
			return
		}
		render.Status(r, http.StatusOK)
		render.JSON(w, r, struct {
			Success bool `json:"success"`
		}{
			Success: true,
		})
	}
}
