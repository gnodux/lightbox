package vm

import (
	"lightbox/httputil"
	"lightbox/loghub"
	"net/http"
)

const (
	sandboxName = "sandbox"
)

var (
	appletAPIs = map[string]http.Handler{
		"/exec": httputil.HandleJSONWithRequestAndVars(runFile),
		"/log":  loghub.NewWebsocketSubscribeHandler(subscriber),
	}
)

type RunFileArg struct {
	FileName string
	Args     map[string]interface{}
}
type RunFileResult struct {
	Result map[string]interface{}
}

func runFile(r *http.Request, vars map[string]string, arg *RunFileArg) (*RunFileResult, error) {
	//sandbox, ok := vars[sandboxName]
	//if !ok || sandbox == "" {
	//	return nil, errors.New("require sandbox name")
	//}
	//box, ok := manager.Get(sandbox)
	//if !ok {
	//	return nil, fmt.Errorf("sandbox [%s] not found", sandbox)
	//}
	//compiled, err := box.RunFileWithArgsContext(context.Background(), arg.FileName, arg.Args)
	//if err != nil {
	//	return nil, err
	//}
	//result := &RunFileResult{
	//	Result: map[string]interface{}{},
	//}
	//for _, v := range compiled.GetAll() {
	//	result.Result[v.Code()] = v.Value()
	//}
	//return result, nil
	return nil, nil
}
