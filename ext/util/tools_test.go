package util

import (
	"fmt"
	"github.com/d5/tengo/v2"
	"github.com/d5/tengo/v2/stdlib"
	log "github.com/sirupsen/logrus"
	"sync"
	"testing"
)

func TestMergeMap(t *testing.T) {
	m1 := map[string]interface{}{
		"name": "hector",
		"age":  30,
	}
	m2 := map[string]interface{}{
		"birth": "189",
	}
	var m3 map[string]interface{}
	result := MergeMap(m1, m2, m3)
	fmt.Printf("%+v", result)
}

func TestTengoFunc(t *testing.T) {
	script := tengo.NewScript([]byte(`
os:=import("os")
text:=import("text")
fmt:=import("fmt")
pl:=fmt.println
f1:=func(name){

}`))
	m := tengo.NewModuleMap()
	m.AddMap(stdlib.GetModuleMap(stdlib.AllModuleNames()...))

	script.SetImports(m)

	run, err := script.Run()
	if err != nil {
		fmt.Println(err)
		return
	}
	for _, v := range run.GetAll() {
		switch vv := v.Value().(type) {
		case *tengo.UserFunction:
			println("user function:", vv.Name, vv.TypeName())
		case *tengo.CompiledFunction:
			println("compiled function", vv.String(), vv.NumParameters)
		default:
			println(vv)
		}
	}
}

func TestLogger(t *testing.T) {
	entry := log.WithField("sandbox", "mysandbox")
	wg := sync.WaitGroup{}
	total := 10000
	wg.Add(total)
	for i := 0; i < 10000; i++ {
		go func() {
			defer wg.Done()
			for n := 0; n < 100; n++ {
				entry.Infof("this is entry %d:%d", i, n)
			}
		}()
	}
	wg.Wait()
}
