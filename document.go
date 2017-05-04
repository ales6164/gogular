package gogular

import (
	"os"
	"golang.org/x/net/html"
	"github.com/tdewolff/minify"
	html_min "github.com/tdewolff/minify/html"
	css_min "github.com/tdewolff/minify/css"
	"bytes"
	"io/ioutil"
	"bufio"
	"github.com/PuerkitoBio/goquery"
	"log"
)

type DocumentType int

const (
	HTML DocumentType = iota
	CSS
)

type Document struct {
	App *App

	Theme *Theme

	Node        *html.Node
	Stylesheets []*Stylesheet

	DocumentDir string

	HTMLFile *DocumentFile
	Files    []*DocumentFile

	Analysis Analysis
}

type DocumentFile struct {
	DocumentType
	FileName    string
	FileDir     string
	TmpFilePath string
	Version     int
}

/**
	Parses and executes document node tree, which can consist of components, and creates a render friendly HTML5 document node
 */
func (d *Document) ExecuteNodeTree(documentNode *html.Node) {
	d.Node = documentNode

	var exec func(*html.Node)
	exec = func(node *html.Node) {
		if node.Type == html.ElementNode {
			// Check if the tag name belongs to components selector
			if component, isComponent := d.App.Components[node.Data]; isComponent {
				// Replace current node with corresponding component
				ConsoleLog("parsing: " + component.Configuration.Selector)

				attributes := node.Attr

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
				content := contentWriter.String()

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
				tempNode, compStylesheet := component.GetNode(attributes, content)
				d.Stylesheets = append(d.Stylesheets, compStylesheet)

				// Get component direct node children
				componentChildren := []*html.Node{}
				f = func(n *html.Node) {
					if n.Parent == tempNode {
						componentChildren = append(componentChildren, n)
					}

					for c := n.FirstChild; c != nil; c = c.NextSibling {
						f(c)
					}
				}
				f(tempNode)

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

				exec(nodeParent)
			}
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			exec(c)
		}
	}
	exec(documentNode)
}

/**
	Themes the document
 */
func (d *Document) SetTheme(theme Theme) {
	d.Theme = &theme
	d.Theme.Stylesheet.EmbedNode(d.Node)
}

/**
	Writes document style node tree to a file
 */
func (d *Document) RenderStylesheet() {
	sheetStr := ""
	for _, sheet := range d.Stylesheets {
		sheetStr += sheet.RenderShadow() + "\n"
	}

	if d.Theme != nil {
		sheetStr += d.Theme.Render()
	}

	if len(sheetStr) > 0 {
		cssFile := &DocumentFile{
			DocumentType: CSS,
			FileDir:      d.App.TmpDirectory,
		}

		d.Files = append(d.Files, cssFile)

		_, newPath := cssFile.UpdateFilePath(d.App.TmpDirectory, false)

		distFile, err := os.Create(newPath)
		if err != nil {
			ConsoleLog(err)
		}
		defer distFile.Close()

		distFile.WriteString(sheetStr)
	}
}

/**
	Writes document node tree to a file
 */
func (d *Document) Render() {
	os.MkdirAll(d.App.TmpDirectory, os.ModePerm)
	os.MkdirAll(d.App.DistDirectory, os.ModePerm)

	d.RenderStylesheet()

	for _, file := range d.Files {
		switch file.DocumentType {
		case HTML:
			_, newPath := file.UpdateFilePath(d.App.TmpDirectory, false)
			distFile, err := os.Create(newPath)
			if err != nil {
				ConsoleLog(err)
			}

			err = html.Render(distFile, d.Node)
			if err != nil {
				ConsoleLog(err)
			}

			distFile.Close()
			break
		}

	}
}

func (d *Document) Minify(types ...DocumentType) {
	var doHtml, doCss bool

	for _, doType := range types {
		switch doType {
		case HTML:
			doHtml = true
			break
		case CSS:
			doCss = true
			break
		}
	}

	for _, file := range d.Files {
		switch file.DocumentType {
		case HTML:
			if doHtml {
				oldPath, newPath := file.UpdateFilePath(d.App.TmpDirectory, false)

				m := minify.New()
				m.AddFunc("text/html", html_min.Minify)

				file, err := os.Open(oldPath)
				if err != nil {
					ConsoleLog(err)
				}

				newFile, err := os.Create(newPath)
				if err != nil {
					ConsoleLog(err)
				}

				if err := m.Minify("text/html", newFile, file); err != nil {
					ConsoleLog(err)
				}

				d.Analysis.LogSize("minify_html", file, newFile)

				file.Close()
				newFile.Close()
			}
			break
		case CSS:
			if doCss {
				oldPath, newPath := file.UpdateFilePath(d.App.TmpDirectory, false)

				m := minify.New()
				m.AddFunc("text/css", css_min.Minify)

				file, err := os.Open(oldPath)
				if err != nil {
					ConsoleLog(err)
				}

				newFile, err := os.Create(newPath)
				if err != nil {
					ConsoleLog(err)
				}

				if err := m.Minify("text/css", newFile, file); err != nil {
					ConsoleLog(err)
				}

				d.Analysis.LogSize("minify_css", file, newFile)

				file.Close()
				newFile.Close()
			}
		}

	}
}

func (f *DocumentFile) UpdateFilePath(destDir string, keepName bool) (string, string) {
	oldPath := f.TmpFilePath
	if len(oldPath) == 0 {
		oldPath = f.FileName
	}

	if keepName {
		f.TmpFilePath = destDir + "/" + f.FileName
	} else {
		f.TmpFilePath = destDir + "/" + RandString(12)
	}

	f.Version++

	return oldPath, f.TmpFilePath
}

func (d *Document) Compile() {
	for _, docFile := range d.Files {
		switch docFile.DocumentType {
		case HTML:
			oldPath, newPath := docFile.UpdateFilePath(d.App.DistDirectory, true)

			file, err := ioutil.ReadFile(oldPath)
			if err != nil {
				ConsoleLog(err)
			}

			newFile, err := os.Create(newPath)
			if err != nil {
				ConsoleLog(err)
			}

			w := bufio.NewWriter(newFile)
			w.Write(file)
			w.Flush()
			break
		case CSS:
			// Get HTML file
			htmlFile, err := os.Open(d.HTMLFile.TmpFilePath)
			if err != nil {
				ConsoleLog(err)
			}

			// Get css file
			cssFileBytes, err := ioutil.ReadFile(docFile.TmpFilePath)
			if err != nil {
				ConsoleLog(err)
			}

			// Init query library on HTML file
			doc, err := goquery.NewDocumentFromReader(htmlFile)
			if err != nil {
				log.Fatal(err)
			}

			// Place style inside html
			htmlStyleBit := "<style>" + string(cssFileBytes) + "<style>"
			var hasHead bool
			doc.Find("head").Each(func(i int, s *goquery.Selection) {
				hasHead = true
				s.AppendHtml(htmlStyleBit)
			})
			if !hasHead {
				doc.PrependHtml(htmlStyleBit)
			}

			// Save HTML file
			newFile, err := os.Create(d.HTMLFile.TmpFilePath)
			if err != nil {
				ConsoleLog(err)
			}

			htmlWithStyle, err := doc.Html()
			if err != nil {
				ConsoleLog(err)
			}

			w := bufio.NewWriter(newFile)
			w.WriteString(htmlWithStyle)
			w.Flush()
		}

	}
}
