package helplib

import (
	"encoding/json"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/ast"
	"io/fs"
	"lightbox/httputil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// tips: yaml序列化中遇到多行文本出现错误，原因不明，官方issue: https://github.com/go-yaml/yaml/issues/861

var (
	indexRe = regexp.MustCompile("(?m)-\\s+`(\\S+)\\((.*)\\)\\s+=>\\s+(\\S+)`:\\s+(.*?)\\.")
)

type DocumentType = int

const (
	MethodDoc DocumentType = 0
	ConstDoc  DocumentType = 1 << iota
	MemberDoc
)

type Document struct {
	Type   DocumentType
	Module string
	Object string
	Code   string
	Body   string
	Desc   string
}

func (md *Document) String() string {
	return md.Body
}

func Match(pattern string, val string) func() bool {
	return func() bool {
		b, _ := regexp.Match(pattern, []byte(val))
		return b
	}
}

func All(matches ...func() bool) bool {
	for _, m := range matches {
		if !m() {
			return false
		}
	}
	return true
}

//IndexBase 为markdown文档创建索引
type IndexBase []*Document

func (mi *IndexBase) Search(keyword string) []*Document {
	return mi.AdvanceSearch(Document{Body: keyword})
}
func (mi *IndexBase) AdvanceSearch(filter Document) []*Document {
	var result []*Document
	for _, m := range *mi {
		if All(Match(filter.Module, m.Module), Match(filter.Desc, m.Desc), Match(filter.Object, m.Object),
			Match(filter.Body, m.Body), Match(filter.Code, m.Code)) {
			result = append(result, m)
		}
	}
	return result
}
func (mi *IndexBase) Save(name string) error {
	idxFile := filepath.Join(name, "index.json")
	of, err := os.Create(idxFile)
	if err != nil {
		return err
	}
	defer of.Close()
	encoder := json.NewEncoder(of)
	return encoder.Encode(mi)

}
func (mi *IndexBase) Load(dir string) error {
	*mi = IndexBase{}
	idxFile := filepath.Join(dir, "index.json")
	f, err := os.Open(idxFile)
	if err != nil {
		return err
	}
	defer f.Close()
	decoder := json.NewDecoder(f)
	return decoder.Decode(mi)
}
func (mi *IndexBase) Index(f fs.FS) error {
	mds, err := fs.Glob(f, "*.md")
	if err != nil {
		return err
	}
	for _, md := range mds {
		data, err := fs.ReadFile(f, md)
		if err != nil {
			return err
		}
		node := markdown.Parse(data, nil)

		ast.WalkFunc(node, func(sub ast.Node, entering bool) ast.WalkStatus {
			var doc *Document
			switch c := sub.(type) {
			case *ast.Code:
				doc = genCodeDoc(md, node, c)
			}
			if doc != nil {
				*mi = append(*mi, doc)
			}
			return ast.GoToNext
		})
	}
	return nil
}

func NewSearchHandler(idx IndexBase) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		filter := &Document{}
		if err := httputil.ReadJSON(r, filter); err != nil {
			kw := r.URL.Query().Get("q")
			filter.Body = kw
		}
		result := idx.AdvanceSearch(*filter)
		httputil.WriteJSON(w, 200, result)
	}
}
func genDesc(doc *Document, code *ast.Code) {
	if code.Parent == nil {
		return
	}
	nodes := code.Parent.GetChildren()
	for idx, cur := range nodes {
		if cur == code && idx+1 < len(nodes) {
			for _, next := range nodes[idx+1:] {
				switch n := next.(type) {
				case *ast.Text:
					doc.Desc += string(n.Literal)
				case *ast.Code:
					doc.Desc += string(n.Literal)
				}
			}
			break
		}
	}
	p := code.GetParent()
	for {
		if p == nil {
			break
		}
		switch n := p.(type) {
		case *ast.List:
			if n.Parent != nil {
				var prev ast.Node
				for _, h := range n.Parent.GetChildren() {
					if h == n && prev != nil {
						if hp, ok := prev.(*ast.Heading); ok && len(hp.Children) > 0 {
							if txt, ok := hp.Children[0].(*ast.Text); ok {
								doc.Object = string(txt.Literal)
							}
						}
						return
					}
					prev = h
				}
			}
		}
		p = p.GetParent()
	}

}
func genCodeDoc(name string, root ast.Node, code *ast.Code) *Document {
	doc := &Document{}
	doc.Module = strings.ReplaceAll(strings.ReplaceAll(name, ".md", ""), "stdlib-", "")
	doc.Code = string(code.Literal)
	genDesc(doc, code)
	doc.Body = doc.Code + doc.Object + doc.Desc
	return doc
}
