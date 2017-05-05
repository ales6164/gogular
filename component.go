package gogular

import (
	"io"
	"golang.org/x/net/html"
	"bytes"
	"html/template"
	"fmt"
	"net/http"
	"strings"
)

type TagType int
type AttributeType int

type Component struct {
	App           *App
	Configuration *ComponentConfiguration

	Node     *html.Node
	Template *template.Template
	*ComponentData

	Styles []*Style

	*File
}

type ComponentConfiguration struct {
	Selector string

	Template    string
	TemplateUrl string

	Styles    []string
	StyleUrls []string
}

type ComponentData struct {
	Attributes map[string]string
	Content    template.HTML
}

func (a *App) NewComponent(dir string, shadow bool) *Component {
	component := Component{}

	component.Configuration = &ComponentConfiguration{}
	readConfiguration(dir+"/config.json", component.Configuration)

	component.ComponentData = &ComponentData{}
	component.File = NewFile(HTML, component.Configuration.TemplateUrl, dir, a.TmpDirectory)

	component.Styles = []*Style{}
	for _, fileName := range component.Configuration.StyleUrls {
		s := a.NewStyle(fileName, dir, shadow)
		s.Parse()
		component.Styles = append(component.Styles, s)
	}

	component.App = a

	component.Parse()

	return &component
}

func (c *Component) Parse() {
	var err error
	buf := c.OpenFileBuffer()
	str := "<template>" + buf.String() + "</template>"

	var r io.Reader
	r = strings.NewReader(str)

	// Parse component
	c.Node, err = html.Parse(r)
	if err != nil {
		fmt.Print(err)
	}

	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "template" {
			c.Node = n
			return
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(c.Node)

	c.EmbedStyles()
}

func (comp *Component) LoadTree(node *html.Node) {
	if node.Type == html.ElementNode {
		// Check if the tag name belongs to components selector
		if component, isComponent := comp.App.Components[node.Data]; isComponent {
			// Replace current node with corresponding component

			if component.Attributes != nil {
				component.Attributes = map[string]string{}

				for _, attr := range node.Attr {
					component.Attributes[attr.Key] = attr.Val
				}

				// Write node content to string
				contentWriter := new(bytes.Buffer)
				var f func(*html.Node)
				f = func(n *html.Node) {
					if n.Parent == node {
						html.Render(contentWriter, n)
					}
					for c := n.FirstChild; c != nil; c = c.NextSibling {
						f(c)
					}
				}
				f(node)
				component.Content = template.HTML(contentWriter.String())

				// Remove node content
				nodesToRemove := []*html.Node{}
				f = func(n *html.Node) {
					if n.Parent == node {
						nodesToRemove = append(nodesToRemove, n)
					}
					for c := n.FirstChild; c != nil; c = c.NextSibling {
						f(c)
					}
				}
				f(node)
				for _, nodeToRemove := range nodesToRemove {
					nodeToRemove.Parent.RemoveChild(nodeToRemove)
				}

				// Get template node tree

				comp.Styles = append(comp.Styles, component.Styles...)

				// Get component direct node children
				componentChildren := []*html.Node{}
				f = func(n *html.Node) {
					if n.Parent == component.Node {
						componentChildren = append(componentChildren, n)
					}

					for c := n.FirstChild; c != nil; c = c.NextSibling {
						f(c)
					}
				}
				f(component.Node)

				nodeParent := node.Parent

				// Add component node children to original node
				for _, componentNode := range componentChildren {
					componentNode.Parent = nil
					componentNode.NextSibling = nil
					componentNode.PrevSibling = nil

					nodeParent.AppendChild(componentNode)
				}

				// Remove original node
				nodeParent.RemoveChild(node)

				component.LoadTree(nodeParent)
			}


		}
	}
	for cN := node.FirstChild; cN != nil; cN = cN.NextSibling {
		comp.LoadTree(cN)
	}
}

func (c *Component) EmbedStyles() {
	for _, style := range c.Styles {
		style.EmbedNodes(c.Node)
	}
}

func (c *Component) CompileToFile(destDir string, originalName bool) {
	f := c.GetNewFile(destDir, originalName)
	defer f.Close()
	err := html.Render(f, c.Node)
	if err != nil {
		fmt.Print(err)
	}
}

func (c *Component) PreLoad() {
	tempStr := `{{define "` + c.Configuration.Selector + `"}}` + c.File.String() + `{{end}}`

	f := c.GetNewFile(c.App.DistDirectory, true)
	defer f.Close()
	f.Write([]byte(tempStr))
}

func (c *Component) Execute(w http.ResponseWriter) {
	var err error
	c.Template, err = template.New(c.Configuration.Selector).Parse(c.File.String())
	if err != nil {
		fmt.Print(err)
	}
	c.Template.Execute(w, c.ComponentData)
}

func (c *Component) render(w *io.PipeWriter) {

}
