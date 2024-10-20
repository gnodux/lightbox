package uuidlib

import (
	"github.com/d5/tengo/v2"
	uuid "github.com/satori/go.uuid"
	"lightbox/ext/util"
	"lightbox/sandbox"
)

var module = map[string]tengo.Object{
	"v1": &tengo.ImmutableMap{
		Value: map[string]tengo.Object{
			"new": &tengo.UserFunction{
				Value: func(args ...tengo.Object) (ret tengo.Object, err error) {
					return tengo.FromInterface(uuid.NewV1().String())
				},
			},
			"new_bytes": &tengo.UserFunction{
				Value: func(args ...tengo.Object) (ret tengo.Object, err error) {
					return tengo.FromInterface(uuid.NewV1().Bytes())
				},
			},
			"to_str": &tengo.UserFunction{
				Value: func(args ...tengo.Object) (ret tengo.Object, err error) {
					if len(args) != 1 {
						return nil, tengo.ErrWrongNumArguments
					}
					v := tengo.ToInterface(args[0])
					switch vv := v.(type) {
					case string:
						return args[0], nil
					case []byte:
						uuid := uuid.NewV1()
						if err = uuid.UnmarshalBinary([]byte(vv)); err != nil {
							return util.Error(err), nil
						}
					}
					return tengo.UndefinedValue, nil
				},
			},
			"to_bytes": &tengo.UserFunction{
				Value: func(args ...tengo.Object) (ret tengo.Object, err error) {
					if len(args) != 1 {
						return nil, tengo.ErrWrongNumArguments
					}
					v := tengo.ToInterface(args[0])
					switch vv := v.(type) {
					case string:
						uuid := uuid.NewV1()
						if err = uuid.UnmarshalText([]byte(vv)); err != nil {
							return util.Error(err), nil
						}
						return tengo.FromInterface(uuid.Bytes())
					case []byte:
						return args[0], nil
					}
					return tengo.UndefinedValue, nil
				},
			},
		},
	},
	"v4": &tengo.ImmutableMap{
		Value: map[string]tengo.Object{
			"new": &tengo.UserFunction{
				Value: func(args ...tengo.Object) (ret tengo.Object, err error) {
					return tengo.FromInterface(uuid.NewV4().String())
				},
			},
			"new_bytes": &tengo.UserFunction{
				Value: func(args ...tengo.Object) (ret tengo.Object, err error) {
					return tengo.FromInterface(uuid.NewV4().Bytes())
				},
			},
			"to_str": &tengo.UserFunction{
				Value: func(args ...tengo.Object) (ret tengo.Object, err error) {
					if len(args) != 1 {
						return nil, tengo.ErrWrongNumArguments
					}
					v := tengo.ToInterface(args[0])
					switch vv := v.(type) {
					case string:
						return args[0], nil
					case []byte:
						uuid := uuid.NewV4()
						if err = uuid.UnmarshalBinary([]byte(vv)); err != nil {
							return util.Error(err), nil
						}
					}
					return tengo.UndefinedValue, nil
				},
			},
			"to_bytes": &tengo.UserFunction{
				Value: func(args ...tengo.Object) (ret tengo.Object, err error) {
					if len(args) != 1 {
						return nil, tengo.ErrWrongNumArguments
					}
					v := tengo.ToInterface(args[0])
					switch vv := v.(type) {
					case string:
						uuid := uuid.NewV4()
						if err = uuid.UnmarshalText([]byte(vv)); err != nil {
							return util.Error(err), nil
						}
						return tengo.FromInterface(uuid.Bytes())
					case []byte:
						return args[0], nil
					}
					return tengo.UndefinedValue, nil
				},
			},
		},
	},
}

var Entry = sandbox.NewRegistry("uuid", module, nil)
