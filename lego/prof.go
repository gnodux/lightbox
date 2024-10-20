package main

import (
	"flag"
	"github.com/gorilla/mux"
	"github.com/pingcap/log"
	"github.com/sirupsen/logrus"
	"lightbox/loghub"
	"net/http"
	"net/http/pprof"
)

var (
	prof         = false
	profAddr     = ":8018"
	logSubscribe = false
	enableHttp   = false

	router *mux.Router = mux.NewRouter()
)

func init() {
	flag.BoolVar(&prof, "prof", false, "enable prof trace")
	flag.StringVar(&profAddr, "http_addr", ":8018", "http server(log/profile).... port")
	flag.BoolVar(&logSubscribe, "log_tail", false, "enable log subscribe")
}

func enableProf() {
	if prof {
		log.Debug("profile enabled")
		enableHttp = true
		router.HandleFunc("/debug/pprof/", pprof.Index)
		router.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
		router.HandleFunc("/debug/pprof/profile", pprof.Profile)
		router.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
		router.HandleFunc("/debug/pprof/trace", pprof.Trace)
	}
}
func enableLogger() {
	if logSubscribe {
		enableHttp = true
		subscriber := loghub.NewLogSubscriber()
		router.HandleFunc("/{sandbox}/log", loghub.NewWebsocketSubscribeHandler(subscriber))
	}

}
func startHttpServer() {
	if !enableHttp {
		return
	}
	go func() {
		server := &http.Server{
			Addr:    profAddr,
			Handler: router,
		}
		err := server.ListenAndServe()
		if err != nil {
			logrus.Error("http server error", err)
		}
	}()
}
