package gogular

import (
	"io"
	"golang.org/x/net/html"
	"html/template"
	"fmt"
	"github.com/PuerkitoBio/goquery"
)

type TagType int
type AttributeType int

type Component struct {
	App           *App
	Configuration *ComponentConfiguration

	Id string

	Document *goquery.Document

	Styles []*Style
	*TmpFile

	KeepBodyWrapper bool
	AppendStyles    bool
}

type ComponentConfiguration struct {
	Selector string

	Template    string
	TemplateUrl string

	Styles    []string
	StyleUrls []string
	Shadow    bool
}

type Data struct {
	Attr map[string]string
}

func newId(i int) string {
	id := ""

	if i < len(letterBytes) {
		id += string(letterBytes[i])
		return id
	}

	id += string(letterBytes[i%len(letterBytes)])
	i -= len(letterBytes)

	return newId(i) + id
}

func (a *App) NewComponent(dir string, conf *ComponentConfiguration) *Component {
	c := &Component{}

	c.Configuration = conf

	c.TmpFile = a.NewTempFile(dir + "/" + conf.TemplateUrl)

	c.Id = newId(a.LastComponentIndex)

	a.LastComponentIndex++

	c.Styles = []*Style{}
	for _, fileName := range c.Configuration.StyleUrls {
		s := a.NewStyle(fileName, dir, c.Configuration.Shadow)
		s.Parse()
		c.Styles = append(c.Styles, s)
	}

	c.App = a

	c.Parse()
	// 1. Get styling in order
	c.EmbedStyles()

	return c
}

func (c *Component) Parse() {
	var err error

	f := c.Open()
	defer f.Close()

	doc, err := goquery.NewDocumentFromReader(f)
	if err != nil {
		fmt.Println(err)
	}

	c.Document = doc
}

/**
	Compiles components
 */
func (c *Component) Execute(data Data) {
	// Get to the last component first

	c.Document.Find("*").Each(func(_ int, s *goquery.Selection) {
		node := s.Nodes[0]
		if node.Type == html.ElementNode {
			if c2, is := c.App.Components[node.Data]; is {
				data := Data{}
				data.Attr = getAttributes(node.Attr)

				c2.Document.Find("content").First().ReplaceWithSelection(s.Children())

				c2.Execute(data)

				c.Styles = append(c.Styles, c2.Styles...)

				buf := c2.TmpFile.GetBuffer()

				s.ReplaceWithHtml(buf.String())
			}
		}
	})

	//fmt.Println(c.Configuration.Selector, c.ReadStyles())

	f := c.Create()
	defer f.Close()

	c.ExecuteTemplate(f, data)
}

func getAttributes(attrs []html.Attribute) map[string]string {
	attrMap := map[string]string{}
	for _, attr := range attrs {
		attrMap[attr.Key] = attr.Val
	}
	return attrMap
}

func (c *Component) EmbedStyles() {
	for _, style := range c.Styles {
		style.EmbedNodes(c.Id, c.Document)
	}
}

const (
	trailStart = "<html><head></head><body>"
	trailEnd   = "</body></html>"
)

func (c *Component) ReadStyles() string {
	style := ""
	for _, s := range c.Styles {
		style += s.Stylesheet.String()
	}
	return style
}

func (c *Component) ExecuteTemplate(w io.Writer, data Data) {
	if c.AppendStyles {
		style := c.ReadStyles()
		c.Document.Find("head").AppendHtml("<style>" + style + "</style>")
	}

	doc, err := c.Document.Html()
	if err != nil {
		fmt.Println(err)
	}

	if !c.KeepBodyWrapper {
		if doc[:len(trailStart)] == trailStart {
			doc = doc[len(trailStart):]
		}
		if doc[len(doc)-len(trailEnd):] == trailEnd {
			doc = doc[:len(doc)-len(trailEnd)]
		}
	}

	s1 := `{{define "` + c.Configuration.Selector + `"}}` + doc + `{{end}}`

	t, err := template.New(c.Configuration.Selector).Parse(s1)
	if err != nil {
		fmt.Print(err)
	}
	t.Execute(w, data)
}
