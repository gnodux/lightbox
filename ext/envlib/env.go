package envlib

import (
	"errors"
	"github.com/d5/tengo/v2"
	"lightbox/env"
	"lightbox/ext/util"
	"lightbox/sandbox"
)

func get(app *sandbox.Applet, args ...tengo.Object) (ret tengo.Object, err error) {
	if len(args) != 1 {
		return nil, tengo.ErrWrongNumArguments
	}
	k, ok := tengo.ToString(args[0])
	if !ok {
		return nil, errors.New("arg must be a string")
	}
	if v, ok := app.Context.Get(k); ok {
		ret, err = tengo.FromInterface(v)
	} else {
		ret = tengo.UndefinedValue
	}
	return
}

func set(app *sandbox.Applet, args ...tengo.Object) (ret tengo.Object, err error) {
	if len(args) != 2 {
		return nil, tengo.ErrWrongNumArguments
	}
	key, ok := tengo.ToString(args[0])
	if !ok {
		return nil, errors.New("key must be a string")
	}
	app.Context.Set(key, args[1])
	return
}
func global(args ...tengo.Object) (ret tengo.Object, err error) {
	if len(args) != 1 {
		return nil, tengo.ErrWrongNumArguments
	}

	k, ok := tengo.ToString(args[0])
	if !ok {
		return nil, errors.New("arg must be a string")
	}
	if v, ok := env.Get[any](k); ok {
		ret, err = tengo.FromInterface(v)
	} else {
		ret = tengo.UndefinedValue
	}
	return
}
func profile(args ...tengo.Object) (ret tengo.Object, err error) {
	s, _ := env.Get[string](env.Profile)
	if s == "" {
		s = "test"
	}
	return &tengo.String{Value: s}, nil

}

var module = map[string]tengo.Object{
	"global":  util.NewUserFunc(global),
	"profile": util.NewUserFunc(profile),
}

var appModule = map[string]sandbox.UserFunction{
	"get": get,
	"set": set,
}

var Entry = sandbox.NewRegistry("env", module, appModule)
