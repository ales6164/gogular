package gogular

import (
	"github.com/aymerick/douceur/css"
	"golang.org/x/net/html"
	"github.com/PuerkitoBio/goquery"
	"github.com/aymerick/douceur/parser"
	"fmt"
	"io"
	"github.com/tdewolff/minify"
	css_min "github.com/tdewolff/minify/css"
)

type Style struct {
	Shadow       bool
	*File
	*css.Stylesheet
	NodePointers []NodeRule
}

type NodeRule struct {
	*html.Node
	*css.Rule
}

func (a *App) NewStyle(fileName string, dir string, shadow bool) *Style {
	f := NewFile(CSS, fileName, dir, a.TmpDirectory)

	return &Style{Shadow: shadow, File: f}
}

func (s *Style) Parse() {
	buf := s.OpenFileBuffer()

	cssStyle, err := parser.Parse(buf.String())
	if err != nil {
		fmt.Print(err)
	}

	s.Stylesheet = cssStyle
}

func (s *Style) EmbedNodes(node *html.Node) {
	docQuery := goquery.NewDocumentFromNode(node)
	for _, rule := range s.Rules {
		docQuery.Find(rule.Prelude).Each(func(i int, sel *goquery.Selection) {
			for _, node := range sel.Nodes {
				s.NodePointers = append(s.NodePointers, NodeRule{node, rule})
			}
		})
	}
}

func (s *Style) CompileToFile(destDir string, originalName bool) {
	pr, pw := io.Pipe()

	go func() {
		if s.Shadow {
			s.renderShadow(pw)
		} else {
			s.render(pw)
		}
		pw.Close()
	}()

	s.WriteFile(destDir, originalName, pr)
}

func (s *Style) Minify(destDir string, originalName bool) {
	m := minify.New()
	m.AddFunc("text/css", css_min.Minify)

	oldF := s.GetFile()
	newF := s.GetNewFile(destDir, originalName)

	go func() {
		if err := m.Minify("text/css", newF, oldF); err != nil {
			fmt.Print(err)
		}
	}()
}

func (s *Style) renderShadow(writer *io.PipeWriter) {
	classNames := map[string]bool{}
	ruleMap := map[*css.Rule]string{}

	for _, styleRule := range s.NodePointers {
		className, ok := ruleMap[styleRule.Rule]

		if !ok {
			for {
				newClassName := RandString(3)
				if exists := classNames[newClassName]; !exists {
					className = newClassName
					break
				}
			}

			ruleMap[styleRule.Rule] = className

			// Change rule class name and write it to stylesheet
			dr := *styleRule.Rule

			dr.Prelude = "." + className
			dr.Selectors = []string{dr.Prelude}

			writer.Write([]byte(dr.String() + "\n"))
		}
		styleRule.Node.Attr = append(styleRule.Node.Attr, html.Attribute{Key: "class", Val: className})
	}
}

func (s *Style) render(writer *io.PipeWriter) {
	for _, styleRule := range s.NodePointers {
		writer.Write([]byte(styleRule.Rule.String() + "\n"))
	}
}
