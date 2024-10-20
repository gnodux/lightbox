package util

import (
	"errors"
	"fmt"
	"github.com/d5/tengo/v2"
	"reflect"
)

type IntSliceIterator struct {
	tengo.ObjectImpl
	wrapper *SliceProxy
	current int
	l       int
}

func (i *IntSliceIterator) Next() bool {
	i.current = i.current + 1
	return i.current < i.l
}

func (i *IntSliceIterator) Key() tengo.Object {
	return &tengo.Int{Value: int64(i.current)}
}

func (i *IntSliceIterator) Value() tengo.Object {
	r, _ := i.wrapper.IndexGet(&tengo.Int{Value: int64(i.current)})
	return r
}
func (i *IntSliceIterator) TypeName() string {
	return "IntSliceIterator"
}
func (i *IntSliceIterator) String() string {
	return "<int-slice>"
}

type SliceProxy struct {
	Self interface{}
	Name string
	tengo.ObjectImpl
}

func (s *SliceProxy) TypeName() string {
	return s.Name
}
func (s *SliceProxy) String() string {
	return fmt.Sprintf("%v", s.Self)
}

func (s *SliceProxy) IndexGet(key tengo.Object) (tengo.Object, error) {
	idx, ok := tengo.ToInt(key)
	if !ok {
		return nil, errors.New("index must be a number")
	}
	value := reflect.ValueOf(s.Self).Index(idx)
	if value.IsValid() {
		switch value.Kind() {
		case reflect.Ptr:
			return WrapObject(value.Interface(), value.Type().Name())
		case reflect.Struct:
			return WrapObject(value.Addr().Interface(), value.Type().Name())
		default:
			return tengo.FromInterface(value.Interface())
		}
	}
	return tengo.UndefinedValue, nil
}

// Iterate returns an iterator.
func (s *SliceProxy) Iterate() tengo.Iterator {
	return &IntSliceIterator{wrapper: s, l: s.Length(), current: -1}
}

func (s *SliceProxy) Length() int {
	return reflect.ValueOf(s.Self).Len()
}

// CanIterate returns whether the Object can be Iterated.
func (s *SliceProxy) CanIterate() bool {
	return true
}
