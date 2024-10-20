package syslib

import (
	"flag"
	"fmt"
	"io/fs"
	"lightbox/sandbox"
	"os"
	"path/filepath"
	"testing"
)

func Test_sysConfig(t *testing.T) {
	app, _ := sandbox.NewWithFS("default_test", os.DirFS("."))
	obj, err := sysConfig(app)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(obj)
	}
}
func Test_AppConfig(t *testing.T) {
	app, _ := sandbox.NewWithFS("default_test", os.DirFS("."))
	app.Context.Set("profile", "test")
	result, ok := app.GetConfig("profile.work-exp")
	fmt.Println(result, ok)
}
func Test_Flag(t *testing.T) {
	t.Log(flag.Arg(0))
}
func Test_FS(t *testing.T) {
	p, err := filepath.Abs("../")
	root := NewF1(p)
	sub, err := fs.Sub(root, "amqplib")
	if err != nil {
		t.Fatal(sub, err)
	}

	fmt.Println(fs.ReadDir(root, "amqplib"))
	fmt.Println(fs.ReadDir(root, "."))
	fmt.Println(fs.Glob(root, "*"))
	fmt.Println(fs.Glob(sub, "*"))
	//dirs, err := fs.ReadDir(sub, "/")
	//if err != nil {
	//	t.Fatal(err)
	//}

	//for _, d := range dirs {
	//	fmt.Println(d)
	//}
}

type F1 struct {
	f fs.FS
}

func NewF1(root string) fs.FS {
	return &F1{
		f: os.DirFS(root),
	}
}

func (f *F1) Stat(name string) (fs.FileInfo, error) {
	fmt.Println("Stat", name)
	return fs.Stat(f.f, name)
}

func (f *F1) Open(name string) (fs.File, error) {
	fmt.Println("Open:", name)
	return f.f.Open(name)
}

func (f *F1) ReadDir(name string) ([]fs.DirEntry, error) {
	fmt.Println("read dir:", name)
	return fs.ReadDir(f.f, name)
}
