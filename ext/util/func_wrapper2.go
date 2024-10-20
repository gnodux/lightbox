package util

import (
	"errors"
	"fmt"
	"github.com/d5/tengo/v2"
	log "github.com/sirupsen/logrus"
)

func FuncASRBE(f func(pattern string, name string) (matched bool, err error)) tengo.CallableFunc {
	return func(args ...tengo.Object) (tengo.Object, error) {
		if len(args) != 2 {
			return nil, tengo.ErrWrongNumArguments
		}
		var (
			name    string
			pattern string
			ok      bool
		)
		if pattern, ok = tengo.ToString(args[0]); !ok {
			return nil, errors.New("first argument must be string")
		}
		if name, ok = tengo.ToString(args[1]); !ok {
			return nil, errors.New("second argument must be string")
		}
		if match, err := f(pattern, name); err != nil {
			return nil, err
		} else {
			if match {
				return tengo.TrueValue, nil
			} else {
				return tengo.FalseValue, nil
			}
		}
	}
}

func FuncABRE(f func(v bool) error) tengo.CallableFunc {
	return func(args ...tengo.Object) (tengo.Object, error) {
		b, _ := tengo.ToBool(args[0])
		err := f(b)
		if err != nil {
			return Error(err), nil
		} else {
			return nil, nil
		}
	}
}

func FuncASRB(fn func(string) bool) tengo.CallableFunc {
	return func(args ...tengo.Object) (tengo.Object, error) {
		if len(args) != 1 {
			return nil, tengo.ErrWrongNumArguments
		}
		s, ok := tengo.ToString(args[0])
		if !ok {
			return nil, errors.New("first argument must be string")
		}
		if fn(s) {
			return tengo.TrueValue, nil
		}
		return tengo.FalseValue, nil
	}
}

func FuncASsRS(fn func(elem ...string) string) tengo.CallableFunc {
	return func(args ...tengo.Object) (tengo.Object, error) {
		if len(args) == 0 {
			return nil, tengo.ErrWrongNumArguments
		}
		var ss []string
		for _, arg := range args {
			s, ok := tengo.ToString(arg)
			if !ok {
				return nil, errors.New("invalidate argument path in join")
			}
			ss = append(ss, s)
		}
		result := fn(ss...)
		return &tengo.String{Value: result}, nil
	}
}

func FuncASRSS(fn func(string) (string, string)) tengo.CallableFunc {
	return func(args ...tengo.Object) (tengo.Object, error) {
		if len(args) != 1 {
			return nil, tengo.ErrWrongNumArguments
		}
		s1, ok := tengo.ToString(args[0])
		if !ok {
			return nil, tengo.ErrInvalidArgumentType{
				Name:     "first",
				Expected: "string(compatible)",
				Found:    args[0].TypeName(),
			}
		}
		s2, s3 := fn(s1)
		if len(s2) > tengo.MaxStringLen || len(s3) > tengo.MaxStringLen {
			return nil, tengo.ErrStringLimit
		}
		return &tengo.Array{Value: []tengo.Object{
			&tengo.String{Value: s2}, &tengo.String{Value: s3},
		}}, nil
	}
}
func FuncASsRE(fn func([]string) error) tengo.CallableFunc {
	return func(args ...tengo.Object) (tengo.Object, error) {
		var ss []string
		for idx, arg := range args {
			s, ok := tengo.ToString(arg)
			if !ok {
				return nil, fmt.Errorf("arg [%d] not a string", idx)
			}
			ss = append(ss, s)
		}
		err := fn(ss)
		return nil, err
	}
}

func CallWith(f func() error, fe func(error) bool) error {
	err := f()
	if err != nil {
		if fe(err) {
			return nil
		} else {
			return err
		}
	}
	return nil
}
func CallWithIgnoreError(f func() error, ignoreError bool) error {
	return CallWith(f, func(err error) bool {
		log.Errorf("error:%s", err)
		return ignoreError
	})
}
