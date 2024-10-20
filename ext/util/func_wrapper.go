package util

import (
	"errors"
	"github.com/d5/tengo/v2"
	"github.com/d5/tengo/v2/stdlib"
	"reflect"
)

var FuncWrapper = map[string]reflect.Value{
	"FuncASSRE":     reflect.ValueOf(stdlib.FuncASSRE),
	"FuncASRE":      reflect.ValueOf(stdlib.FuncASRE),
	"FuncARE":       reflect.ValueOf(stdlib.FuncARE),
	"FuncAR":        reflect.ValueOf(stdlib.FuncAR),
	"FuncAIRS":      reflect.ValueOf(stdlib.FuncAIRS),
	"FuncASRS":      reflect.ValueOf(stdlib.FuncASRS),
	"FuncAFFRF":     reflect.ValueOf(stdlib.FuncAFFRF),
	"FuncAFRB":      reflect.ValueOf(stdlib.FuncAFRB),
	"FuncAFIRB":     reflect.ValueOf(stdlib.FuncAFIRB),
	"FuncAFIRF":     reflect.ValueOf(stdlib.FuncAFIRF),
	"FuncAFRI":      reflect.ValueOf(stdlib.FuncAFRI),
	"FuncASRI64E":   reflect.ValueOf(FuncASRI64E),
	"FuncASSRI64E":  reflect.ValueOf(FuncASSRI64E),
	"FuncASSSRI64E": reflect.ValueOf(FuncASSSRI64E),
	"FuncAI64SRE":   reflect.ValueOf(FuncAI64SRE),
	"FuncASIs":      reflect.ValueOf(FuncASIs),
	"FuncASsRS":     reflect.ValueOf(FuncASsRS),
	"FuncAIs":       reflect.ValueOf(FuncAIs),
	"FuncASRB":      reflect.ValueOf(FuncASRB),
	"FuncASRBE":     reflect.ValueOf(FuncASRBE),
	"FuncABRE":      reflect.ValueOf(FuncABRE),
}

func Error(err error) tengo.Object {
	if err == nil {
		return tengo.TrueValue
	}
	return &tengo.Error{Value: &tengo.String{Value: err.Error()}}
}

func getParameterType(t reflect.Type) string {
	if t.Implements(reflect.TypeOf((*error)(nil)).Elem()) {
		return "E"
	}
	switch t.Kind() {
	case reflect.String:
		return "S"
	case reflect.Int:
		return "I"
	case reflect.Int64:
		return "I64"
	case reflect.Bool:
		return "B"
	case reflect.Float64:
		return "F"
	case reflect.Ptr:
		return getParameterType(t.Elem())
	case reflect.Interface:
		return "I"
	case reflect.Slice:
		return getParameterType(t.Elem()) + "s"
	}
	return ""
}

func FuncSig(fn interface{}) string {
	sig := "FuncA"
	t := reflect.TypeOf(fn)
	for i := 0; i < t.NumIn(); i++ {
		argT := t.In(i)
		sig = sig + getParameterType(argT)
	}
	sig += "R"
	for i := 0; i < t.NumOut(); i++ {
		argt := t.Out(i)
		sig = sig + getParameterType(argt)
	}
	return sig
}

func FuncASRI64E(fn func(s string) (int64, error)) tengo.CallableFunc {
	return func(args ...tengo.Object) (tengo.Object, error) {
		if len(args) < 1 {
			return nil, tengo.ErrWrongNumArguments
		}
		arg1, ok := tengo.ToString(args[0])
		if !ok {
			return nil, errors.New("first argument must be string")
		}
		result, err := fn(arg1)
		if err != nil {
			return Error(err), nil
		} else {
			return tengo.FromInterface(result)
		}
	}
}
func FuncASSRI64E(fn func(string, string) (int64, error)) tengo.CallableFunc {
	return func(args ...tengo.Object) (ret tengo.Object, err error) {
		if len(args) != 2 {
			return nil, tengo.ErrWrongNumArguments
		}
		s1, ok := tengo.ToString(args[0])
		if !ok {
			return nil, errors.New("1st arg must be a string")
		}
		s2, ok := tengo.ToString(args[1])
		if !ok {
			return nil, errors.New("2nd arg must be a string")
		}
		result, err := fn(s1, s2)
		if err != nil {
			return Error(err), nil
		} else {
			return tengo.FromInterface(result)
		}
	}
}
func FuncASSSRI64E(fn func(string, string, string) (int64, error)) tengo.CallableFunc {
	return func(args ...tengo.Object) (ret tengo.Object, err error) {
		if len(args) != 2 {
			return nil, tengo.ErrWrongNumArguments
		}
		s1, ok := tengo.ToString(args[0])
		if !ok {
			return nil, errors.New("1st arg must be a string")
		}
		s2, ok := tengo.ToString(args[1])
		if !ok {
			return nil, errors.New("2nd arg must be a string")
		}
		s3, ok := tengo.ToString(args[1])
		if !ok {
			return nil, errors.New("3rd arg must be a string")
		}
		result, err := fn(s1, s2, s3)
		if err != nil {
			return Error(err), nil
		} else {
			return tengo.FromInterface(result)
		}
	}
}

func FuncASIs(fn func(string, ...interface{})) tengo.CallableFunc {
	return func(args ...tengo.Object) (tengo.Object, error) {
		if len(args) < 1 {
			return nil, tengo.ErrWrongNumArguments
		}
		format, ok := tengo.ToString(args[0])
		if !ok {
			return nil, errors.New("first argument must be string")
		}
		var (
			printArgs []interface{}
			err       error
		)
		if len(args) > 1 {
			printArgs, err = getPrintArgs(args[1:]...)
			if err != nil {
				return nil, err
			}
			fn(format, printArgs...)
		} else {
			fn(format)
		}
		return nil, nil
	}
}

func FuncAI64SRE(fn func(int64, string) error) tengo.CallableFunc {
	return func(args ...tengo.Object) (tengo.Object, error) {
		if len(args) != 2 {
			return nil, tengo.ErrWrongNumArguments
		}
		p1, ok := tengo.ToInt64(args[0])
		if !ok {
			return nil, errors.New("first argument must be number")
		}
		p2, ok := tengo.ToString(args[1])
		if !ok {
			return nil, errors.New("second argument must be string")
		}
		err := fn(p1, p2)
		return Error(err), nil
	}
}
func FuncAIs(fn func(...interface{})) tengo.CallableFunc {
	return func(args ...tengo.Object) (ret tengo.Object, err error) {
		printArgs, err := getPrintArgs(args...)
		if err != nil {
			return nil, err
		}
		fn(printArgs...)
		return nil, nil
	}
}
func FuncASROE(fn func(string) (tengo.Object, error)) tengo.CallableFunc {
	return func(args ...tengo.Object) (ret tengo.Object, err error) {
		if len(args) != 1 {
			return Error(tengo.ErrWrongNumArguments), nil
		}
		s, ok := tengo.ToString(args[0])
		if !ok {
			return Error(&tengo.ErrInvalidArgumentType{Name: "arg0", Expected: "string", Found: args[0].TypeName()}), nil
		}
		r, err := fn(s)
		if err != nil {
			return Error(err), nil
		}
		return r, nil
	}
}
func FuncARO(fn func() tengo.Object) tengo.CallableFunc {
	return func(args ...tengo.Object) (ret tengo.Object, err error) {
		return fn(), nil
	}
}

func getPrintArgs(args ...tengo.Object) ([]interface{}, error) {
	var printArgs []interface{}
	l := 0
	for _, arg := range args {
		s, _ := tengo.ToString(arg)
		slen := len(s)
		// wrap sure length does not exceed the limit
		if l+slen > tengo.MaxStringLen {
			return nil, tengo.ErrStringLimit
		}
		l += slen
		printArgs = append(printArgs, s)
	}
	return printArgs, nil
}
