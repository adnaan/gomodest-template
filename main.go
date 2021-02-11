package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	rl "github.com/adnaan/renderlayout"
	"github.com/go-chi/chi"
)

func main() {
	indexLayout, err := rl.New(
		rl.Layout("index"),
		rl.DisableCache(true),
		rl.DefaultHandler(func(w http.ResponseWriter, r *http.Request) (rl.M, error) {
			return rl.M{
				"route":    r.URL.Path,
				"app_name": "gomodest-template",
			}, nil
		}))

	if err != nil {
		log.Fatal(err)
	}

	r := chi.NewRouter()
	r.NotFound(indexLayout.Handle("404", rl.StaticView))
	r.Get("/", indexLayout.Handle("home",
		func(w http.ResponseWriter, r *http.Request) (rl.M, error) {
			return rl.M{
				"hello": "world",
			}, nil
		}))
	r.Get("/app", indexLayout.Handle("app",
		func(w http.ResponseWriter, r *http.Request) (rl.M, error) {
			appData := struct {
				Title string `json:"title"`
			}{
				Title: "Hello from server for the svelte component",
			}

			d, err := json.Marshal(&appData)
			if err != nil {
				return nil, fmt.Errorf("%v: %w", err, fmt.Errorf("encoding failed"))
			}

			return rl.M{
				"Data": string(d), // notice struct is converted into a string
			}, nil
		}))

	r.Route("/samples", func(r chi.Router) {
		r.Get("/", indexLayout.HandleStatic("samples/list"))
		r.Get("/sidemenu", indexLayout.HandleStatic("samples/sidemenu"))
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
