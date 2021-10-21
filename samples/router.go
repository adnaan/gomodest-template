package samples

import (
	"context"
	"encoding/json"
	"fmt"
	glv "gomodest-template/pkg/goliveview"
	"gomodest-template/pkg/websocketjsonrpc2"
	"gomodest-template/samples/todos"
	"gomodest-template/samples/todos/gen/models"
	"log"
	"net/http"
	"os"
	"strings"

	rl "github.com/adnaan/renderlayout"
	"github.com/go-chi/chi"
	"github.com/go-playground/form"
	"github.com/gorilla/sessions"
	"github.com/vulcand/oxy/forward"
	"github.com/vulcand/oxy/testutils"
)

var store = sessions.NewCookieStore([]byte(os.Getenv("my-secret-key")))

type Result struct {
	Method string      `json:"method"`
	Data   interface{} `json:"data"`
}

func sessionMw(store sessions.Store) func(http.Handler) http.Handler {
	f := func(h http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			session, _ := store.Get(r, "_session_id")
			// Set some session values.
			key := "helloworld123"
			session.Values["key"] = key
			// Save it before we write to the response/return from the handler.
			err := session.Save(r, w)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			h.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
	return f
}
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

	data := struct {
		Title string `json:"title"`
	}{
		Title: "Hello from server for the svelte component",
	}

	d, err := json.Marshal(&data)
	if err != nil {
		panic(err)
	}
	appData := string(d)

	todos.StartRPCServer(db, ctx)
	return func(r chi.Router) {
		r.Get("/", index("samples/list"))
		r.Get("/sidemenu", index("samples/sidemenu"))

		r.Get("/svelte", index("samples/svelte", rl.StaticData(rl.D{"Data": appData})))
		r.Get("/svelte_todos", index("samples/svelte_todos", rl.StaticData(rl.D{"Data": appData})))

		r.Get("/svelte_ws_todos", index("samples/svelte_ws_todos", rl.StaticData(rl.D{"Data": appData})))
		r.Get("/svelte_ws2_todos", index("samples/svelte_ws2_todos", rl.StaticData(rl.D{"Data": appData})))
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
		r.Get("/svelte_ws2_todos_multi/new", index("samples/svelte_todos_multi/new"))
		fwd, _ := forward.New()
		r.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
			r.URL = testutils.ParseURI("http://localhost:3001/")
			fwd.ServeHTTP(w, r)
		})

		r.Route("/api/todos", todosAPIRouter(db))

		// todos turbo-frame sample
		r.Route("/todos", turboFrameSPARouter(index, app))
		// todos multi sample turbo-frame
		r.Route("/todos_multi", turboFrameMPARouter(index, app))

		r.Route("/ws/todos", todosJsonRpc2WebsocketRouter(db))
		r.Route("/live", todosLiveRouter(db))
		r.Route("/live/multi", todosLiveMultiRouter(db))

	}
}

func todosLiveRouter(db *models.Client) func(r chi.Router) {
	return func(r chi.Router) {
		todosEventHandler := todos.ChangeRequestHandlers{DB: db}
		name := "gomodest-template"
		glvc := glv.WebsocketController(&name, glv.EnableHTMLFormatting())
		todosView := glvc.NewView(
			"./templates/samples/todos_live",
			glv.WithOnMount(todosEventHandler.OnListMount),
			glv.WithChangeRequestHandlers(todosEventHandler.Map()))

		r.Handle("/todos", todosView)
	}
}

func todosLiveMultiRouter(db *models.Client) func(r chi.Router) {
	return func(r chi.Router) {
		todosEventHandler := todos.ChangeRequestHandlers{DB: db}
		name := "gomodest-template-multi"
		glvc := glv.WebsocketController(&name, glv.EnableHTMLFormatting())
		partials := glv.WithPartials("./templates/samples/todos_live_multi/partials", "./templates/partials")
		todosView := glvc.NewView(
			"./templates/samples/todos_live_multi/index.html",
			partials,
			glv.WithOnMount(todosEventHandler.OnListMount),
			glv.WithChangeRequestHandlers(todosEventHandler.Map()))

		newTodoView := glvc.NewView(
			"./templates/samples/todos_live_multi/new.html",
			partials,
			glv.WithChangeRequestHandlers(todosEventHandler.Map()))

		editTodoView := glvc.NewView(
			"./templates/samples/todos_live_multi/edit.html",
			partials,
			glv.WithOnMount(todosEventHandler.OnEditMount),
			glv.WithChangeRequestHandlers(todosEventHandler.Map()))

		r.Handle("/todos", todosView)
		r.Handle("/todos/new", newTodoView)
		r.Handle("/todos/{id}/edit", editTodoView)
	}
}

func todosJsonRpc2WebsocketRouter(db *models.Client) func(r chi.Router) {
	return func(r chi.Router) {
		todosJsonRpc2 := todos.TodosJsonRpc2{DB: db}
		methods := map[string]websocketjsonrpc2.Method{
			"todos/list":   todosJsonRpc2.List,
			"todos/insert": todosJsonRpc2.Create,
			"todos/delete": todosJsonRpc2.Delete,
			"todos/update": todosJsonRpc2.Update,
			"todos/get":    todosJsonRpc2.Get,
		}

		options := []websocketjsonrpc2.Option{
			websocketjsonrpc2.WithRequestContext(
				func(r *http.Request) context.Context {
					return context.WithValue(r.Context(), "user_id", "xyz1234")
				}),
			websocketjsonrpc2.WithSubscribeTopic(func(r *http.Request) *string {
				session, _ := store.Get(r, "_session_id")
				v, ok := session.Values["key"]
				if !ok {
					return nil
				}
				key := v.(string)

				topic := fmt.Sprintf("%s_%s",
					strings.Replace(r.URL.Path, "/", "_", -1), key)
				log.Println("subscribed to topic", topic)
				return &topic
			}),
			//websocketjsonrpc2.WithResultHook(
			//	func(method string, result interface{}) interface{} {
			//		return &Result{
			//			Method: method,
			//			Data:   result,
			//		}
			//	}),
		}

		websocketjsonrpc2Router := websocketjsonrpc2.NewRouter()
		r.Route("/", func(r chi.Router) {
			r.Use(sessionMw(store))
			r.HandleFunc("/{id}",
				websocketjsonrpc2Router.HandlerFunc(
					methods,
					options...,
				),
			)
			//options = append(options, websocketjsonrpc2.WithOnConnectMethod("todos/list"))
			r.HandleFunc("/",
				websocketjsonrpc2Router.HandlerFunc(
					methods,
					options...,
				),
			)
		})
	}
}

func turboFrameSPARouter(index rl.Render, app todos.App) func(r chi.Router) {
	return func(r chi.Router) {
		r.Get("/", index("samples/todos/main"))
		// single turbo list which is replaced over and over.
		r.Get("/list", index("samples/todos/list", app.List()))
		r.Post("/new", index("samples/todos/list", app.Create(), app.List()))
		r.Post("/{id}/edit", index("samples/todos/list", app.Edit(), app.List()))
		r.Post("/{id}/delete", index("samples/todos/list", app.Delete(), app.List()))
	}
}

func turboFrameMPARouter(index rl.Render, app todos.App) func(r chi.Router) {
	return func(r chi.Router) {
		todosMulti := pagePath("samples/todos_multi")
		r.Get("/", index(todosMulti("index")))
		r.Get("/list", index("samples/todos_multi/list", app.List()))
		// new
		r.Get("/new", index("samples/todos_multi/new"))
		r.Post("/new", index("samples/todos_multi/new", app.CreateMulti()))
		// edit
		r.Get("/{id}", index("samples/todos_multi/view", app.View()))
		r.Post("/{id}", index("samples/todos_multi/view", app.Edit(), app.View()))
		r.Post("/{id}/delete", index("samples/todos_multi/view", app.DeleteMulti()))
	}
}

func todosAPIRouter(db *models.Client) func(r chi.Router) {
	return func(r chi.Router) {
		r.Get("/", todos.List(db))
		r.Post("/", todos.Create(db))
		r.Route("/{id}", func(r chi.Router) {
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
