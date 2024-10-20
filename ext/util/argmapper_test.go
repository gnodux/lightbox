package util

import (
	"fmt"
	"github.com/d5/tengo/v2"
	"io/ioutil"
	"reflect"
	"testing"
)

type Position struct {
	Name string
	Pos  int
}

type Config struct {
	Name    string
	DBTable string
	Pos     *Position
	Job     []string
}

func TestStructPtr(t *testing.T) {
	c := &Config{}
	fmt.Println(StructFromArgs([]tengo.Object{&tengo.ImmutableMap{Value: map[string]tengo.Object{
		"Code":    &tengo.String{Value: "my name"},
		"DBTable": &tengo.String{Value: "metadata"},
		"Pos": &tengo.Map{
			Value: map[string]tengo.Object{
				"Code": &tengo.String{Value: "binlog.0075"},
				"Pos":  &tengo.Int{Value: 18},
			},
		},
		"JobDetail": &tengo.Array{Value: []tengo.Object{
			&tengo.String{Value: "job1"}, &tengo.String{Value: "job2"},
		}},
	}}}, c))
	fmt.Println(c)
}

func (c *Config) Who() string {
	return c.Name
}

func TestStructFunc(t *testing.T) {
	cfg := &Config{}
	val := reflect.ValueOf(cfg).Elem()
	for i := 0; i < val.NumMethod(); i++ {
		fmt.Println(val.Method(i).Interface())
	}

	for i := 0; i < val.NumField(); i++ {
		fmt.Println(val.Field(i))
	}
}

func TestMethodSig(t *testing.T) {
	sig := FuncSig(ioutil.ReadFile)
	fmt.Println(sig)
	v := reflect.ValueOf(ioutil.ReadFile)
	out := v.Call([]reflect.Value{reflect.ValueOf("./rootdir")})
	fmt.Println(out)
}

type Initialize interface {
	Init()
}

type A struct {
}

func (a *A) Call() {
	(Initialize(a)).Init()
}
func (a *A) Init() {
	panic("not implement")
}

type B struct {
	A
}

func (b *B) Init() {
	fmt.Println("here am I")
}
func TestH2c(t *testing.T) {
	b := &B{}
	b.Init()
}
