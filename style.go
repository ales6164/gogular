package gogular

import (
	"github.com/aymerick/douceur/css"
	"github.com/PuerkitoBio/goquery"
	"github.com/aymerick/douceur/parser"
	"fmt"
	"github.com/tdewolff/minify"
	css_min "github.com/tdewolff/minify/css"
	"bytes"
)

type Style struct {
	Shadow bool
	*TmpFile
	*css.Stylesheet

	RuleMap map[*css.Rule]string

	LastStyleIndex int
}

type NodeRule struct {
	className string
	*css.Rule
}

func (a *App) NewStyle(filePath string, dir string, shadow bool) *Style {
	f := a.NewTempFile(dir + "/" + filePath)

	return &Style{Shadow: shadow, TmpFile: f, RuleMap: map[*css.Rule]string{}}
}

func (s *Style) Parse() {
	f := s.Open()
	defer f.Close()

	buf := new(bytes.Buffer)
	buf.ReadFrom(f)

	cssStyle, err := parser.Parse(buf.String())
	if err != nil {
		fmt.Print(err)
	}

	s.Stylesheet = cssStyle
}

func (s *Style) EmbedNodes(shadowId string, doc *goquery.Document) {
	for _, rule := range s.Rules {
		oldClassName := rule.Prelude

		doc.Find(rule.Prelude).Each(func(i int, sel *goquery.Selection) {
			if s.Shadow {
				className, ok := s.RuleMap[rule]

				if !ok {
					className = shadowId + newId(s.LastStyleIndex)
					s.LastStyleIndex++

					s.RuleMap[rule] = className

					rule.Prelude = "." + className
					rule.Selectors = []string{rule.Prelude}
				}

				if oldClassName[:1] == "." {
					sel.RemoveClass(oldClassName[1:])
				}

				sel.AddClass(className)
			}
		})
	}
}

func (s *Style) Minify(destDir string, originalName bool) {
	m := minify.New()
	m.AddFunc("text/css", css_min.Minify)

	oldF := s.Open()
	newF := s.Create()
	defer oldF.Close()
	defer newF.Close()

	if err := m.Minify("text/css", newF, oldF); err != nil {
		fmt.Print(err)
	}
}
