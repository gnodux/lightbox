package httplib

import (
	"github.com/d5/tengo/v2"
	"github.com/gorilla/mux"
	"net/http"
)

type ScriptMiddleWare struct {
	ScriptHandler
	next http.Handler
}

func (s *ScriptMiddleWare) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	app := s.server.app
	compiled, err := app.GetCompiled(s.scriptFile, defaultHttpPlaceHolder)
	if err != nil {
		responseError(request.RequestURI, writer, 500, "compile failed", err)
		return
	}
	if err = compiled.Set("request", WrapRequest(request)); err != nil {
		responseError(request.RequestURI, writer, 500, "set request object", err)
		return
	}
	if err = compiled.Set("response", WrapResponse(writer)); err != nil {
		responseError(request.RequestURI, writer, 500, "set response object", err)
	}
	vars := mux.Vars(request)
	vm := &tengo.ImmutableMap{Value: map[string]tengo.Object{}}
	for k, v := range vars {
		vm.Value[k] = &tengo.String{Value: v}
	}
	if err = compiled.Set("vars", vm); err != nil {
		responseError(request.RequestURI, writer, 500, "set vars object", err)
	}

	process := &tengo.UserFunction{Name: "process", Value: func(args ...tengo.Object) (ret tengo.Object, err error) {
		s.next.ServeHTTP(writer, request)
		return nil, nil
	}}
	err = compiled.Set("process", process)
	if err != nil {
		responseError(request.RequestURI, writer, 500, "set process error:", err)
	}

	err = compiled.Run()
	if err != nil {
		responseError(request.RequestURI, writer, 500, "execute before middle ware "+s.scriptFile+" error: ", err)
	}

}

func NewScriptMiddleWare(server *httpServer, srcFile string) mux.MiddlewareFunc {
	//if !filepath.IsAbs(srcFile) {
	//	srcFile, _ = filepath.Abs(srcFile)
	//}
	return func(handler http.Handler) http.Handler {
		s := &ScriptMiddleWare{}
		s.server = server
		s.scriptFile = srcFile
		s.next = handler
		return s
	}
}
