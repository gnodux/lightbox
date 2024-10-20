package badgerlib

import (
	"fmt"
	"github.com/d5/tengo/v2"
	"lightbox/ext/util"
	"lightbox/kvstore"
	"lightbox/sandbox"
)

//todo: lock directory to applet base

var module = map[string]tengo.Object{

	//snippet:name=badger.take(),prefix=take;body=take($1);desc=take a badger db,usage=badger.take("db_name")
	"take": &tengo.UserFunction{
		Name: "take",
		Value: func(args ...tengo.Object) (ret tengo.Object, err error) {
			if len(args) != 1 {
				return util.Error(tengo.ErrWrongNumArguments), nil
			}
			if arg, ok := tengo.ToString(args[0]); ok {
				if db, err := kvstore.Get(arg); err == nil {
					return wrapBadgerClient(db, arg), nil
				}
			}
			return tengo.FromInterface(fmt.Errorf("badger %v not found", args[0]))
		},
	},
	//snippet:name=badger.open();prefix=open;body=open();desc=get default badger database;
	//snippet:name=badger.open(name);prefix=open;body=open($1);desc=open a exists badger database,if badger not exists,use name as dir and create a badger client;
	//snippet:name=badger.open;prefix=open;body=open(${1:name},${2:dir});desc=open a badger database;
	//snippet:name=badger.open;prefix=open;body=open(${1:name},${2:option});desc=open a badger database;
	"open": &tengo.UserFunction{Name: "open", Value: openBadger},
	//snippet:name=badger.option;prefix=option;body=option(${1:path});desc=create a badger option;
	"option": &tengo.UserFunction{Name: "option", Value: func(args ...tengo.Object) (ret tengo.Object, err error) {
		if len(args) == 0 {
			opt := kvstore.DefaultOptions("./kv-store")
			return util.NewReflectProxy(&opt), nil
		}
		if len(args) != 1 {
			return util.Error(tengo.ErrWrongNumArguments), nil
		}
		p, ok := tengo.ToString(args[0])
		if !ok {
			return util.Error(&tengo.ErrInvalidArgumentType{Name: "path", Expected: "string", Found: args[0].TypeName()}), nil
		}
		opt := kvstore.DefaultOptions(p)
		return util.NewReflectProxy(&opt), nil
	}},
	"mini_size": &tengo.UserFunction{
		Name: "mini_size",
		Value: func(args ...tengo.Object) (ret tengo.Object, err error) {
			kvstore.SetMiniSize()
			return nil, nil
		},
	},
	"small_size": &tengo.UserFunction{
		Name: "small_size",
		Value: func(args ...tengo.Object) (ret tengo.Object, err error) {
			kvstore.SetSmallSize()
			return nil, nil
		}},
	"large_size": &tengo.UserFunction{Name: "large_size",
		Value: func(args ...tengo.Object) (ret tengo.Object, err error) {
			kvstore.SetLargeSize()
			return nil, nil
		},
	},
}

var Entry = sandbox.NewRegistry("badger", module, nil)
