package gohotwired

import (
	"html/template"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
)

func contains(arr []string, s string) bool {
	for _, a := range arr {
		if a == s {
			return true
		}
	}
	return false
}

type OnMount func(r *http.Request) (int, interface{})
type ViewOption func(opt *viewOpt)

type viewOpt struct {
	errorPage         string
	layout            string
	layoutContentName string
	partials          []string
	extensions        []string
	funcMap           template.FuncMap
	onMountFunc       OnMount
	eventHandlers     map[string]EventHandler
}

func WithLayout(layout string) ViewOption {
	return func(o *viewOpt) {
		o.layout = layout
	}
}

func WithLayoutContentName(layoutContentName string) ViewOption {
	return func(o *viewOpt) {
		o.layoutContentName = layoutContentName
	}
}

func WithPartials(partials ...string) ViewOption {
	return func(o *viewOpt) {
		o.partials = partials
	}
}

func WithExtensions(extensions ...string) ViewOption {
	return func(o *viewOpt) {
		o.extensions = extensions
	}
}

func WithFuncMap(funcMap template.FuncMap) ViewOption {
	return func(o *viewOpt) {
		o.funcMap = funcMap
	}
}

func WithOnMount(onMountFunc OnMount) ViewOption {
	return func(o *viewOpt) {
		o.onMountFunc = onMountFunc
	}
}

func WithErrorPage(errorPage string) ViewOption {
	return func(o *viewOpt) {
		o.errorPage = errorPage
	}
}

func WithEventHandlers(eventHandlers map[string]EventHandler) ViewOption {
	return func(o *viewOpt) {
		o.eventHandlers = eventHandlers
	}
}

func find(p string, extensions []string) []string {
	var files []string

	fi, err := os.Stat(p)
	if os.IsNotExist(err) {
		return files
	}
	if !fi.IsDir() {
		if !contains(extensions, filepath.Ext(p)) {
			return files
		}
		files = append(files, p)
		return files
	}
	err = filepath.WalkDir(p, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if contains(extensions, filepath.Ext(d.Name())) {
			files = append(files, path)
		}
		return nil
	})

	if err != nil {
		panic(err)
	}

	return files
}
