package httplib

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"net/http"
	"path/filepath"
	"strings"
)

type ScriptDirHandler struct {
	scriptDir string
	server    *httpServer
}

func NewScriptDirHandler(scriptDir string, server *httpServer) *ScriptDirHandler {
	return &ScriptDirHandler{scriptDir: scriptDir, server: server}
}

func responseError(url string, writer http.ResponseWriter, code int, message string, err error) {
	log.Error("get compile action error:", err)
	writer.WriteHeader(code)
	_, err = writer.Write([]byte(fmt.Sprintf("reqeust url %s :%s:%s", url, message, err)))
	if err != nil {
		log.Error("write response error:", err)
	}
}

func (t *ScriptDirHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	srcFile := filepath.Join(t.scriptDir, request.URL.Path)
	log.Infof("request %s,target script: %s:", request.RequestURI, srcFile)
	if !strings.HasSuffix(srcFile, ".tengo") {
		srcFile = srcFile + ".tengo"
	}
	compiled, err := t.server.app.GetCompiled(srcFile, defaultHttpPlaceHolder)
	if err != nil {
		responseError(request.RequestURI, writer, 500, "compile failed", err)
		return
	}
	if err = compiled.Set("request", WrapRequest(request)); err != nil {
		responseError(request.RequestURI, writer, 500, "set request object", err)
		return
	}
	if err = compiled.Set("r", WrapRequest(request)); err != nil {
		responseError(request.RequestURI, writer, 500, "set request object", err)
		return
	}
	if err = compiled.Set("response", WrapResponse(writer)); err != nil {
		responseError(request.RequestURI, writer, 500, "set response object", err)
	}
	if err = compiled.Set("w", WrapResponse(writer)); err != nil {
		responseError(request.RequestURI, writer, 500, "set response object", err)
	}
	if err = compiled.Run(); err != nil {
		responseError(request.RequestURI, writer, 500, "run action", err)
	}
}
