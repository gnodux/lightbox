package sandbox

import (
	"fmt"
	"github.com/d5/tengo/v2"
	"github.com/d5/tengo/v2/stdlib"
	"gopkg.in/yaml.v3"
	"os"
	"testing"
)

func TestApplet_Run(t *testing.T) {

	app, err := NewWithDir("my app", "../tengo-examples")
	if err != nil {
		t.Error(err)
	}
	app.WithModule(stdlib.GetModuleMap(stdlib.AllModuleNames()...))
	compiled, err := app.Run([]byte(`
fmt:=import("fmt")
fmt.println("hello")
userName:="xudong"
`), nil, "main")
	fmt.Println(err)
	if compiled != nil {
		for _, v := range compiled.GetAll() {
			fmt.Println(v.Name(), ":", v.Value())
		}
	}
}

func TestScript(t *testing.T) {
	script := tengo.NewScript([]byte(`
fmt:=import("fmt")
fmt.println("hello")
userName:="xudong"
`)).WithTrace(os.Stderr)
	script.SetImports(stdlib.GetModuleMap(stdlib.AllModuleNames()...))
	c, err := script.Compile()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(c)
}

func TestYml(t *testing.T) {
	yml1 := []byte(`
appName:
	testApp
`)
	yml2 := []byte(`
version: 1.0
env: prod234
`)
	m := make(map[string]interface{})
	yaml.Unmarshal(yml1, m)
	yaml.Unmarshal(yml2, m)
	fmt.Printf("%+v", m)

}
