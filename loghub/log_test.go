package loghub

import (
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
	"testing"
	"time"
)

func TestLogWriter(t *testing.T) {
	router := mux.NewRouter()
	subscribe := NewLogSubscriber()
	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stderr)
	router.HandleFunc("/log/{sandbox}", NewWebsocketSubscribeHandler(subscribe))
	go func() {
		for {
			time.Sleep(50 * time.Millisecond)
			log.WithField("sandbox", "DEFAULT").Info("msg time:", time.Now())
		}
	}()
	http.ListenAndServe(":9091", router)
}
