package gogular

import (
	_"cloud.google.com/go/datastore"
	"net/http"
	"github.com/ales6164/sp/core"
)

const appDir = "web/app"
const distDir = "web/dist"
const tmpDir = "web/.tmp"

var builder *core.Builder

func main() {
	fs := http.FileServer(http.Dir("web/dist"))

	http.Handle("/", fs)
	http.HandleFunc("/build", build)

	http.ListenAndServe(":3000", nil)
}

func build(w http.ResponseWriter, r *http.Request) {
	builder = core.Init(appDir, distDir, tmpDir)
	builder.ClearTmpDir()
	builder.ClearDistDir()

	doc := builder.LoadDocument(builder.IndexFile)
	doc.Render()
	doc.Minify(core.HTML, core.CSS)
	doc.Compile()

	core.ConsoleLog("Analysis", doc.Analysis)
}

/*type Post struct {
	Title       string
	Body        string `datastore:",noindex"`
	PublishedAt time.Time
}*/

/*func Post() {
  keys := []*datastore.Key{
    datastore.NameKey("Post", "post1", nil),
    datastore.NameKey("Post", "post2", nil),
  }
  posts := []*Post{
    {Title: "Post 1", Body: "...", PublishedAt: time.Now()},
    {Title: "Post 2", Body: "...", PublishedAt: time.Now()},
  }
  if _, err := client.PutMulti(ctx, keys, posts); err != nil {
    log.Fatal(err)
  }
}*/
