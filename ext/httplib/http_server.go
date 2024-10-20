package httplib

import (
	"fmt"
	"github.com/d5/tengo/v2"
	"github.com/d5/tengo/v2/stdlib"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"lightbox/ext/util"
	"lightbox/sandbox"
	"net/http"
	"os"
	"path/filepath"
)

var defaultHttpPlaceHolder = util.PlaceHolders{
	"request":  nil,
	"r":        nil,
	"response": nil,
	"w":        nil,
	"process":  nil,
	"vars":     nil,
	"view":     nil,
	"get_json": nil,
	"set_json": nil,
}

type httpServer struct {
	http.Server
	util.ReflectProxy
	tplDir      string
	certFile    string
	keyFile     string
	indexGetMap map[string]tengo.Object
	router      *mux.Router
	bootScript  string
	app         *sandbox.Applet
	root        util.RootDirectory
	converter   *MediaConverterManager
	//sync.Once
}

func (s *httpServer) TypeName() string {
	return "http_server"
}
func (s *httpServer) String() string {
	return fmt.Sprintf("http_server[addr:%s]", s.Addr)
}
func (s *httpServer) Boot() error {
	s.converter = NewDefaultConverterManager()
	for _, sc := range []string{ /*s.root.Abs("bootstrap.tengo"), s.root.Abs(s.bootScript))*/ } {
		if _, err := os.Stat(sc); err != nil {
			log.Infof("ignore execute:%s", err)
			continue
		}
		if compiled, err := s.app.GetCompiled(sc, util.DefaultPlaceHolder); err == nil {
			err = compiled.Run()
			if err != nil {
				log.Errorf("run %s error:%s", sc, err.Error())
				return err
			} else {
				log.Infof("execute startup script %s success", sc)
			}
		} else {
			log.Infof("compile %s script error:%s", sc, err.Error())
		}
	}
	return nil
}
func (s *httpServer) ListenAndServe() error {
	if err := s.Boot(); err != nil {
		log.Errorf("boot http server error:%s", err)
		return err
	}
	log.Infof("start listen & serve on %s", s.Addr)
	return s.Server.ListenAndServe()
}
func (s *httpServer) ListenAndServeTLS(cert string, key string) error {
	if err := s.Boot(); err != nil {
		log.Errorf("boot http server error:%s", err)
		return err
	}
	log.Infof("start listen & serve with (%s,%s) on %s", cert, key, s.Addr)
	return s.Server.ListenAndServeTLS(cert, key)
}

func (s *httpServer) handle(path string, scriptFile string, methods ...string) error {
	if filepath.Ext(scriptFile) == "" {
		scriptFile = scriptFile + ".tengo"
	}
	log.WithField("path", path).WithField("script", scriptFile).Infof("handle script")
	if methods != nil && len(methods) > 0 {
		s.router.Path(path).Methods(methods...).Handler(NewScriptHandler(scriptFile, s))
	} else {
		s.router.Path(path).Handler(NewScriptHandler(scriptFile, s))
	}
	return nil
}
func (s *httpServer) init() {
	s.router = mux.NewRouter()
	s.Handler = s.router
	logger := log.WithField("sandbox", s.app.Name)
	s.indexGetMap = map[string]tengo.Object{
		//snippet:name=httpserver.serve;prefix=serve;body=serve();desc=启动http服务;
		"serve": &tengo.UserFunction{Value: stdlib.FuncARE(s.ListenAndServe)},
		//snippet:name=httpserver.serve_tls;prefix=serve_tls;body=serve_tls();desc=启动http服务并使用证书;
		"serve_tls": &tengo.UserFunction{Value: stdlib.FuncASSRE(func(cert string, key string) error {
			s.certFile = cert
			s.keyFile = key
			return s.ListenAndServeTLS(cert, key)
		})},
		"close": &tengo.UserFunction{Value: stdlib.FuncARE(func() error {
			return s.Close()
		})},
		"start": &tengo.UserFunction{Value: stdlib.FuncARE(func() error {
			go func() {
				var err error
				if s.certFile != "" && s.keyFile != "" {
					err = s.ListenAndServeTLS(s.certFile, s.keyFile)
				} else {
					err = s.ListenAndServe()
				}
				if err != nil {
					log.Errorf("listen %s serve tls error:%s", s.Addr, err)
				}
				log.Infof("serve %s end", s.Addr)
			}()
			return nil
		})},
		//snippet:name=handle_static;prefix=handle_dir;body=handle_static(${1:prefix,${2:dir}});
		"handle_static": &tengo.UserFunction{
			Name: "handle_static",
			Value: stdlib.FuncASSRE(func(prefix string, dir string) error {
				dir = s.root.Abs(dir)
				s.router.PathPrefix(prefix).Handler(http.StripPrefix(prefix, http.FileServer(http.Dir(dir))))
				return nil
			}),
		},
		//snippet:name=handle_script_dir;prefix=handle_script_dir;body=handle_script_dir(${1:prefix},${2:dir})
		"handle_script_dir": &tengo.UserFunction{
			Name: "handle_script_dir",
			Value: stdlib.FuncASSRE(func(prefix string, dir string) error {
				//dir = s.root.Abs(dir)
				logger.WithField("scriptDir", dir).Info("handle script dir")
				s.router.PathPrefix(prefix).Handler(http.StripPrefix(prefix, NewScriptDirHandler(dir, s)))
				return nil
			}),
		},
		//snippet:name=httpserver.use_cors;prefix=use_cors;body=use_cors();
		"use_cors": &tengo.UserFunction{
			Name: "use_cors",
			Value: func(args ...tengo.Object) (ret tengo.Object, err error) {
				logger.Info("enable cors")
				s.router.Use(mux.CORSMethodMiddleware(s.router))
				return
			},
		},
		//snippet:name=httpserver.use;prefix=use;body=use(${1:middleware});
		"use": &tengo.UserFunction{
			Name: "use",
			Value: util.FuncASsRE(func(scripts []string) error {
				var middleWares []mux.MiddlewareFunc
				for _, script := range scripts {
					//if filepath.Ext(script) == "" {
					//	script = script + ".tengo"
					//}
					////script = s.root.Abs(script)
					logger.WithField("script", script).Info("use middleware")
					m := NewScriptMiddleWare(s, script)
					middleWares = append(middleWares, m)
				}
				s.router.Use(middleWares...)
				return nil
			}),
		},
		//snippet:httpserver.handle;prefix=handle;body=handle(${1:route},${2:script})
		"handle": &tengo.UserFunction{
			Name: "handle",
			Value: stdlib.FuncASSRE(func(path string, scriptFile string) error {
				return s.handle(path, scriptFile)
			}),
		},
		//snippet:httpserver.get;prefix=get;body=get(${1:route},${2:script})
		"get": &tengo.UserFunction{
			Name: "get",
			Value: stdlib.FuncASSRE(func(path string, scriptFile string) error {
				return s.handle(path, scriptFile, "GET")
			}),
		},
		//snippet:httpserver.post;prefix=post;body=post(${1:route},${2:script})
		"post": &tengo.UserFunction{
			Name: "get",
			Value: stdlib.FuncASSRE(func(path string, scriptFile string) error {
				return s.handle(path, scriptFile, "POST")
			}),
		},
		"put": &tengo.UserFunction{
			Name: "get",
			Value: stdlib.FuncASSRE(func(path string, scriptFile string) error {
				return s.handle(path, scriptFile, "PUT")
			}),
		},
		//snippet:name=httpserver.handle_ws;prefix=handle_ws;body=handle_ws(${1:route},${2:script});
		"handle_ws": &tengo.UserFunction{
			Name: "handle_ws",
			Value: stdlib.FuncASSRE(func(path string, scriptFile string) error {
				//if filepath.Ext(scriptFile) == "" {
				//	scriptFile = scriptFile + ".tengo"
				//}
				//scriptFile = s.root.Abs(scriptFile)
				log.WithField("script", scriptFile).WithField("path", path).Info("handle websocket")
				s.router.Path(path).Handler(NewWebsocketHandler(s, true, scriptFile))
				return nil
			}),
		},
		//snippet:name=httpserver.tpl;body=tpl($1);prefix=tpl;
		"tpl": &tengo.UserFunction{Value: func(args ...tengo.Object) (ret tengo.Object, err error) {
			if len(args) == 1 {
				if arg, ok := tengo.ToString(args[0]); ok {
					s.tplDir = arg
				}
			}
			return &tengo.String{Value: s.tplDir}, nil
		}},
		//snippet:name=httpserver.root;body=root($1);prefix=root;
		"root": &tengo.UserFunction{Value: func(args ...tengo.Object) (ret tengo.Object, err error) {
			if len(args) == 1 {
				if arg, ok := tengo.ToString(args[0]); ok {
					s.root = util.RootDirectory(arg)
				}
			}
			return &tengo.String{Value: string(s.root)}, nil
		}},
		"boot_script": &tengo.UserFunction{
			Name: "boot_script",
			Value: func(args ...tengo.Object) (ret tengo.Object, err error) {
				if len(args) == 1 {
					if arg, ok := tengo.ToString(args[0]); ok {
						s.bootScript = arg
					}
				}
				return &tengo.String{Value: string(s.bootScript)}, nil
			},
		},
		//snippet:name=http.addr;prefix=addr;body=addr($1);
		"addr": &tengo.UserFunction{Value: func(args ...tengo.Object) (ret tengo.Object, err error) {
			if len(args) == 1 {
				if arg, ok := tengo.ToString(args[0]); ok {
					s.Addr = arg
				}
			}
			return &tengo.String{Value: s.Addr}, nil
		}},
		//snippet:name=httpserver.cert;prefix=cert;body=cert($1);
		"cert": &tengo.UserFunction{Value: func(args ...tengo.Object) (ret tengo.Object, err error) {
			if len(args) == 1 {
				if arg, ok := tengo.ToString(args[0]); ok {
					s.certFile = arg
				}
			}
			return &tengo.String{Value: s.certFile}, nil
		}},
		//snippet:name=httpserver.key;prefix=key;body=key($1);
		"key": &tengo.UserFunction{Value: func(args ...tengo.Object) (ret tengo.Object, err error) {
			if len(args) == 1 {
				if arg, ok := tengo.ToString(args[0]); ok {
					s.keyFile = arg
				}
			}
			return &tengo.String{Value: s.keyFile}, nil
		}},
	}
}

func (s *httpServer) IndexGet(name tengo.Object) (tengo.Object, error) {

	key, ok := tengo.ToString(name)
	if !ok {
		return nil, tengo.ErrInvalidIndexType
	}
	if o, ok := s.indexGetMap[key]; ok {
		return o, nil
	}
	return nil, fmt.Errorf("field or method not found:%s", name)
}
func (s *httpServer) IndexSet(name, val tengo.Object) error {
	key, ok := tengo.ToString(name)
	if !ok {
		return tengo.ErrInvalidIndexType
	}
	if f, ok := s.indexGetMap[key]; ok {
		_, err := f.Call(val)
		return err
	}
	return fmt.Errorf("index set method not found:%s", key)
}

func newHttpServer(app *sandbox.Applet, args ...tengo.Object) (tengo.Object, error) {
	server := &httpServer{app: app}
	server.init()
	if len(args) == 1 {
		if m, ok := args[0].(*tengo.Map); ok {
			for k, v := range m.Value {
				err := server.IndexSet(&tengo.String{Value: k}, v)
				if err != nil {
					return nil, err
				}
			}
		}
	}
	return server, nil
}
