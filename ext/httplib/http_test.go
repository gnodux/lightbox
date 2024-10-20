package httplib

import (
	"fmt"
	"github.com/gorilla/mux"
	"io/ioutil"
	"lightbox/ext/util"
	"net/http"
	"path/filepath"
	"testing"
)

func TestFilePathGlob(t *testing.T) {
	if ff, err := filepath.Glob("../*.go"); err == nil {
		for _, f := range ff {
			t.Log(f)
		}
	}
}

func TestServer(t *testing.T) {
	mux := http.NewServeMux()
	s := &httpServer{}
	s.Addr = ":8091"
	s.Handler = mux
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./"))))
	err := s.ListenAndServe()
	if err != nil {
		return
	}
}
func TestCookie(t *testing.T) {
	cookie := &http.Cookie{}
	cookie.Name = "name"
	fmt.Printf("T:%T\n", cookie)
	fmt.Printf("v:%v\n", cookie)
	fmt.Printf("+v:%+v\n", cookie)
	fmt.Printf("#v:%#v\n", cookie)
	ref := util.NewReflectProxy(cookie)
	fmt.Println(ref)
}
func TestFS(t *testing.T) {
	fs := http.Dir("./")
	file, err := fs.Open("./badger.go")
	file.Stat()
	if err == nil {
		if bytes, err := ioutil.ReadAll(file); err == nil {
			fmt.Println(bytes)
		}
	}
}

func TestServerRoute(t *testing.T) {
	router := mux.NewRouter()
	app := router.PathPrefix("/app")
	app.Path("/{sandbox}/log").Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(fmt.Sprintf("sandbox log:%+v", mux.Vars(r))))

	}))

	http.ListenAndServe(":8099", router)
}
