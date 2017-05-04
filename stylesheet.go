package gogular

import (
	"github.com/aymerick/douceur/css"
	"golang.org/x/net/html"
	"io/ioutil"
	"github.com/PuerkitoBio/goquery"
	"github.com/aymerick/douceur/parser"
)

type Stylesheet struct {
	Path string
	*css.Stylesheet
	NodePointers []NodeRule
}

type NodeRule struct {
	*html.Node
	*css.Rule
}

func ParseStylesheet(dir string, urls ...string) *Stylesheet {
	sheet := &Stylesheet{}
	cssString := ""

	for _, url := range urls {
		styleBytes, err := ioutil.ReadFile(dir + "/" + url)
		if err != nil {
			ConsoleLog(err)
		}
		cssString += string(styleBytes) + "\n"
	}

	cssStylesheet, err := parser.Parse(cssString)
	if err != nil {
		ConsoleLog(err)
	}

	// Get nodes and corresponding css rules
	sheet.Stylesheet = cssStylesheet
	sheet.NodePointers = []NodeRule{}

	return sheet
}

func (s *Stylesheet) EmbedNode(node *html.Node) {
	docQuery := goquery.NewDocumentFromNode(node)
	for _, rule := range s.Rules {
		docQuery.Find(rule.Prelude).Each(func(i int, sel *goquery.Selection) {
			for _, node := range sel.Nodes {
				s.NodePointers = append(s.NodePointers, NodeRule{node, rule})
			}
		})
	}
}

func (s *Stylesheet) RenderShadow() string {
	classNames := map[string]bool{}
	ruleMap := map[*css.Rule]string{}
	stylesheetString := ""

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
			stylesheetString += dr.String() + "\n"
		}
		styleRule.Node.Attr = append(styleRule.Node.Attr, html.Attribute{Key: "class", Val: className})
	}

	return stylesheetString
}

func (s *Stylesheet) Render() string {
	stylesheetString := ""

	for _, styleRule := range s.NodePointers {
		stylesheetString += styleRule.Rule.String()
	}

	return stylesheetString
}
