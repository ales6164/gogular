package gogular

import (
	"os"
	"strings"
	"fmt"
	"net/http"
	"github.com/gorilla/mux"
	"github.com/PuerkitoBio/goquery"
)

type App struct {
	*Configuration

	*mux.Router

	Components map[string]Component
	Ids        map[string]bool

	Directory     string
	DistDirectory string
	TmpDirectory  string

	LastComponentIndex int
}

type Configuration struct {
	Name      string
	IndexFile string

	StaticDirs  []string
	StaticFiles []string

	ComponentsDir string
	Components    []string

	RouteSelector string
	Routes        map[string]string
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
	os.Mkdir(app.DistDirectory+"/c", os.ModePerm)
	os.Mkdir(app.TmpDirectory, os.ModePerm)

	app.ReadConfiguration()

	app.BuildComponents()

	app.CopyStatic()

	return app
}

func (a *App) ReadConfiguration() {
	a.Configuration = &Configuration{}
	readConfiguration(a.Directory+"/config.json", a.Configuration)
}

func (a *App) CopyStatic() {
	for _, path := range a.StaticFiles {
		f := a.NewTempFile(a.Directory + "/" + path)
		f.Copy(a.DistDirectory)
	}
	for _, path := range a.StaticDirs {
		copy_folder(a.Directory+"/"+path, a.DistDirectory+"/"+path)
	}
}

func (a *App) HandleFunc(path string, pre func(http.ResponseWriter, *http.Request)) {
	var f func(string, http.ResponseWriter, *http.Request)
	f = func(path string, w http.ResponseWriter, r *http.Request) {
		c := a.Components[a.Configuration.Routes["/"]]

		c.Execute(Data{})
		c.ExecuteTemplate(w, Data{})
	}
	a.Router.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		pre(w, r)
		f(path, w, r)
	})
}

func (a *App) buildRoute(route string, index *Component, page *Component) {
	index.Parse()

	index.Document.Find(a.Configuration.RouteSelector).Each(func(_ int, s *goquery.Selection) {
		s.ReplaceWithHtml(
			`<div data-router><` + page.Configuration.Selector + `></` + page.Configuration.Selector + `></div`,
		)
	})

	index.Execute(Data{})


	if len(route) > 1 {
		index.Filename = route + "/index.html"

		os.Mkdir(a.DistDirectory+"/"+route, os.ModePerm)
	}
	index.Copy(a.DistDirectory, MinifyHtml)

	// also save bare page component
	page.Execute(Data{})
	page.Filename = route + "/template.html"
	page.Copy(a.DistDirectory, MinifyHtml)
}

func (a *App) BuildPages() {

	//c.Copy(a.DistDirectory)

	for k, v := range a.Routes {
		if c, ok := a.Components[v]; ok {
			compConf := &ComponentConfiguration{
				Selector:    "index",
				TemplateUrl: a.IndexFile,
				Shadow:      false,
			}

			index := a.NewComponent(a.Directory, compConf)
			index.KeepBodyWrapper = true
			index.AppendStyles = true
			index.Execute(Data{})
			a.buildRoute(k, index, &c)
		} else {
			fmt.Printf("Component '%s' for route '%s' doesn't exist", v, k)
		}
	}
}

func (a *App) BuildComponents() {
	a.Components = map[string]Component{}

	for _, selector := range a.Configuration.Components {
		compDir := a.Directory + "/" + a.ComponentsDir + "/" + selector
		compConf := &ComponentConfiguration{}
		readConfiguration(compDir+"/config.json", compConf)
		a.Components[selector] = *a.NewComponent(compDir, compConf)

	}
}

func (a *App) ClearTmpDir() {
	os.RemoveAll(a.TmpDirectory)
}

func (a *App) ClearDistDir() {
	os.RemoveAll(a.DistDirectory)
}
