package gogular

import (
	"os"
	"encoding/json"
	"strings"
	"golang.org/x/net/html"
	"bufio"
)

type App struct {
	*Configuration

	Themes     map[string]*Theme
	Components map[string]*Component

	Directory     string
	DistDirectory string
	TmpDirectory  string
}

type Configuration struct {
	Name      string
	IndexFile string

	DefaultTheme string

	ThemesDir string
	Themes    []string

	ComponentsDir string
	Components    []string

	Routes map[string]string
}

func Init(appDir string, distDir string, tmpDir string) *App {
	// Remove '/' from directory name
	appDir = strings.Trim(appDir, "/")
	distDir = strings.Trim(distDir, "/")
	tmpDir = strings.Trim(tmpDir, "/")

	app := &App{
		Directory:     appDir,
		DistDirectory: distDir,
		TmpDirectory:  tmpDir,
		Themes:        map[string]*Theme{},
		Components:    map[string]*Component{},
	}

	app.ReadConfiguration()
	app.ReadComponents()
	app.ReadThemes()

	return app
}

func (a *App) ReadConfiguration() {
	file, err := os.Open(a.Directory + "/config.json")
	if err != nil {
		ConsoleLog(err)
	}

	decoder := json.NewDecoder(file)

	a.Configuration = new(Configuration)
	err = decoder.Decode(a.Configuration)
	if err != nil {
		ConsoleLog(err)
	}
}

func (a *App) ReadThemes() {
	// Read components from app config and load them in builder
	for _, themeName := range a.Configuration.Themes {
		if len(themeName) > 0 {
			theme := &Theme{
				Directory: a.Directory + "/" + a.Configuration.ThemesDir + "/" + themeName,
			}
			theme.ReadConfiguration()
			theme.Stylesheet = ParseStylesheet(theme.Directory, theme.Configuration.StyleUrls...)
			a.Themes[theme.Configuration.Selector] = theme
		}
	}
}

func (a *App) ReadComponents() {
	// Read components from app config and load them in builder
	for _, componentName := range a.Configuration.Components {
		if len(componentName) > 0 {
			component := &Component{
				Directory: a.Directory + "/" + a.Configuration.ComponentsDir + "/" + componentName,
			}
			component.ReadConfiguration()
			component.Stylesheet = ParseStylesheet(component.Directory, component.Configuration.StyleUrls...)
			a.Components[component.Configuration.Selector] = component
		}
	}
}

func (a *App) ClearTmpDir() {
	os.RemoveAll(a.TmpDirectory)
}

func (a *App) ClearDistDir() {
	os.RemoveAll(a.DistDirectory)
}

func (a *App) CopyStaticDir() {
	CopyDir(a.Directory+"/static", a.DistDirectory+"/static")
}

func (a *App) LoadDocument(filePath string, theme string) *Document {
	fullPath := a.Directory + "/" + filePath
	path := strings.Split(fullPath, "/")
	fileName := path[len(path)-1]
	pathDir := path[:len(path)-1]
	documentDir := strings.Join(pathDir, "/")

	documentFile := &DocumentFile{
		DocumentType: HTML,
		FileName:     fileName,
		FileDir:      documentDir,
	}

	document := &Document{
		App:         a,
		HTMLFile:    documentFile,
		Files:       []*DocumentFile{documentFile},
		DocumentDir: documentDir,
		Analysis:    Analysis{},
	}

	file, err := os.Open(a.Directory + "/" + filePath)
	if err != nil {
		ConsoleLog(err)
	}
	defer file.Close()

	r := bufio.NewReader(file)

	node, _ := html.Parse(r)

	document.ExecuteNodeTree(node)

	t := a.Themes[theme]
	if t != nil {
		document.SetTheme(*t)
	}

	return document
}
