package tpllib

import (
	"fmt"
	"github.com/d5/tengo/v2"
	"html/template"
	"io/fs"
	"os"
	"regexp"
	"strings"
	"testing"
)

func TestHtmlTemplate_Render(t *testing.T) {
	type testData struct {
		Name string
	}
	td := &testData{Name: "hello"}
	tp, err := NewHtmlTemplate("{{.Code}}")
	if err != nil {
		t.Error("create template error:", err)
		t.FailNow()
	}
	v, err := tp.Render(td)
	if err != nil {
		t.Error("render error:", err)
	}
	t.Log(v)
}

func TestTextTemplate_Render(t *testing.T) {
	td := map[string]tengo.Object{
		"Code": &tengo.String{Value: "hello"},
		"Age": &tengo.Map{Value: map[string]tengo.Object{
			"cal": &tengo.String{Value: "1v"},
		}},
	}
	tp, err := NewTextTemplate("{{.Code.Value}}/age:{{.Age.Value.cal.Value}}")
	if err != nil {
		t.Error("create template error:", err)
		t.FailNow()
	}
	v, err := tp.Render(td)
	if err != nil {
		t.Error("render error:", err)
	}
	t.Log(v)
}

type myModuleMap struct {
	*tengo.ModuleMap
}

func (receiver *myModuleMap) Get(name string) tengo.Importable {
	return &tengo.SourceModule{Src: []byte(`
export func(){
	return "jump,jump"
}
`)}
}

func TestModuleMap(t *testing.T) {
	s := tengo.NewScript([]byte(`
j:=import("nn")

`))
	s.SetImports(&myModuleMap{})

	compile, err := s.Compile()
	if err != nil {
		return
	}
	compile.Run()

}

func TestSnippetExtractor(t *testing.T) {

	re := regexp.MustCompile("snippet:((name=(?P<name>.*?);)|(prefix=(?P<prefix>.*?);)|(body=(?P<body>.*?);)|(desc=(?P<desc>.*?);))*")
	content := `
// SearchPrefix snippet:desc=测试的;name=badger.search(prefix);prefix=search,badger.search;body=search(${1:prefix});
func (b *badgerClient) SearchPrefix(prefix string) (map[string][]byte, error) {
	values := map[string][]byte{}
	err := b.View(func(txn *badger.Txn) error {
		opt := badger.DefaultIteratorOptions
		p := []byte(prefix)
		opt.Prefix = p
		iterator := txn.NewIterator(opt)
		for iterator.Seek(p); iterator.ValidForPrefix(p); iterator.Next() {
			itm := iterator.Item()
			err := itm.Value(func(val []byte) error {
				values[string(itm.Key())] = val
				return nil
			})
			if err != nil {
				return err
			}
		}
		defer iterator.CloseDb()
		return nil
	})
	return values, err
}

// SearchPrefix snippet:name=badger.search(prefix);prefix=search;body=search(${1:prefix});
func (b *badgerClient) SearchPrefix(prefix string) (map[string][]byte, error) {
	values := map[string][]byte{}
	err := b.View(func(txn *badger.Txn) error {
		opt := badger.DefaultIteratorOptions
		p := []byte(prefix)
		opt.Prefix = p
		iterator := txn.NewIterator(opt)
		for iterator.Seek(p); iterator.ValidForPrefix(p); iterator.Next() {
			itm := iterator.Item()
			err := itm.Value(func(val []byte) error {
				values[string(itm.Key())] = val
				return nil
			})
			if err != nil {
				return err
			}
		}
		defer iterator.CloseDb()
		return nil
	})
	return values, err
}

// SearchPrefix snippet:name=badger.search(prefix);prefix=search;body=search(${1:prefix});
func (b *badgerClient) SearchPrefix(prefix string) (map[string][]byte, error) {
	values := map[string][]byte{}
	err := b.View(func(txn *badger.Txn) error {
		opt := badger.DefaultIteratorOptions
		p := []byte(prefix)
		opt.Prefix = p
		iterator := txn.NewIterator(opt)
		for iterator.Seek(p); iterator.ValidForPrefix(p); iterator.Next() {
			itm := iterator.Item()
			err := itm.Value(func(val []byte) error {
				values[string(itm.Key())] = val
				return nil
			})
			if err != nil {
				return err
			}
		}
		defer iterator.CloseDb()
		return nil
	})
	return values, err
}
`
	result := re.FindAllStringSubmatch(content, -1)
	fmt.Println(strings.Join(re.SubexpNames(), ","))
	for _, r := range result {
		for _, rr := range r {
			fmt.Println(rr)
		}
	}
}

func TestTpl1(t *testing.T) {
	r := os.DirFS("./")
	t1 := template.New("html")
	data, _ := fs.ReadFile(r, "test.tpl.html")
	t1.New("h1").Parse(string(data))
	fmt.Printf("%#v\n", t1.Lookup("h1"))
	for _, tt := range t1.Templates() {
		fmt.Println(tt.Name())
		tt.Execute(os.Stdout, map[string]interface{}{
			"name": map[string]interface{}{
				"Value": "myvalue",
			},
		})
	}

}
