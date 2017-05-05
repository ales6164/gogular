package gogular

import (
	"os"
	"strings"
	"fmt"
	"net/http"
	"github.com/gorilla/mux"
)

type App struct {
	*Configuration

	*mux.Router

	Components map[string]*Component

	Directory     string
	DistDirectory string
	TmpDirectory  string
}

type Configuration struct {
	Name      string
	IndexFile string

	ComponentsDir string
	Components    []string

	Routes map[string]string
}

func NewApp(appDir string) *App {
	var app *App = &App{}

	app.Directory = strings.Trim(appDir, "/")
	d1 := strings.Split(app.Directory, "/")
	d2 := strings.Join(d1[:len(d1)-1], "/")
	app.DistDirectory = d2 + "/dist"
	app.TmpDirectory = d2 + "/.tmp"

	os.RemoveAll(app.DistDirectory)
	os.RemoveAll(app.TmpDirectory)

	os.Mkdir(app.DistDirectory, os.ModePerm)
	os.Mkdir(app.TmpDirectory, os.ModePerm)

	app.Configuration = &Configuration{}
	readConfiguration(appDir+"/config.json", app.Configuration)

	app.Components = map[string]*Component{}

	for _, selector := range app.Configuration.Components {
		app.Components[selector] = app.NewComponent(appDir+"/"+app.ComponentsDir+"/"+selector, true)
	}

	for k, v := range app.Routes {
		if _, ok := app.Components[v]; !ok {
			fmt.Printf("Component '%s' for route '%s' doesn't exist", v, k)
			delete(app.Components, v)
		}
	}

	for _, comp := range app.Components {
		comp.LoadTree(comp.Node)
		comp.CompileToFile(app.TmpDirectory, false)
		comp.PreLoad()
	}

	return app
}

func (a *App) ClearTmpDir() {
	os.RemoveAll(a.TmpDirectory)
}

func (a *App) ClearDistDir() {
	os.RemoveAll(a.DistDirectory)
}

func (a *App) HandleFunc(path string, pre func(http.ResponseWriter, *http.Request)) {
	var f func(string, http.ResponseWriter, *http.Request)
	f = func(path string, w http.ResponseWriter, r *http.Request) {
		c := a.Components[a.Configuration.Routes["/"]]
		c.Execute(w)
	}
	a.Router.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		pre(w, r)
		f(path, w, r)
	})
}
