package httputil

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"io"
	"net/http"
)

type JSONResponse[TPayload any] struct {
	Code    int      `json:"code"`
	Message string   `json:"message"`
	Data    TPayload `json:"data"`
}

func WriteJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	encoder := json.NewEncoder(w)
	if err := encoder.Encode(data); err != nil {
		w.Write([]byte(err.Error()))
	}
}
func ReadJSON(r *http.Request, out interface{}) error {
	decoder := json.NewDecoder(r.Body)
	return decoder.Decode(out)
}

func HandleJSONWithVars[TArg interface{}, TResponse interface{}](h func(TArg, map[string]string) (TResponse, error)) http.Handler {
	return HandleJSONWithRequestAndVars(func(r *http.Request, m map[string]string, arg TArg) (TResponse, error) {
		return h(arg, m)
	})
}
func HandleJSON[TArg interface{}, TResponse interface{}](h func(TArg) (TResponse, error)) http.Handler {
	return HandleJSONWithRequestAndVars(func(r *http.Request, vars map[string]string, arg TArg) (TResponse, error) {
		return h(arg)
	})
}

func HandleJSONWithRequestAndVars[TArg interface{}, TResponse interface{}](h func(*http.Request, map[string]string, TArg) (TResponse, error)) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			ex := recover()
			if ex != nil {
				WriteJSON(w, 500, &JSONResponse[interface{}]{
					Code:    500,
					Message: fmt.Sprintf("%v", ex),
				})
			}
		}()
		vars := mux.Vars(r)
		var arg TArg

		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&arg)
		if err != nil && err != io.EOF {
			WriteJSON(w, 200, &JSONResponse[interface{}]{
				Code:    500,
				Message: err.Error(),
				Data:    nil,
			})
			return
		}
		ret, callErr := h(r, vars, arg)
		if callErr != nil {
			WriteJSON(w, 200, &JSONResponse[interface{}]{
				Code:    500,
				Message: callErr.Error(),
				Data:    nil,
			})
		} else {
			WriteJSON(w, 200, &JSONResponse[TResponse]{
				Code:    200,
				Message: "success",
				Data:    ret,
			})
		}
	})
}
