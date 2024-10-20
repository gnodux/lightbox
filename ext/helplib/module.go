package helplib

import (
	"bytes"
	"context"
	"errors"
	"github.com/d5/tengo/v2"
	"github.com/d5/tengo/v2/stdlib"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"lightbox/ext/util"
	"lightbox/sandbox"
	"net/http"
	"os"
	"path/filepath"
	"sync"
)

var (
	docRoot string

	defaultIndexer = IndexBase{}
	option         = &Option{}
	once           = &sync.Once{}
)

const (
	defaultAddr string = ":8012"
)

func initOnce() {
	once.Do(func() {
		err := AutoInit()
		if err != nil {
			log.Error("init help lib error", err)
		}
	})
}
func Search(m Document) []*Document {
	initOnce()
	return defaultIndexer.AdvanceSearch(m)
}

func SetRoot(dir string) (err error) {
	docRoot = dir
	root := os.DirFS(docRoot)
	defaultIndexer = IndexBase{}
	if err = defaultIndexer.Load(docRoot); err != nil {
		err = defaultIndexer.Index(root)
		if err != nil {
			return
		} else {
			defaultIndexer.Save(docRoot)
		}
	}
	option.FileSys = root
	return
}
func ViewDocHandler() http.HandlerFunc {
	return NewDocumentHandler(option)
}
func SearchHandler() http.HandlerFunc {
	return NewSearchHandler(defaultIndexer)
}
func StartHttpServer(addr string) error {
	initOnce()
	if localServer != nil {
		localServer.Shutdown(context.Background())
		localServer = nil
	}

	router := mux.NewRouter()
	router.HandleFunc("/search", SearchHandler())
	router.HandleFunc("/doc/{document}", ViewDocHandler())
	go func() {
		log.Info("start help http server")
		localServer = &http.Server{
			Addr:    addr,
			Handler: router,
		}
		err := localServer.ListenAndServe()
		if err != nil {
			log.Error("start document server")
		}
	}()
	return nil
}
func getInstallPath() string {
	ex, err := os.Executable()
	if err != nil {
		return ""
	}
	_, err = os.Stat(ex)
	if err == nil {
		if ex, err = filepath.EvalSymlinks(ex); err != nil {
			return ""
		}
	} else {
		return ""
	}
	return filepath.Dir(ex)
}

func AutoInit() error {
	p := getInstallPath()
	docPath := filepath.Join(p, "docs/")
	if _, err := os.Stat(docPath); err != nil {
		docPath = "docs"
		if _, err = os.Stat(docPath); err != nil {
			return err
		}
	}
	return SetRoot(docPath)
}

var localServer *http.Server

var Entry = sandbox.NewRegistry("help", map[string]tengo.Object{
	"set_root": &tengo.UserFunction{
		Name:  "set_root",
		Value: stdlib.FuncASRE(SetRoot),
	},
	"reindex": &tengo.UserFunction{Name: "reindex",
		Value: stdlib.FuncASRE(func(s string) error {
			if s == "" {
				s = docRoot
			}
			if s == "" {
				return errors.New("reindex document dir is empty")
			}
			var b IndexBase
			b.Index(os.DirFS(s))
			b.Save(s)
			b.Load(s)
			if s == docRoot {
				SetRoot(s)
			}
			return nil
		})},
	"serve": &tengo.UserFunction{
		Name: "serve",
		Value: func(args ...tengo.Object) (ret tengo.Object, err error) {
			addr := defaultAddr
			if len(args) != 0 {
				addr, _ = tengo.ToString(args[0])
			}
			err = StartHttpServer(addr)
			return &tengo.String{Value: addr}, err
		},
	},
	"search": &tengo.UserFunction{Name: "search", Value: func(args ...tengo.Object) (ret tengo.Object, err error) {
		if len(args) == 0 {
			err = tengo.ErrWrongNumArguments
			return
		}
		var result []*Document
		var filter = &Document{}
		switch arg := args[0].(type) {
		case *tengo.String:
			filter.Body = arg.Value
		case *tengo.ImmutableMap, *tengo.Map:
			err = util.StructFromObject(args[0], filter)
			if err != nil {
				return
			}
		}
		result = Search(*filter)
		var doc string
		if result == nil || len(result) == 0 {
			return &tengo.String{Value: "not found"}, nil
		}
		for _, r := range result {
			doc += "\n" + r.Module + ":\n" + r.Object + ":" + r.Code + r.Desc
		}

		return tengo.FromInterface(doc)
	}},
}, map[string]sandbox.UserFunction{
	"list": func(app *sandbox.Applet, args ...tengo.Object) (ret tengo.Object, err error) {
		buf := bytes.NewBuffer(nil)
		//for k, m := range lnlib.RegistryTable.GetRegistryMap() {
		//	buf.WriteString(fmt.Sprintln(k, ":"))
		//	mm := m.GetModule(app)
		//	for n, v := range mm {
		//		buf.WriteString(fmt.Sprintln("\t", n, ":", v))
		//	}
		//}
		return tengo.FromInterface(buf.String())
	},
})
