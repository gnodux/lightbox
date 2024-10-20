package httplib

import (
	"errors"
	"github.com/d5/tengo/v2"
	"github.com/d5/tengo/v2/stdlib"
	"github.com/gorilla/mux"
	"net/http"
)

type ScriptHandler struct {
	scriptFile string
	server     *httpServer
}

type HttpFunc func(*ScriptHandler, http.ResponseWriter, *http.Request)

func NewScriptHandler(scriptFile string, server *httpServer) *ScriptHandler {
	return &ScriptHandler{scriptFile: scriptFile, server: server}
}

func (s *ScriptHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {

	app := s.server.app
	app.Logger.WithField("url", request.RequestURI).WithField("method", request.Method).Debug("handle http request")

	r := WrapRequest(request)
	w := WrapResponse(writer)
	vars := mux.Vars(request)
	varsMap := &tengo.ImmutableMap{Value: map[string]tengo.Object{}}
	jsonWriter := writeJsonFunc(s, writer, request)
	headerFn := &tengo.UserFunction{
		Value: stdlib.FuncASSRE(func(key string, value string) error {
			writer.Header().Set(key, value)
			return nil
		}),
	}
	for k, v := range vars {
		varsMap.Value[k] = &tengo.String{Value: v}
	}
	_, err := app.RunFile(s.scriptFile, map[string]interface{}{
		"request":    r,
		"r":          r,
		"response":   w,
		"w":          w,
		"view":       viewFunc(s, writer, request),
		"get_json":   readJsonFunc(s, writer, request),
		"set_json":   jsonWriter,
		"write_json": jsonWriter,
		"vars":       varsMap,
		"get_var": stdlib.FuncASRS(func(key string) string {
			return vars[key]
		}),
		"get_header": stdlib.FuncASRS(func(key string) string {
			return request.Header.Get(key)
		}),
		"set_status": stdlib.FuncAIR(func(code int) {
			writer.WriteHeader(code)
		}),
		"set_header": headerFn,
		//提示：使用UserFunction会导致报错，tengo中的Set方法会检查参数类型，但是UserFunction的参数switch不是tengo.Object，所以会报错
		//解决办法就是直接使用CallableFunc或强制转换成tengo.Object
		"write": func(args ...tengo.Object) (ret tengo.Object, err error) {
			total := 0
			for _, arg := range args {
				v := tengo.ToInterface(arg)
				count := 0
				switch vv := v.(type) {
				case string:
					count, err = writer.Write([]byte(vv))
				case []byte:
					count, err = writer.Write(vv)
				}
				if err != nil {
					return tengo.FromInterface(err)
				}
				total += count
			}
			return tengo.FromInterface(total)
		},
	})
	if err != nil {
		responseError(request.RequestURI, writer, 500, "run script handler error", err)
	}

	//compiled, err := app.GetCompiled(s.scriptFile, defaultHttpPlaceHolder)
	//if err != nil {
	//	responseError(request.RequestURI, writer, 500, "compile ", err)
	//	return
	//}
	//
	//if err = compiled.Set("view", viewFunc(s, writer, request)); err != nil {
	//	responseError(request.RequestURI, writer, 500, "set view function", err)
	//	return
	//}
	//
	//if err = compiled.Set("get_json", readJsonFunc(s, writer, request)); err != nil {
	//	responseError(request.RequestURI, writer, 500, "set get_json function", err)
	//	return
	//}
	//if err = compiled.Set("set_json", writeJsonFunc(s, writer, request)); err != nil {
	//	responseError(request.RequestURI, writer, 500, "set set_json function", err)
	//	return
	//}
	//if err = compiled.Set("request", WrapRequest(request)); err != nil {
	//	responseError(request.RequestURI, writer, 500, "set request object", err)
	//	return
	//}
	//if err = compiled.Set("r", WrapRequest(request)); err != nil {
	//	responseError(request.RequestURI, writer, 500, "set request object", err)
	//	return
	//}
	//if err = compiled.Set("response", WrapResponse(writer)); err != nil {
	//	responseError(request.RequestURI, writer, 500, "set response object", err)
	//}
	//if err = compiled.Set("w", WrapResponse(writer)); err != nil {
	//	responseError(request.RequestURI, writer, 500, "set response object", err)
	//}
	//
	//vars := mux.Vars(request)
	//vm := &tengo.ImmutableMap{Value: map[string]tengo.Object{}}
	//for k, v := range vars {
	//	vm.Value[k] = &tengo.String{Value: v}
	//}
	//if err = compiled.Set("vars", vm); err != nil {
	//	responseError(request.RequestURI, writer, 500, "set vars object", err)
	//}
	//if err = compiled.Run(); err != nil {
	//	responseError(request.RequestURI, writer, 500, "run script file:", err)
	//}
}

func viewFunc(handler *ScriptHandler, w http.ResponseWriter, r *http.Request) tengo.Object {
	return &tengo.UserFunction{Value: func(args ...tengo.Object) (ret tengo.Object, err error) {
		if len(args) != 2 {
			return nil, tengo.ErrWrongNumArguments
		}

		ctx := NewHttpContext(handler.server, w, r, mux.Vars(r))
		viewName, ok := tengo.ToString(args[0])
		if !ok {
			return nil, errors.New("view name not a string")
		}
		ctx.Vars["view"] = viewName
		data := tengo.ToInterface(args[1])
		err = handler.server.converter.Write(data, ctx, "text/html")
		return nil, err
	}}
}

func readJsonFunc(handler *ScriptHandler, w http.ResponseWriter, r *http.Request) tengo.Object {
	return &tengo.UserFunction{Value: func(args ...tengo.Object) (ret tengo.Object, err error) {
		if len(args) != 0 {
			return nil, tengo.ErrWrongNumArguments
		}

		ctx := NewHttpContext(handler.server, w, r, mux.Vars(r))
		data, err := handler.server.converter.Read(ctx, "application/json")
		if err != nil {
			return nil, err
		}
		return tengo.FromInterface(data)
	}}
}
func writeJsonFunc(handler *ScriptHandler, w http.ResponseWriter, r *http.Request) tengo.Object {
	return &tengo.UserFunction{Value: func(args ...tengo.Object) (ret tengo.Object, err error) {
		if len(args) != 1 {
			return nil, tengo.ErrWrongNumArguments
		}

		ctx := NewHttpContext(handler.server, w, r, mux.Vars(r))
		data := tengo.ToInterface(args[0])
		err = handler.server.converter.Write(data, ctx, "application/json")
		if err != nil {
			return nil, err
		}
		return nil, nil
	}}
}
