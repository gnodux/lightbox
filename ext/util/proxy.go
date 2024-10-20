package util

import (
	"fmt"
	"github.com/d5/tengo/v2"
	"reflect"
	"sync"
)

/**
Proxy 好用、简单且高效的golang对象代理
*/

type ConstructFn[T interface{}] func(T)
type StringerFn func() string

type Proxy[T interface{}] struct {
	Value T
	tengo.ObjectImpl
	once        sync.Once
	Props       map[string]tengo.Object
	constructor ConstructFn[*Proxy[T]]
	typeName    string
	stringer    StringerFn
}

func (w *Proxy[T]) WithTypeName(typeName string) *Proxy[T] {
	w.typeName = typeName
	return w
}

func (w *Proxy[T]) WithConstructor(f ConstructFn[*Proxy[T]]) *Proxy[T] {
	w.constructor = f
	return w
}
func (w *Proxy[T]) WithValue(value T) *Proxy[T] {
	w.Value = value
	return w
}
func (w *Proxy[T]) WithStringer(fs StringerFn) *Proxy[T] {
	w.stringer = fs
	return w
}

func (w *Proxy[T]) IndexGet(key tengo.Object) (tengo.Object, error) {
	w.Init()
	k, _ := tengo.ToString(key)
	if v, ok := w.Props[k]; ok {
		return v, nil
	}
	return nil, fmt.Errorf("property/method [%s] not found", k)
}
func (w *Proxy[T]) IndexSet(key tengo.Object, value tengo.Object) error {
	w.Init()
	k, _ := tengo.ToString(key)
	if v, ok := w.Props[k]; ok {
		switch vv := v.(type) {
		case *tengo.UserFunction:
			if _, err := vv.Value(value); err != nil {
				return err
			}
		default:
			w.Props[k] = value
		}
	} else {
		w.Props[k] = value
	}
	return nil
}
func (w *Proxy[T]) Init() {
	w.once.Do(func() {
		if w.constructor != nil {
			w.constructor(w)
		}
	})
}
func (w *Proxy[T]) String() string {
	w.Init()
	if w.stringer != nil {
		return w.stringer()
	}
	v := interface{}(w.Value)
	if s, ok := v.(fmt.Stringer); ok {
		return s.String()
	}

	return ""
}
func (w *Proxy[T]) TypeName() string {
	w.Init()
	if w.typeName != "" {
		return w.typeName
	}
	t := reflect.TypeOf(w.Value)
	switch t.Kind() {
	case reflect.Pointer:
		return t.Elem().PkgPath() + "/" + t.Elem().Name()
	default:
		return t.PkgPath() + "/" + t.Name()
	}
}

func NewProxy[T interface{}](value T) *Proxy[T] {
	w := new(Proxy[T])
	w.WithValue(value)
	return w
}
