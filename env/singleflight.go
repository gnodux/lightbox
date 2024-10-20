package env

import (
	"go/types"
	"golang.org/x/sync/singleflight"
	"sync"
)

type GroupCache[TResult any | types.Nil, TOption any | types.Nil] struct {
	group singleflight.Group
	cache map[string]TResult
	build func(TOption) (TResult, error)
	mx    sync.Mutex
}

func (g *GroupCache[TResult, TOption]) Delete(name string) {
	g.mx.Lock()
	defer g.mx.Unlock()
	if g.cache != nil {
		delete(g.cache, name)
	}
}
func (g *GroupCache[TResult, TOption]) EvictWith(cleanFn func(string, TResult)) {
	g.mx.Lock()
	defer g.mx.Unlock()
	if g.cache != nil {
		for key, itm := range g.cache {
			cleanFn(key, itm)
		}
	}
	g.cache = map[string]TResult{}
}
func (g *GroupCache[TResult, TOption]) Get(name string, option TOption) (TResult, error) {
	var ret TResult
	result, err, _ := g.group.Do(name, func() (interface{}, error) {
		g.mx.Lock()
		defer g.mx.Unlock()
		var (
			v  TResult
			e  error
			ok bool
		)
		if g.cache == nil {
			g.cache = make(map[string]TResult)
		}
		v, ok = g.cache[name]
		if ok {
			return v, nil
		}

		v, e = g.build(option)
		if e == nil {
			g.cache[name] = v
			return v, nil
		} else {
			return v, e
		}
	})
	ret = result.(TResult)
	return ret, err
}

func NewCache[TResult any, TOption any](builder func(TOption) (TResult, error)) *GroupCache[TResult, TOption] {
	return &GroupCache[TResult, TOption]{
		build: builder,
	}
}
