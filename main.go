package main

import (
	"context"
	"encoding/json"
	"fmt"
	"gomodest-template/samples/todos"
	"gomodest-template/samples/todos/gen/models"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-playground/form"

	rl "github.com/adnaan/renderlayout"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
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

	index, err := rl.New(
		rl.Layout("index"),
		rl.DisableCache(true),
		rl.Debug(true),
		rl.DefaultData(func(w http.ResponseWriter, r *http.Request) (rl.D, error) {
			return rl.D{
				"route":    r.URL.Path,
				"app_name": "gomodest-template",
			}, nil
		}))

	if err != nil {
		log.Fatal(err)
	}

	r := chi.NewRouter()
	r.Use(middleware.Compress(5))
	r.Use(middleware.StripSlashes)
	r.NotFound(index("404"))
	r.Get("/", index("home", rl.StaticData(rl.D{"hello": "world"})))
	r.Route("/samples", func(r chi.Router) {
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
	})

	workDir, _ := os.Getwd()
	public := http.Dir(filepath.Join(workDir, "./", "public", "assets"))
	staticHandler(r, "/static", public)

	fmt.Println("listening on http://localhost:3000")
	err = http.ListenAndServe(":3000", r)
	if err != nil {
		log.Fatal(err)
	}
}

func pagePath(base string) func(page string) string {
	return func(page string) string {
		base = strings.TrimLeft(base, "/")
		return strings.Join([]string{base, page}, "/")
	}
}

func staticHandler(r chi.Router, path string, root http.FileSystem) {
	if strings.ContainsAny(path, "{}*") {
		panic("FileServer does not permit any URL parameters.")
	}

	if path != "/" && path[len(path)-1] != '/' {
		r.Get(path, http.RedirectHandler(path+"/", 301).ServeHTTP)
		path += "/"
	}
	path += "*"

	r.Get(path, func(w http.ResponseWriter, r *http.Request) {
		rctx := chi.RouteContext(r.Context())
		pathPrefix := strings.TrimSuffix(rctx.RoutePattern(), "/*")
		fs := http.StripPrefix(pathPrefix, http.FileServer(root))
		fs.ServeHTTP(w, r)
	})
}
