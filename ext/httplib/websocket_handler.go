package httplib

import (
	"errors"
	"github.com/d5/tengo/v2"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
	"lightbox/ext/util"
	"net/http"
)

type WebsocketHandler struct {
	server            *httpServer
	EnableCompression bool
	ScriptFile        string
}

func NewWebsocketHandler(server *httpServer, enableCompression bool, scriptFile string) *WebsocketHandler {
	return &WebsocketHandler{server: server, EnableCompression: enableCompression, ScriptFile: scriptFile}
}

var WebsocketPlaceholder = util.PlaceHolders{
	"ws_read": nil, "ws_write": nil, "binary_message": nil, "text_message": nil, "ping_message": nil, "pong_message": nil, "close_message": nil,
}

func (w *WebsocketHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	upgrader := websocket.Upgrader{
		EnableCompression: w.EnableCompression,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	ws, err := upgrader.Upgrade(writer, request, nil)
	if err != nil {
		log.Error("upgrade http request error:", err)
		return
	}
	compiled, err := w.server.app.GetCompiled(w.ScriptFile, WebsocketPlaceholder)
	if err != nil {
		log.Error("compile script error", err)
		return
	}

	if err = compiled.Set("binary_message", websocket.BinaryMessage); err != nil {
		log.Error("set binary_message error:", err)
	}
	if err = compiled.Set("text_message", websocket.TextMessage); err != nil {
		log.Error("set text_message error:", err)
	}
	if err = compiled.Set("ping_message", websocket.PingMessage); err != nil {
		log.Error("set ping_message error:", err)
	}
	if err = compiled.Set("pong_message", websocket.PongMessage); err != nil {
		log.Error("set pong_message error:", err)
	}
	if err = compiled.Set("close_message", websocket.CloseMessage); err != nil {
		log.Error("set close_message error:", err)
	}

	err = compiled.Set("ws_read", &tengo.UserFunction{Value: func(args ...tengo.Object) (tengo.Object, error) {
		var (
			msgType int
			payload []byte
		)
		msgType, payload, err = ws.ReadMessage()
		if err != nil {
			return &tengo.Error{Value: &tengo.String{Value: err.Error()}}, nil
		} else {
			return &tengo.ImmutableMap{Value: map[string]tengo.Object{
				"type": &tengo.Int{Value: int64(msgType)},
				"body": &tengo.Bytes{Value: payload},
			}}, nil
		}
	}})

	err = compiled.Set("ws_write", &tengo.UserFunction{Value: func(args ...tengo.Object) (ret tengo.Object, err error) {
		var (
			msgType int
			payload []byte
			ok      bool
		)
		if len(args) == 1 {
			msgType = websocket.TextMessage
			if payload, ok = tengo.ToByteSlice(args[0]); !ok {
				return tengo.FromInterface(errors.New("require a binary message"))
			}
		}
		if len(args) == 2 {
			if msgType, ok = tengo.ToInt(args[0]); !ok {
				return tengo.FromInterface(errors.New("message type require a int value(default to 2)"))
			}
			if payload, ok = tengo.ToByteSlice(args[1]); !ok {
				return tengo.FromInterface(errors.New("require a binary message"))
			}
		}
		err = ws.WriteMessage(msgType, payload)
		if err != nil {
			return &tengo.Error{Value: &tengo.String{Value: err.Error()}}, nil
		} else {
			return nil, nil
		}
	}})
	err = compiled.Run()
	if err != nil {
		log.Error("run script error")
	}
}
