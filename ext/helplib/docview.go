package helplib

import (
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	"github.com/gorilla/mux"
	"io/fs"
	"net/http"
	"regexp"
	"strings"
)

type Option struct {
	FileSys fs.FS
	Head    []byte //在<head></head>标签中嵌入的内容
	CSS     string
	Prefix  []byte //自定义页头<body>及之前,默认为空，如果为空则生成完成的html
	Suffix  []byte //自定义页尾</body>及之后
}

var (
	urlRe = regexp.MustCompile("\\(https://github.com/d5/tengo/blob/master/docs/(.*?).md\\)")
)

func NewDocumentHandler(opt *Option) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		docName := vars["document"]
		if strings.HasSuffix(docName, ".html") {
			docName = strings.ReplaceAll(docName, ".html", ".md")
		}
		data, err := fs.ReadFile(opt.FileSys, docName)
		if err != nil {
			w.WriteHeader(404)
			w.Write([]byte("read document " + docName + " failed:" + err.Error()))
			return
		}
		data = urlRe.ReplaceAll(data, []byte(`($1.md)`))
		extensions := parser.CommonExtensions | parser.AutoHeadingIDs | parser.FencedCode
		p := parser.NewWithExtensions(extensions)
		option := html.RendererOptions{
			Flags: html.CommonFlags,
			Head:  opt.Head,
			CSS:   opt.CSS,
		}
		if opt.Prefix == nil || opt.Suffix == nil {
			option.Flags = option.Flags | html.CompletePage
		}
		render := html.NewRenderer(option)
		htmlBuf := markdown.ToHTML(data, p, render)
		w.Header().Set("Content-Type", "text/html")
		w.Write(opt.Prefix)
		w.Write(htmlBuf)
		w.Write(opt.Suffix)
	}
}
