package sandbox

import (
	"github.com/d5/tengo/v2"
	"lightbox/ext/transpile"
	"lightbox/ext/util"
	"sync"
)

type RegistryTable struct {
	regMap    map[string]Registry
	sourceMap map[string]string
	moduleMap map[string]map[string]tengo.Object
	sync.RWMutex
}

func (r *RegistryTable) AllTranspiler() []transpile.TransFunc {
	var transpiler []transpile.TransFunc
	for _, reg := range r.regMap {
		transpiler = append(transpiler, reg.GetTranspiler()...)
	}
	return transpiler
}

func (r *RegistryTable) GetRegistryMap() map[string]Registry {
	r.RLock()
	defer r.RUnlock()
	return util.MapClone[string, Registry](r.regMap)
}
func (r *RegistryTable) GetSourceMap() map[string]string {
	r.RLock()
	defer r.RUnlock()
	return util.MapClone(r.sourceMap)
}
func (r *RegistryTable) AllNames() []string {
	r.RLock()
	defer r.RUnlock()
	var names []string
	for k := range r.regMap {
		names = append(names, k)
	}
	for k := range r.sourceMap {
		names = append(names, k)
	}
	for k := range r.moduleMap {
		names = append(names, k)
	}
	return names
}

func (r *RegistryTable) WithRegistry(registries ...Registry) *RegistryTable {
	r.Lock()
	defer r.Unlock()
	for _, reg := range registries {
		r.regMap[reg.Name()] = reg
	}
	return r
}
func (r *RegistryTable) WithSourceModule(m map[string]string) *RegistryTable {
	for k, v := range m {
		r.sourceMap[k] = v
	}
	return r
}
func (r *RegistryTable) WithModule(m map[string]map[string]tengo.Object) *RegistryTable {
	for k, v := range m {
		r.moduleMap[k] = v
	}
	return r
}
func (r *RegistryTable) Remove(names ...string) {
	r.Lock()
	defer r.Unlock()
	for _, name := range names {
		delete(r.regMap, name)
	}
}
func (r *RegistryTable) GetModuleMap(app *Applet, names ...string) *tengo.ModuleMap {
	r.RLock()
	defer r.RUnlock()
	moduleMap := tengo.NewModuleMap()
	for _, name := range names {
		if reg, ok := r.regMap[name]; ok {
			m := reg.GetModule(app)
			moduleMap.AddBuiltinModule(reg.Name(), m)
		}
		if m, ok := r.moduleMap[name]; ok {
			moduleMap.AddBuiltinModule(name, m)
		}
		if src, ok := r.sourceMap[name]; ok {
			moduleMap.AddSourceModule(name, []byte(src))
		}
	}
	return moduleMap
}
func (r *RegistryTable) GetTranspiler(app *Applet, names ...string) transpile.Group {
	r.RLock()
	defer r.RUnlock()
	var trans []transpile.TransFunc
	for _, name := range names {
		if reg, ok := r.regMap[name]; ok {
			trans = append(trans, reg.GetTranspiler()...)
		}
	}
	return trans
}

func (r *RegistryTable) GetHooks(app *Applet, names ...string) []SignalHookFn {
	r.RLock()
	defer r.RUnlock()
	var hooks []SignalHookFn
	for _, name := range names {
		if reg, ok := r.regMap[name]; ok {
			hooks = append(hooks, reg.GetHooks()...)
		}
	}
	return hooks
}

func (r *RegistryTable) GetAll(app *Applet, names ...string) (*tengo.ModuleMap, transpile.Group, []SignalHookFn) {
	return r.GetModuleMap(app, names...), r.GetTranspiler(app, names...), r.GetHooks(app, names...)
}

func NewRegistryTable(registries ...Registry) *RegistryTable {
	table := new(RegistryTable)
	table.regMap = map[string]Registry{}
	table.sourceMap = map[string]string{}
	table.moduleMap = map[string]map[string]tengo.Object{}
	return table.WithRegistry(registries...)
}
