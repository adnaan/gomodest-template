package samples

import (
	"context"
	"encoding/json"
	"fmt"
	"gomodest-template/pkg/websocketjsonrpc2"
	"gomodest-template/samples/todos"
	"gomodest-template/samples/todos/gen/models"
	"net/http"
	"strings"

	"github.com/vulcand/oxy/testutils"

	"github.com/vulcand/oxy/forward"

	"github.com/go-playground/form"

	rl "github.com/adnaan/renderlayout"
	"github.com/go-chi/chi"
)

func Router(index rl.Render) func(r chi.Router) {
	ctx := context.Background()
	db, err := models.Open("sqlite3", "file:app.db?mode=memory&cache=shared&_fk=1")
	if err != nil {
		panic(err)
	}
	if err := db.Schema.Create(ctx); err != nil {
		panic(err)
	}

	app := todos.App{
		DB:          db,
		FormDecoder: form.NewDecoder(),
	}
	todos.StartRPCServer(db, ctx)
	return func(r chi.Router) {
		r.Get("/", index("samples/list"))
		r.Get("/sidemenu", index("samples/sidemenu"))

		r.Get("/svelte", index("samples/svelte",
			func(w http.ResponseWriter, r *http.Request) (rl.D, error) {
				appData := struct {
					Title string `json:"title"`
				}{
					Title: "Hello from server for the svelte component",
				}

				d, err := json.Marshal(&appData)
				if err != nil {
					return nil, fmt.Errorf("%v: %w", err, fmt.Errorf("encoding failed"))
				}

				return rl.D{
					"Data": string(d), // notice struct is converted into a string
				}, nil
			}))
		r.Get("/svelte_todos", index("samples/svelte_todos",
			func(w http.ResponseWriter, r *http.Request) (rl.D, error) {
				appData := struct {
					Title string `json:"title"`
				}{
					Title: "Hello from server for the svelte todos component",
				}

				d, err := json.Marshal(&appData)
				if err != nil {
					return nil, fmt.Errorf("%v: %w", err, fmt.Errorf("encoding failed"))
				}

				return rl.D{
					"Data": string(d), // notice struct is converted into a string
				}, nil
			}))

		r.Get("/svelte_ws_todos", index("samples/svelte_ws_todos",
			func(w http.ResponseWriter, r *http.Request) (rl.D, error) {
				appData := struct {
					Title string `json:"title"`
				}{
					Title: "Hello from server for the svelte todos component",
				}

				d, err := json.Marshal(&appData)
				if err != nil {
					return nil, fmt.Errorf("%v: %w", err, fmt.Errorf("encoding failed"))
				}

				return rl.D{
					"Data": string(d), // notice struct is converted into a string
				}, nil
			}))
		r.Get("/svelte_ws2_todos", index("samples/svelte_ws2_todos",
			func(w http.ResponseWriter, r *http.Request) (rl.D, error) {
				appData := struct {
					Title string `json:"title"`
				}{
					Title: "Hello from server for the svelte todos component",
				}

				d, err := json.Marshal(&appData)
				if err != nil {
					return nil, fmt.Errorf("%v: %w", err, fmt.Errorf("encoding failed"))
				}

				return rl.D{
					"Data": string(d), // notice struct is converted into a string
				}, nil
			}))
		r.Get("/svelte_ws2_todos_multi", index("samples/svelte_todos_multi/list"))
		r.Get("/svelte_ws2_todos_multi/{id}", index("samples/svelte_todos_multi/view",
			func(w http.ResponseWriter, r *http.Request) (rl.D, error) {
				appData := struct {
					ID string `json:"id"`
				}{
					ID: chi.URLParam(r, "id"),
				}

				d, err := json.Marshal(&appData)
				if err != nil {
					return nil, fmt.Errorf("%v: %w", err, fmt.Errorf("encoding failed"))
				}

				return rl.D{
					"Data": string(d), // notice struct is converted into a string
				}, nil
			}))
		fwd, _ := forward.New()
		r.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
			r.URL = testutils.ParseURI("http://localhost:3001/")
			fwd.ServeHTTP(w, r)
		})

		todosJsonRpc2 := todos.TodosJsonRpc2{DB: db}
		methods := map[string]websocketjsonrpc2.Method{
			"todos/list":   todosJsonRpc2.List,
			"todos/create": todosJsonRpc2.Create,
			"todos/delete": todosJsonRpc2.Delete,
			"todos/update": todosJsonRpc2.Update,
			"todos/get":    todosJsonRpc2.Get,
		}

		r.HandleFunc("/ws2",
			websocketjsonrpc2.HandlerFunc(
				methods,
				websocketjsonrpc2.WithRequestContext(
					func(r *http.Request) context.Context {
						return context.WithValue(r.Context(), "key", "value")
					})),
		)
		// todos sample
		r.Get("/todos", index("samples/todos/main"))
		// single turbo list which is replaced over and over.
		r.Get("/todos/list", index("samples/todos/list", app.List()))
		r.Post("/todos/new", index("samples/todos/list", app.Create(), app.List()))
		r.Post("/todos/{id}/edit", index("samples/todos/list", app.Edit(), app.List()))
		r.Post("/todos/{id}/delete", index("samples/todos/list", app.Delete(), app.List()))

		// todos multi sample
		todosMulti := pagePath("samples/todos_multi")
		// home
		r.Get("/todos_multi", index(todosMulti("index")))
		r.Get("/todos_multi/list", index("samples/todos_multi/list", app.List()))
		// new
		r.Get("/todos_multi/new", index("samples/todos_multi/new"))
		r.Post("/todos_multi/new", index("samples/todos_multi/new", app.CreateMulti()))
		// edit
		r.Get("/todos_multi/{id}", index("samples/todos_multi/view", app.View()))
		r.Post("/todos_multi/{id}", index("samples/todos_multi/view", app.Edit(), app.View()))
		r.Post("/todos_multi/{id}/delete", index("samples/todos_multi/view", app.DeleteMulti()))

		r.Route("/api/todos", func(r chi.Router) {
			r.Get("/", todos.List(db))
			r.Post("/", todos.Create(db))
		})
		r.Route("/api/todos/{id}", func(r chi.Router) {
			r.Put("/status", todos.UpdateStatus(db))
			r.Put("/text", todos.UpdateText(db))
			r.Delete("/", todos.Delete(db))
		})
	}
}

func pagePath(base string) func(page string) string {
	return func(page string) string {
		base = strings.TrimLeft(base, "/")
		return strings.Join([]string{base, page}, "/")
	}
}
