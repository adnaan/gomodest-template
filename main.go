package main

import (
	"fmt"
	"gomodest-template/samples"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	rl "github.com/adnaan/renderlayout"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	_ "github.com/mattn/go-sqlite3"
)

func main() {

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

	if err != nil {
		log.Fatal(err)
	}

	r := chi.NewRouter()
	r.Use(middleware.Compress(5))
	r.Use(middleware.StripSlashes)
	r.NotFound(index("404"))
	r.Get("/", index("home", rl.StaticData(rl.D{"hello": "world"})))
	r.Route("/samples", samples.Router(index))

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
