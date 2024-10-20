package util

import (
	"errors"
	"fmt"
	"github.com/d5/tengo/v2"
	"reflect"
)

type ReflectProxy struct {
	Self interface{}
	Name string
	tengo.ObjectImpl
	MethodCache map[string]tengo.UserFunction
}

func NewReflectProxy(self interface{}) *ReflectProxy {
	sw := &ReflectProxy{
		Self:        self,
		MethodCache: map[string]tengo.UserFunction{},
	}
	return sw
}

func (i *ReflectProxy) IndexGet(key tengo.Object) (tengo.Object, error) {
	mName, ok := tengo.ToString(key)
	if !ok {
		return nil, errors.New("method/field name must be a string")
	}
	self := i.Self
	if self == nil {
		self = interface{}(i)
	}
	v := reflect.ValueOf(self)
	casedName := BigCamelCase(mName)
	if f := v.Elem().FieldByNameFunc(func(name string) bool {
		return name == casedName || name == mName
	}); f.IsValid() {
		switch f.Kind() {
		case reflect.Uint, reflect.Uint32, reflect.Uint8, reflect.Uint16, reflect.Uint64:
			return &UInt{Value: f.Uint()}, nil
		case reflect.Ptr:
			return FromInterface(f.Interface())
		case reflect.Slice:
			return FromInterface(f.Interface())
		case reflect.Struct:
			return FromInterface(f.Addr().Interface())
		default:
			return tengo.FromInterface(f.Interface())
		}
	}

	m := v.MethodByName(casedName)
	if !m.IsValid() {
		m = v.MethodByName(mName)
	}
	if m.IsValid() {
		if callableFunc, ok := m.Interface().(tengo.CallableFunc); ok {
			return &tengo.UserFunction{
				Name:  mName,
				Value: callableFunc,
			}, nil
		}
		methodSignature := FuncSig(m.Interface())
		if mv, ok := FuncWrapper[methodSignature]; ok {
			vv := mv.Call([]reflect.Value{m})
			if m, ok := vv[0].Interface().(tengo.CallableFunc); ok {
				return &tengo.UserFunction{
					Value: m,
				}, nil
			}
		}
	}
	return nil, fmt.Errorf("method or field %s not exists", mName)
}
func (i *ReflectProxy) IndexSet(key tengo.Object, value tengo.Object) error {
	rawName, ok := tengo.ToString(key)
	if !ok {
		return errors.New("field name must be a string")
	}
	self := i.Self
	if self == nil {
		self = i
	}
	v := reflect.ValueOf(self)
	casedName := BigCamelCase(rawName)
	if f := v.Elem().FieldByNameFunc(func(name string) bool {
		return name == casedName || name == rawName
	}); f.IsValid() {
		return BindingField(f, value)
	}
	return fmt.Errorf("field %s not found", rawName)
}
func (i *ReflectProxy) TypeName() string {
	if i.Name != "" {
		return i.Name
	}
	if i != nil {
		//el := reflect.ValueOf(i.Self)
		//for el.Kind() == reflect.Ptr {
		//	el = el.Elem()
		//}
		//return el.Type().PkgPath() + "/" + el.Type().Code()
		return fmt.Sprintf("%T", i.Self)
	}
	return "undefined"
}
func (i *ReflectProxy) String() string {
	if i.Self != nil {
		if stringer, ok := i.Self.(fmt.Stringer); ok {
			return stringer.String()
		} else {
			return fmt.Sprintf("%#v", i.Self)
		}
	}
	return i.Name
}

func FromInterface(self interface{}) (tengo.Object, error) {
	if self == nil {
		return tengo.UndefinedValue, nil
	}
	var name = ""
	t := reflect.TypeOf(self)
	switch t.Kind() {
	case reflect.Bool:
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return &tengo.Int{Value: reflect.ValueOf(self).Int()}, nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return &UInt{Value: reflect.ValueOf(self).Uint()}, nil
	case reflect.Uintptr:
	case reflect.Float32, reflect.Float64:
	case reflect.Complex64, reflect.Complex128:
	case reflect.Array:
		name = "array-" + MiddleScoreCase(t.Elem().Name())
	case reflect.Chan:
		name = "chan-" + MiddleScoreCase(t.Elem().Name())
	case reflect.Func:
	case reflect.Interface:

	case reflect.Map:
		name = "map-" + MiddleScoreCase(t.Key().Name()) + "-" + MiddleScoreCase(t.Elem().Name())
	case reflect.Ptr:
		name = MiddleScoreCase(t.Elem().Name())
	case reflect.Slice:
		name = "array-" + MiddleScoreCase(t.Elem().String())
	case reflect.String:
	case reflect.Struct:
		name = MiddleScoreCase(t.Name())
	case reflect.UnsafePointer:
	}
	return WrapObject(self, name)
}

func WrapObject(self interface{}, typeName string) (tengo.Object, error) {

	if self == nil {
		return tengo.UndefinedValue, nil
	}
	v := reflect.ValueOf(self)
	switch v.Kind() {
	case reflect.Uint, reflect.Uint32, reflect.Uint8, reflect.Uint16, reflect.Uint64:
		return &UInt{Value: v.Uint()}, nil
	case reflect.Ptr:
		if v.IsNil() {
			return tengo.UndefinedValue, nil
		}
		if v.Elem().Kind() == reflect.Struct {
			if obj, ok := self.(tengo.Object); ok {
				return obj, nil
			} else {
				return &ReflectProxy{Self: self, Name: typeName}, nil
			}
		}
	case reflect.Map:
		kv := make(map[string]tengo.Object)
		iter := v.MapRange()
		for iter.Next() {
			vo, err := FromInterface(iter.Value().Interface())
			if err != nil {
				return nil, err
			}
			kv[iter.Key().String()] = vo
		}
		return &tengo.Map{Value: kv}, nil
	case reflect.Slice:
		switch v.Type().Elem().Kind() {
		case reflect.Uint8:
			return &tengo.Bytes{Value: v.Bytes()}, nil
		default:
			return &SliceProxy{Self: self, Name: typeName}, nil
		}

	}
	return tengo.FromInterface(self)
}

func NewUserFunc(f func(args ...tengo.Object) (tengo.Object, error)) *tengo.UserFunction {
	return &tengo.UserFunction{Value: f}
}
