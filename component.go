package gogular

import (
	"io"
	"strings"
	"golang.org/x/net/html"
	"bytes"
	"html/template"
	"os"
	"encoding/json"
	"io/ioutil"
)

type TagType int
type AttributeType int

type Component struct {
	Directory     string
	Configuration *ComponentConfiguration

	*Stylesheet
}

type ComponentData struct {
	Attr    map[string]string
	Content template.HTML
}

type ComponentConfiguration struct {
	Selector string

	Template    string
	TemplateUrl string

	Styles    []string
	StyleUrls []string
}

func (c *Component) ReadConfiguration() {
	file, err := os.Open(c.Directory + "/config.json")
	if err != nil {
		ConsoleLog(err)
	}

	decoder := json.NewDecoder(file)

	c.Configuration = new(ComponentConfiguration)
	err = decoder.Decode(c.Configuration)
	if err != nil {
		ConsoleLog(err)
	}
}

func (c *Component) GetNode(attrs []html.Attribute, content string) (*html.Node, *Stylesheet) {

	// Parse template from original file
	t := c.ParseTemplate()

	// Insert data to parsed template
	d := c.AddTemplateData(attrs, content)

	// Get template string
	temp := c.ExecuteTemplate(t, d)

	node := c.ParseTemplateString(temp)

	c.Stylesheet.EmbedNode(node)

	return node, c.Stylesheet
}

func (c *Component) ParseTemplateString(temp string) *html.Node {
	htmlTemplate := "<template>" + temp + "</template>"

	var r io.Reader
	r = strings.NewReader(htmlTemplate)

	doc, err := html.Parse(r)
	if err != nil {
		ConsoleLog(err)
	}

	var f func(*html.Node)

	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "template" {
			doc = n
			return
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)

	return doc
}

/**
	Prepare and parse HTML component file
	- Adds functions for retrieving and modifying data
 */
func (c *Component) ParseTemplate() *template.Template {
	var templateString string = c.Configuration.Template

	if len(c.Configuration.TemplateUrl) > 0 {
		file, err := ioutil.ReadFile(c.Directory + "/" + c.Configuration.TemplateUrl)
		if err != nil {
			ConsoleLog(err)
		}
		templateString = string(file)
	}

	templateString = `{{define "` + c.Configuration.Selector + `"}}` + templateString + `{{end}}`
	t, _ := template.New(c.Configuration.Selector).Parse(templateString)

	return t
}

/**
	Executes component using go template library
	- Adds user data
	- Stylizes the HTML
 */
func (c *Component) ExecuteTemplate(template *template.Template, data *ComponentData) string {
	temp := new(bytes.Buffer)

	err := template.Execute(temp, data)
	if err != nil {
		ConsoleLog(err)
	}

	return temp.String()
}

func (c *Component) AddTemplateData(attrs []html.Attribute, content string) *ComponentData {
	data := &ComponentData{
		Attr:    map[string]string{},
		Content: template.HTML(content),
	}

	for _, attr := range attrs {
		data.Attr[attr.Key] = attr.Val
	}

	return data
}
