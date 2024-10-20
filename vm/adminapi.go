package vm

import (
	"lightbox/httputil"
	"net/http"
)

var (
	adminAPIs = map[string]http.Handler{
		"create": httputil.HandleJSONWithRequestAndVars(createApplet),
	}
)

type CreateAppletArg struct {
	Name string
	Root string
	Env  map[string]string
}
type CreateAppletResponse struct {
	Status  string
	Message string
}

func createApplet(r *http.Request, vars map[string]string, arg *CreateAppletArg) (*CreateAppletResponse, error) {
	return nil, nil
}
