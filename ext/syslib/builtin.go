package syslib

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"github.com/d5/tengo/v2"
	log "github.com/sirupsen/logrus"
	"lightbox/ext/transpile"
	"lightbox/ext/util"
	"lightbox/sandbox"
)

func sysArgs(args ...tengo.Object) (ret tengo.Object, err error) {
	if len(args) == 0 {
		var argObjs []tengo.Object
		for _, arg := range flag.Args() {
			argObjs = append(argObjs, &tengo.String{Value: arg})
		}
		return &tengo.ImmutableArray{Value: argObjs}, nil
	}
	if len(args) == 1 {
		if idx, ok := tengo.ToInt(args[0]); ok {
			return tengo.FromInterface(flag.Arg(idx))
		}
	}
	return tengo.UndefinedValue, nil
}

func sysWait(args ...tengo.Object) (tengo.Object, error) {
	c := make(chan int, 1)
	<-c
	return nil, nil
}
func sysMust(args ...tengo.Object) (tengo.Object, error) {
	if len(args) != 1 {
		return nil, tengo.ErrWrongNumArguments
	}
	if v, ok := args[0].(*tengo.Error); ok {
		return nil, fmt.Errorf("%v", v)
	}
	return args[0], nil
}

func sysRequire(args ...tengo.Object) (tengo.Object, error) {
	if args[0] == tengo.UndefinedValue {
		if len(args) > 1 {
			return args[1], nil
		} else {
			return nil, errors.New("value is nil")
		}
	}
	if v, ok := args[0].(*tengo.Error); ok {
		return nil, fmt.Errorf(v.String())
	}
	return args[0], nil
}

var ErrUserExit = errors.New("exit")

func sysExit(args ...tengo.Object) (tengo.Object, error) {
	if len(args) == 1 {
		if err, ok := tengo.ToString(args[0]); ok {
			return nil, errors.New(err)
		}
	}
	return nil, ErrUserExit
}

func sysExec(app *sandbox.Applet, args ...tengo.Object) (tengo.Object, error) {
	if len(args) == 0 {
		return util.Error(tengo.ErrWrongNumArguments), nil
	}
	var values *tengo.ImmutableMap
	argLen := len(args)
	switch mapArg := args[len(args)-1].(type) {
	case *tengo.ImmutableMap:
		values = mapArg
		argLen = argLen - 1
	case *tengo.Map:
		values = &tengo.ImmutableMap{Value: mapArg.Value}
		argLen = argLen - 1
	default:
		values = &tengo.ImmutableMap{Value: map[string]tengo.Object{}}
	}
	placeHolder := util.PlaceHolders{}
	for k, v := range values.Value {
		placeHolder[k] = v
	}
	results := &tengo.ImmutableMap{Value: map[string]tengo.Object{}}
	for n := 0; n < argLen; n++ {
		script, ok := tengo.ToString(args[n])
		if !ok {
			log.Errorf("args: %s not a string", args[n])
			results.Value[script] = util.Error(fmt.Errorf("args: %s not a string", args[n]))
			continue
		}
		compiled, err := app.GetCompiled(script, placeHolder)
		if err != nil {
			log.Errorf("compile script %s error:%s", script, err)
			results.Value[script] = util.Error(err)
			continue
		}
		for k, v := range values.Value {
			err = compiled.Set(k, v)
			if err != nil {
				log.Errorf("set value %s error:%s", k, err)
				results.Value[script] = util.Error(err)
				continue
			}
		}
		err = compiled.Run()
		if err != nil {
			log.Errorf("run script %s error %s", script, err)
			results.Value[script] = util.Error(err)
		} else {
			results.Value[script] = tengo.TrueValue
		}
	}
	return results, nil
}

func sysConfig(app *sandbox.Applet, args ...tengo.Object) (tengo.Object, error) {
	cfg := app.Config()
	return tengo.FromInterface(cfg)
}
func sysGetConfig(app *sandbox.Applet, args ...tengo.Object) (tengo.Object, error) {
	if len(args) != 1 {
		return util.Error(tengo.ErrWrongNumArguments), nil
	}
	key, ok := tengo.ToString(args[0])
	if !ok {
		return tengo.UndefinedValue, nil
	}
	v, _ := app.GetConfig(key)
	return tengo.FromInterface(v)
}

func sysGetEnv(app *sandbox.Applet, args ...tengo.Object) (tengo.Object, error) {
	if len(args) != 1 {
		return util.Error(tengo.ErrWrongNumArguments), nil
	}
	key, ok := tengo.ToString(args[0])
	if !ok {
		return tengo.UndefinedValue, nil
	}
	v, _ := app.Context.Get(key)
	return tengo.FromInterface(v)
}

func sysFork(app *sandbox.Applet, args ...tengo.Object) (tengo.Object, error) {
	if len(args) == 0 {
		return util.Error(tengo.ErrWrongNumArguments), nil
	}
	var values *tengo.ImmutableMap
	argLen := len(args)
	switch mapArg := args[argLen-1].(type) {
	case *tengo.ImmutableMap:
		values = mapArg
		argLen = argLen - 1
	case *tengo.Map:
		values = &tengo.ImmutableMap{Value: mapArg.Value}
		argLen = argLen - 1
	default:
		values = &tengo.ImmutableMap{Value: map[string]tengo.Object{}}
	}
	placeHolder := util.PlaceHolders{}
	for k, v := range values.Value {
		placeHolder[k] = v
	}
	var cancelFuncs []context.CancelFunc
	for n := 0; n < argLen; n++ {
		script, ok := tengo.ToString(args[n])
		if !ok {
			log.Errorf("args: %s not a string", args[n])
			continue
		}
		compiled, err := app.GetCompiled(script, placeHolder)
		if err != nil {
			log.Errorf("compile script %s error:%s", script, err)
			continue
		}
		for k, v := range values.Value {
			err = compiled.Set(k, v)
			if err != nil {
				log.Errorf("set value %s error:%s", k, err)
				continue
			}
		}
		cancelCtx, cancelFunc := context.WithCancel(context.Background())
		cancelFuncs = append(cancelFuncs, cancelFunc)
		go func(c *tengo.Compiled) {
			err := c.RunContext(cancelCtx)
			if err != nil {
				log.Errorf("run script %s error %s", script, err)
			}
		}(compiled)

	}
	return &tengo.UserFunction{Value: func(args ...tengo.Object) (ret tengo.Object, err error) {
		for _, f := range cancelFuncs {
			f()
		}
		return
	}}, nil
}

func sysEnv(app *sandbox.Applet, args ...tengo.Object) (ret tengo.Object, err error) {
	if len(args) == 2 {
		k, ok := tengo.ToString(args[0])
		if !ok {
			return util.Error(tengo.ErrWrongNumArguments), nil
		}
		app.Context.Set(k, args[1])
		return args[1], nil
	}
	if len(args) == 1 {
		k, ok := tengo.ToString(args[0])
		if !ok {
			return util.Error(tengo.ErrWrongNumArguments), nil
		}
		if v, ok := app.Context.Get(k); ok {
			return tengo.FromInterface(v)
		} else {
			return tengo.UndefinedValue, nil
		}
	}
	if len(args) == 0 {
		envs := make(map[string]interface{})
		app.Context.Range(func(key, value any) bool {
			if kk, ok := key.(string); ok {
				switch vv := value.(type) {
				case string, int, int32, int64, uint, uint32, uint64:
					envs[kk] = vv
				default:
					if _, ok := vv.(tengo.Object); ok {
						envs[kk] = vv
					}
					//ignore
				}
			}
			return true
		})
		return tengo.FromInterface(envs)
	}
	return tengo.UndefinedValue, nil
}

func sysSet(app *sandbox.Applet, args ...tengo.Object) (tengo.Object, error) {
	if len(args) != 2 {
		return nil, tengo.ErrWrongNumArguments
	}
	key, _ := tengo.ToString(args[0])
	app.Context.Set(key, args[1])
	return args[1], nil
}

func sysAddTranspiler(app *sandbox.Applet, args ...tengo.Object) (tengo.Object, error) {
	for _, arg := range args {
		if name, ok := tengo.ToString(arg); ok {
			t, err := transpile.NewScriptReplaceFunc(app, name)
			if err != nil {
				return nil, err
			} else {
				app.WithTranspiler(t)
			}
		}
	}
	return nil, nil
}
