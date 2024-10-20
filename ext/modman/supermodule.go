package modman

import (
	"github.com/d5/tengo/v2"
)

type ImportFunc func(string) tengo.Importable

type ModuleMapper interface {
	tengo.ModuleGetter
	AddMap(m *tengo.ModuleMap)
	AddSourceModule(name string, src []byte)
	AddImporter(...ImportFunc)
}

type Composite struct {
	getter   tengo.ModuleGetter
	fallback []ImportFunc
}

func (u *Composite) Get(name string) tengo.Importable {
	if u != nil && u.getter != nil {
		if m := u.getter.Get(name); m != nil {
			return m
		}
	}

	if u != nil && u.fallback != nil {
		imp := ImportChain(u.fallback).Get(name)
		if imp != nil {
			return imp
		}
	}
	return nil
}

func (u *Composite) GetModuleMap() *tengo.ModuleMap {
	if mm, ok := u.getter.(*tengo.ModuleMap); ok {
		return mm
	}

	return nil
}

func (u *Composite) AddMap(m *tengo.ModuleMap) {
	if u != nil && u.getter != nil {
		if self, ok := u.getter.(*tengo.ModuleMap); ok {
			self.AddMap(m)
		}
	}
}
func (u *Composite) AddSourceModule(name string, src []byte) {
	if u != nil && u.getter != nil {
		if self, ok := u.getter.(*tengo.ModuleMap); ok {
			self.AddSourceModule(name, src)
		}
	}
}
func (u *Composite) AddImporter(imports ...ImportFunc) {
	if u != nil {
		u.fallback = append(u.fallback, imports...)
	}
}

type ImportChain []ImportFunc

func (c ImportChain) Get(name string) tengo.Importable {
	if c != nil {
		for _, f := range c {
			if m := f(name); m != nil {
				return m
			}
		}
	}
	return nil
}

func NewModule(m *tengo.ModuleMap, f ...ImportFunc) *Composite {
	return &Composite{
		getter:   m,
		fallback: f,
	}
}
