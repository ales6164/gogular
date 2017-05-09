package gogular

import (
	html "github.com/tdewolff/minify/html"
	css "github.com/tdewolff/minify/css"
	js "github.com/tdewolff/minify/js"
	"fmt"
	"io"
	"github.com/tdewolff/minify"
)

func MinifyHtml(dest io.Writer, src io.Reader) {
	m := minify.New()
	m.AddFunc("text/css", css.Minify)
	m.AddFunc("text/html", html.Minify)
	m.AddFunc("text/javascript", js.Minify)

	if err := m.Minify("text/html", dest, src); err != nil {
		fmt.Print(err)
	}
}

