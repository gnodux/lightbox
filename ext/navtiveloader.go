package ext

import (
	"fmt"
	"github.com/d5/tengo/v2"
	"path/filepath"
	"plugin"
	"strings"
)

const (
	AuthorSymbol  = "Author"
	VersionSymbol = "BuildVersion"
	ModuleSymbol  = "Module"
	BuiltinSymbol = "Builtin"
)

type NativePlugin struct {
	Path    string
	Author  string
	Version string
	Module  map[string]map[string]tengo.Object
	Builtin []*tengo.BuiltinFunction
}

func (n *NativePlugin) Get(name string) tengo.Importable {
	if m, ok := n.Module[name]; ok {
		return &tengo.BuiltinModule{Attrs: m}
	}
	return nil
}

func (n *NativePlugin) String() string {
	return fmt.Sprintf("%s(version: %s,author: %s)", n.Path, n.Version, n.Author)
}
func (n *NativePlugin) List() string {
	b := strings.Builder{}
	b.WriteString(fmt.Sprintf("path:    %s\nauthor:  %s\nversion: %s\n", n.Path, n.Author, n.Version))
	b.WriteString("module:\n")
	for name, m := range n.Module {
		b.WriteString(fmt.Sprintf("  %s: \n", name))
		for k, o := range m {
			b.WriteString(fmt.Sprintf("    %s:%v\n", k, o))
		}
	}
	b.WriteString("builtin:\n")
	for _, bf := range n.Builtin {
		b.WriteString(fmt.Sprintf("  %s: %s\n", bf.Name, bf))
	}
	return b.String()
}

func (n *NativePlugin) GetAllModule() *tengo.ModuleMap {
	return n.GetModules(n.GetAllName()...)
}
func (n *NativePlugin) GetModules(names ...string) *tengo.ModuleMap {
	m := tengo.NewModuleMap()
	if n.Module != nil {
		for _, v := range names {
			if mm, ok := n.Module[v]; ok {
				m.AddBuiltinModule(v, mm)
			}
		}
	}
	return m
}
func (n *NativePlugin) GetAllName() []string {
	var names []string
	if n.Module != nil {
		for k := range n.Module {
			names = append(names, k)
		}
	}
	return names
}

func NewNativePlugin(libPath string) (*NativePlugin, error) {

	absPath, err := filepath.Abs(libPath)
	if err != nil {
		return nil, err
	}
	p, err := plugin.Open(absPath)
	if err != nil {
		return nil, err
	}
	plg := &NativePlugin{
		Path: absPath,
	}

	symbol, err := p.Lookup(AuthorSymbol)
	if err == nil {
		plg.Author = *symbol.(*string)
	}

	symbol, err = p.Lookup(VersionSymbol)
	if err == nil {
		plg.Version = *symbol.(*string)
	}
	symbol, err = p.Lookup(ModuleSymbol)
	if err == nil && symbol != nil {
		plg.Module = *symbol.(*map[string]map[string]tengo.Object)
	}

	symbol, err = p.Lookup(BuiltinSymbol)
	if err != nil && symbol != nil {
		plg.Builtin = *symbol.(*[]*tengo.BuiltinFunction)
	}

	return plg, nil
}
