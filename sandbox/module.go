package sandbox

import (
	"github.com/d5/tengo/v2"
	"lightbox/ext/transpile"
	"sync"
)

type Registry interface {
	Name() string
	AllNames() []string
	GetModule(applet *Applet, names ...string) map[string]tengo.Object
	GetTranspiler() []transpile.TransFunc
	GetHooks() []SignalHookFn

	WithTranspiler(...transpile.TransFunc) Registry
	WithHook(...SignalHookFn) Registry
}

type moduleRegistry struct {
	name        string
	allNames    []string
	entry       map[string]tengo.Object
	appFunc     map[string]UserFunction
	initializer func(applet *Applet)
	hooks       []SignalHookFn
	transpiler  []transpile.TransFunc
	once        sync.Once
	initOnce    sync.Once
}

func (m *moduleRegistry) Name() string {
	return m.name
}

func (m *moduleRegistry) GetTranspiler() []transpile.TransFunc {
	return m.transpiler
}
func (m *moduleRegistry) GetHooks() []SignalHookFn {
	return m.hooks
}

func (m *moduleRegistry) WithHook(hookFunc ...SignalHookFn) Registry {
	m.hooks = append(m.hooks, hookFunc...)
	return m
}
func (m *moduleRegistry) WithHookScript(scripts ...string) Registry {
	for _, script := range scripts {
		m.hooks = append(m.hooks, func(applet *Applet, signal Signal) {
			_, err := applet.RunFile(script, map[string]interface{}{
				"app":    applet,
				"signal": signal,
			})
			if err != nil {
				applet.Logger.Info("execute hook script error", err)
			}
		})
	}
	return m
}
func (m *moduleRegistry) WithTranspiler(transFunc ...transpile.TransFunc) Registry {
	m.transpiler = append(m.transpiler, transFunc...)
	return m
}
func (m *moduleRegistry) AllNames() []string {
	m.once.Do(func() {
		for k := range m.entry {
			m.allNames = append(m.allNames, k)
		}
		for k := range m.appFunc {
			m.allNames = append(m.allNames, k)
		}
	})
	return m.allNames
}

func (m *moduleRegistry) GetModule(applet *Applet, names ...string) map[string]tengo.Object {
	var mod = map[string]tengo.Object{}
	if len(names) == 0 {
		names = m.AllNames()
	}
	if len(names) == 1 && names[0] == "*" {
		names = m.AllNames()
	}

	//m.initOnce.Do(func() {
	//	if m.initializer != nil {
	//		m.initializer(applet)
	//	}
	//})
	for _, name := range names {
		if v, ok := m.entry[name]; ok {
			mod[name] = v
		}
		if v, ok := m.appFunc[name]; ok {
			mod[name] = &tengo.UserFunction{Value: NewUserFunction(applet, v)}
		}
	}
	return mod
}

func NewRegistry(name string, entry map[string]tengo.Object, appEntry map[string]UserFunction) Registry {
	if entry == nil {
		entry = map[string]tengo.Object{}
	}
	if appEntry == nil {
		appEntry = map[string]UserFunction{}
	}
	return &moduleRegistry{name: name, entry: entry, appFunc: appEntry}
}
