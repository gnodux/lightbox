package helplib

import (
	"fmt"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/ast"
	"github.com/gorilla/mux"
	"io/fs"
	"io/ioutil"
	"net/http"
	"os"
	"reflect"
	"testing"
)

func TestIndexFile(t *testing.T) {

	helpfs := os.DirFS("../../docs")
	var index IndexBase
	index.Index(helpfs)
	fmt.Println("keyword search")
	for _, kr := range index.Search("getpid") {
		fmt.Println(kr)
	}
	fmt.Println("advance search")
	for _, sr := range index.AdvanceSearch(Document{Body: "process id", Code: "pid"}) {
		fmt.Println(sr)
	}

}

func TestIndexDoc(t *testing.T) {
	base := IndexBase{}
	base.Index(os.DirFS("testdata"))
	fmt.Println(base.Save("testdata"))
	fmt.Println(base.Search("seed\\(*"))
}
func TestIndexLoad(t *testing.T) {
	var base IndexBase
	fmt.Println(base.Load("testdata"))
	fmt.Println(base.Search("seed\\(*"))
}

func TestIndexStruct(t *testing.T) {
	b, e := ioutil.ReadFile("testdata/stdlib-rand.md")
	if e != nil {
		panic(e)
	}
	root := markdown.Parse(b, nil)
	for _, cn := range root.GetChildren() {
		fmt.Println(reflect.TypeOf(cn))
	}
	ast.WalkFunc(root, func(node ast.Node, entering bool) ast.WalkStatus {
		//fmt.Printf("%s:%+v\n", reflect.TypeOf(node), node)
		if h, ok := node.(*ast.Heading); ok {
			for _, c := range h.Children {
				fmt.Printf("%s:%+v\n", reflect.TypeOf(c), string(c.AsLeaf().Literal))
				for _, cc := range c.GetChildren() {
					fmt.Printf("%s:%+v\n", reflect.TypeOf(cc), cc)
				}
			}
		}

		return ast.GoToNext
	})
}

func TestMDServer(t *testing.T) {
	router := mux.NewRouter()
	opt := Option{
		FileSys: os.DirFS("../../docs"),
	}
	router.HandleFunc("/{document}", NewDocumentHandler(&opt))
	http.ListenAndServe(":9099", router)
}

func TestMDParse(t *testing.T) {
	f := os.DirFS("../../docs")
	mds, _ := fs.Glob(f, "*.md")
	for _, md := range mds {
		buf, _ := fs.ReadFile(f, md)
		nodes := markdown.Parse(buf, nil)
		ast.WalkFunc(nodes, func(node ast.Node, entering bool) ast.WalkStatus {
			//fmt.Println(reflect.TypeOf(node))
			switch n := node.(type) {
			case *ast.ListItem:
				fmt.Println("list item:", string(n.Literal))
				for _, cn := range n.Children {
					switch child := cn.(type) {
					case *ast.Paragraph:
						for _, cc := range child.Children {
							switch ct := cc.(type) {
							case *ast.Code:
								fmt.Println(string(ct.Literal))
							case *ast.Paragraph:
								fmt.Println(string(ct.Literal))
							}
						}
					}
				}
				//case *ast.List:
				//fmt.Println("list:", string(n.Literal))
				//case *ast.CodeBlock:
				//	fmt.Println(reflect.TypeOf(n.Parent))
				//	fmt.Println("code block:", string(n.Literal))
				//case *ast.Code:
				//	if bytes.Contains(n.Literal, []byte{'('}) {
				//		showParent(n.Parent)
				//		fmt.Println("code:", string(n.Literal))
				//	}
				//
			}
			return ast.GoToNext
		})
	}

}
func showParent(node ast.Node) {
	if pp := node.GetParent(); pp != nil {
		fmt.Println(reflect.TypeOf(pp))
	}
	fmt.Println(reflect.TypeOf(node))

	for _, n := range node.GetChildren() {
		if l := n.AsLeaf(); l != nil {
			fmt.Println(string(l.Literal))
		}
	}
}
