package util

import (
	"encoding/json"
	"errors"
	"github.com/d5/tengo/v2"
	"reflect"
	"strings"
)

func StructFromArgs(args []tengo.Object, out interface{}) error {
	if len(args) == 0 {
		return tengo.ErrWrongNumArguments
	}
	return StructFromObject(args[0], out)
}

func UnmashalObject(object tengo.Object, out interface{}) error {
	v := tengo.ToInterface(object)
	buf, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return json.Unmarshal(buf, out)
}

//StructFromObject 简单绑定
func StructFromObject(obj tengo.Object, out interface{}) error {
	var values map[string]tengo.Object
	if m, ok := obj.(*tengo.ImmutableMap); ok {
		values = m.Value
	}
	if m, ok := obj.(*tengo.Map); ok {
		values = m.Value
	}
	if values == nil {
		return errors.New("arg[0] is not a map ")
	}
	return FromObjectMap(values, out)
}

func FromObjectMap(values map[string]tengo.Object, out interface{}) error {
	valueOf := reflect.ValueOf(out)
	if valueOf.Kind() == reflect.Ptr {
		valueOf = valueOf.Elem()
	}
	if valueOf.Kind() != reflect.Struct {
		return errors.New("binding not a struct")
	}
	for k, v := range values {
		if field := valueOf.FieldByNameFunc(func(n string) bool {
			return n == BigCamelCase(k) || strings.ToLower(n) == strings.Replace(strings.ToLower(k), "_", "", -1)
		}); field.IsValid() {
			err := BindingField(field, v)
			if err != nil {
				return err
			}
		}

	}
	return nil
}
func BindingArrayField(field reflect.Value, array *tengo.Array) error {
	switch field.Type().Elem().Kind() {
	case reflect.Int:
		var newValues []int
		for _, v := range array.Value {
			if i, ok := tengo.ToInt(v); ok {
				newValues = append(newValues, i)
			}
		}
		field.Set(reflect.ValueOf(newValues))
	case reflect.Int64:
		var newValues []int64
		for _, v := range array.Value {
			if i, ok := tengo.ToInt64(v); ok {
				newValues = append(newValues, i)
			}
		}
		field.Set(reflect.ValueOf(newValues))
	case reflect.String:
		var newValues []string
		for _, v := range array.Value {
			if i, ok := tengo.ToString(v); ok {
				newValues = append(newValues, i)
			}
		}
		field.Set(reflect.ValueOf(newValues))
	}
	return nil
}
func BindingField(field reflect.Value, v tengo.Object) error {
	switch field.Kind() {
	case reflect.Int8, reflect.Int16, reflect.Int, reflect.Int32, reflect.Int64:
		if i, ok := tengo.ToInt64(v); ok {
			field.SetInt(i)
		}
	case reflect.Uint8, reflect.Uint16, reflect.Uint, reflect.Uint32, reflect.Uint64:
		if i, ok := tengo.ToInt64(v); ok {
			field.SetUint(uint64(i))
		}
	case reflect.Bool:
		if b, ok := tengo.ToBool(v); ok {
			field.SetBool(b)
		}
	case reflect.Slice:
		switch vv := v.(type) {
		case *tengo.Bytes:
			field.SetBytes(vv.Value)
		case *tengo.Array:
			if err := BindingArrayField(field, vv); err != nil {
				return err
			}
		}
	case reflect.Ptr:

		if field.Type().Elem().Kind() == reflect.Struct {
			pt := reflect.New(field.Type().Elem())
			err := StructFromObject(v, pt.Interface())
			if err != nil {
				return err
			} else {
				field.Set(pt)
			}
		}
	case reflect.Struct:
		err := StructFromObject(v, field.Addr().Interface())
		if err != nil {
			return err

		}
	case reflect.Float32, reflect.Float64:
		if b, ok := tengo.ToFloat64(v); ok {
			field.SetFloat(b)
		}
	case reflect.String:
		if t, ok := tengo.ToString(v); ok && field.CanSet() {
			field.SetString(t)
		}
	}
	return nil
}
