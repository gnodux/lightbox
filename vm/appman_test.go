package vm

import (
	"fmt"
	"github.com/d5/tengo/v2/stdlib"
	"github.com/gorilla/mux"
	"lightbox/ext"
	"lightbox/httputil"
	"lightbox/sandbox"
	"net/http"
	"os"
	"testing"
)

func TestFullJsonHandler(t *testing.T) {
	type arg struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}
	type response struct {
		Vars map[string]string
		arg
	}

	router := mux.NewRouter()
	router.Handle("/log/{sandbox}", httputil.HandleJSONWithRequestAndVars(func(request *http.Request, m map[string]string, a arg) (response, error) {
		fmt.Printf("%+v", a)
		return response{Vars: m, arg: a}, nil
	}))
	router.Handle("/logm/{sandbox}", httputil.HandleJSONWithRequestAndVars(func(request *http.Request, m map[string]string, a *arg) (*response, error) {
		fmt.Printf("%+v", a)
		return &response{Vars: m, arg: *a}, nil
	}))
	http.ListenAndServe(":8088", router)
}

func TestAPI(t *testing.T) {
	router := mux.NewRouter()
	RegisterAPI(router)
	http.ListenAndServe(":8088", router)
}

func TestSubRouter(t *testing.T) {
	router := mux.NewRouter()
	sub := router.PathPrefix("/a").Subrouter()
	sub.HandleFunc("/p1", func(writer http.ResponseWriter, request *http.Request) {
		writer.Write([]byte("p1"))
	})
	sub2 := router.PathPrefix("/box/{sandbox}").Subrouter()
	sub2.HandleFunc("/exec", func(writer http.ResponseWriter, request *http.Request) {
		writer.Write([]byte(fmt.Sprint(mux.Vars(request))))
	})
	http.ListenAndServe(":8089", router)
}

func TestHost(t *testing.T) {
	t.Log(ext.RegistryTable.AllNames())
	t.Log(stdlib.AllModuleNames())
}

func TestFS(t *testing.T) {

}

func TestVirtualHost_NewApplet(t *testing.T) {
	vh := new(VirtualHost)
	vh.RepoSource = "./.packages"
	vh.RepoDest = "./.modules"
	vh.rootFS = os.DirFS("./vhs")

	opt := AppOption{Option: sandbox.Option{Name: "sandbox1", Root: "sandbox1"}, Modules: ext.RegistryTable.AllNames()}
	t.Logf("sandbox create with :%+v", opt)
	app, err := vh.NewApplet(opt)
	if err != nil {
		t.Log("error", err)
		return
	}
	fmt.Println(app)

	c, e := app.Run([]byte(`
import(log,fmt,database)
`), "main.tengo")
	fmt.Println(c, e)
	app.WithHook(sandbox.NewHook(sandbox.SigInitialized, func(applet *sandbox.Applet) error {
		fmt.Println(app.Name, "applet initialize")
		return nil
	}))
	app.WithHook(sandbox.NewHook(sandbox.SigStop, func(applet *sandbox.Applet) error {
		fmt.Println(app.Name, "shutting down...")
		return nil
	}))
	app.Shutdown("shutdown with")
}
