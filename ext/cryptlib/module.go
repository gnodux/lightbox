package cryptlib

import (
	"github.com/d5/tengo/v2"
	"github.com/golang-jwt/jwt"
	"lightbox/ext/util"
	"lightbox/sandbox"
)

var module = map[string]tengo.Object{
	"md5": &tengo.ImmutableMap{
		Value: map[string]tengo.Object{
			"sum": &tengo.UserFunction{Name: "sum",
				Value: sumMD5,
			},
		}},
	"jwt": &tengo.ImmutableMap{
		Value: map[string]tengo.Object{
			"parse": &tengo.UserFunction{
				Value: func(args ...tengo.Object) (tengo.Object, error) {
					var (
						tokenString string
						key         string
					)
					for idx, arg := range args {
						switch idx {
						case 0:
							tokenString, _ = tengo.ToString(arg)
						case 1:
							key, _ = tengo.ToString(arg)
						}
					}
					claims, err := JWTParseWithBase64Key(tokenString, key)
					if err != nil {
						return util.Error(err), nil
					} else {
						return tengo.FromInterface(claims)
					}
				},
			},
			"sign": &tengo.UserFunction{
				Value: func(args ...tengo.Object) (tengo.Object, error) {
					var (
						claims map[string]interface{}
						key    string
						method string
					)
					for idx, arg := range args {
						switch idx {
						case 0:
							switch m := arg.(type) {
							case *tengo.Map:
								claims = util.ToMap[interface{}](m.Value)
							case *tengo.ImmutableMap:
								claims = util.ToMap[interface{}](m.Value)
							}
						case 1:
							key, _ = tengo.ToString(arg)
						case 2:
							method, _ = tengo.ToString(arg)
						}
					}
					tokenStr, err := JWTSigning(method, key, jwt.MapClaims(claims))
					if err != nil {
						return util.Error(err), nil
					} else {
						return tengo.FromInterface(tokenStr)
					}
				},
			},
		},
	},
	"hash_id": &tengo.ImmutableMap{Value: map[string]tengo.Object{
		"encode_int32": &tengo.UserFunction{Value: func(args ...tengo.Object) (tengo.Object, error) {
			if len(args) != 2 {
				return nil, tengo.ErrWrongNumArguments
			}
			salt, _ := tengo.ToString(args[0])
			var values []int64
			switch v := args[1].(type) {
			case *tengo.Array:
				values = util.ToSlice[int64](v.Value)
			case *tengo.ImmutableArray:
				values = util.ToSlice[int64](v.Value)
			}
			var intValues = make([]int, len(values))
			for idx, v := range values {
				intValues[idx] = int(v)
			}
			s, err := hashIDEncode(salt, intValues)
			if err != nil {
				return util.Error(err), nil
			}
			return &tengo.String{Value: s}, nil
		}},
		"decode_int32": &tengo.UserFunction{Value: func(args ...tengo.Object) (tengo.Object, error) {
			if len(args) != 2 {
				return nil, tengo.ErrWrongNumArguments
			}
			salt, _ := tengo.ToString(args[0])
			hash, _ := tengo.ToString(args[1])
			v, err := decodeInt(salt, hash)
			if err != nil {
				return util.Error(err), nil
			}
			return util.ToImmutableArray(v...)
		}},
		"encode_int64": &tengo.UserFunction{Value: func(args ...tengo.Object) (tengo.Object, error) {
			if len(args) != 2 {
				return nil, tengo.ErrWrongNumArguments
			}
			salt, _ := tengo.ToString(args[0])
			var values []int64
			switch v := args[1].(type) {
			case *tengo.Array:
				values = util.ToSlice[int64](v.Value)
			case *tengo.ImmutableArray:
				values = util.ToSlice[int64](v.Value)
			}
			s, err := hashIDEncode(salt, values)
			if err != nil {
				return util.Error(err), nil
			}
			return &tengo.String{Value: s}, nil
		}},
		"decode_int64": &tengo.UserFunction{Value: func(args ...tengo.Object) (tengo.Object, error) {
			if len(args) != 2 {
				return nil, tengo.ErrWrongNumArguments
			}
			salt, _ := tengo.ToString(args[0])
			hash, _ := tengo.ToString(args[1])
			v, err := decodeInt64(salt, hash)
			if err != nil {
				return util.Error(err), nil
			}
			return util.ToImmutableArray(v...)
		}},
		"encode_hex": &tengo.UserFunction{Value: func(args ...tengo.Object) (tengo.Object, error) {
			if len(args) != 2 {
				return nil, tengo.ErrWrongNumArguments
			}
			salt, _ := tengo.ToString(args[0])
			values, _ := tengo.ToString(args[1])
			s, err := hashIDEncode(salt, values)
			if err != nil {
				return util.Error(err), nil
			}
			return &tengo.String{Value: s}, nil
		}},
		"decode_hex": &tengo.UserFunction{Value: func(args ...tengo.Object) (tengo.Object, error) {
			if len(args) != 2 {
				return nil, tengo.ErrWrongNumArguments
			}
			salt, _ := tengo.ToString(args[0])
			values, _ := tengo.ToString(args[1])
			s, err := decodeHex(salt, values)
			if err != nil {
				return util.Error(err), nil
			}
			return &tengo.String{Value: s}, nil
		}},
	}},
	"aes": &tengo.ImmutableMap{Value: map[string]tengo.Object{
		"encode": &tengo.UserFunction{
			Name: "encode",
			Value: func(args ...tengo.Object) (tengo.Object, error) {
				if len(args) != 2 {
					return util.Error(tengo.ErrWrongNumArguments), nil
				}
				var (
					input []byte
					key   []byte
					ok    bool
				)
				input, ok = tengo.ToByteSlice(args[0])
				if !ok {
					return util.Error(&tengo.ErrInvalidArgumentType{
						Name:     "data",
						Expected: "string/[]byte",
						Found:    args[0].TypeName(),
					}), nil
				}
				key, ok = tengo.ToByteSlice(args[1])
				if !ok {
					return util.Error(&tengo.ErrInvalidArgumentType{
						Name:     "key",
						Expected: "string/[]byte",
						Found:    args[0].TypeName(),
					}), nil
				}
				result, err := AesEncrypt(input, key)
				if err != nil {
					return util.Error(err), nil
				}
				return &tengo.Bytes{Value: result}, nil
			},
		},
		"decode": &tengo.UserFunction{
			Name: "decode",
			Value: func(args ...tengo.Object) (tengo.Object, error) {
				if len(args) != 2 {
					return util.Error(tengo.ErrWrongNumArguments), nil
				}
				var (
					input []byte
					key   []byte
					ok    bool
				)
				input, ok = tengo.ToByteSlice(args[0])
				if !ok {
					return util.Error(&tengo.ErrInvalidArgumentType{
						Name:     "data",
						Expected: "string/[]byte",
						Found:    args[0].TypeName(),
					}), nil
				}
				key, ok = tengo.ToByteSlice(args[1])
				if !ok {
					return util.Error(&tengo.ErrInvalidArgumentType{
						Name:     "key",
						Expected: "string/[]byte",
						Found:    args[0].TypeName(),
					}), nil
				}
				result, err := AesDecrypt(input, key)
				if err != nil {
					return util.Error(err), nil
				}
				return &tengo.Bytes{Value: result}, nil
			},
		},
	}},
	"ras": &tengo.ImmutableMap{Value: map[string]tengo.Object{
		"encode": &tengo.UserFunction{
			Name: "encode",
			Value: func(args ...tengo.Object) (tengo.Object, error) {
				if len(args) != 2 {
					return util.Error(tengo.ErrWrongNumArguments), nil
				}
				var (
					input []byte
					key   []byte
					ok    bool
				)
				input, ok = tengo.ToByteSlice(args[0])
				if !ok {
					return util.Error(&tengo.ErrInvalidArgumentType{
						Name:     "data",
						Expected: "string/[]byte",
						Found:    args[0].TypeName(),
					}), nil
				}
				key, ok = tengo.ToByteSlice(args[1])
				if !ok {
					return util.Error(&tengo.ErrInvalidArgumentType{
						Name:     "key",
						Expected: "string/[]byte",
						Found:    args[0].TypeName(),
					}), nil
				}
				result, err := RsaEncrypt(input, key)
				if err != nil {
					return util.Error(err), nil
				}
				return &tengo.Bytes{Value: result}, nil
			},
		},
		"decode": &tengo.UserFunction{
			Name: "decode",
			Value: func(args ...tengo.Object) (tengo.Object, error) {
				if len(args) != 2 {
					return util.Error(tengo.ErrWrongNumArguments), nil
				}
				var (
					input []byte
					key   []byte
					ok    bool
				)
				input, ok = tengo.ToByteSlice(args[0])
				if !ok {
					return util.Error(&tengo.ErrInvalidArgumentType{
						Name:     "data",
						Expected: "string/[]byte",
						Found:    args[0].TypeName(),
					}), nil
				}
				key, ok = tengo.ToByteSlice(args[1])
				if !ok {
					return util.Error(&tengo.ErrInvalidArgumentType{
						Name:     "key",
						Expected: "string/[]byte",
						Found:    args[0].TypeName(),
					}), nil
				}
				result, err := RsaDecrypt(input, key)
				if err != nil {
					return util.Error(err), nil
				}
				return &tengo.Bytes{Value: result}, nil
			},
		},
	}},
}

var Entry = sandbox.NewRegistry("crypt", module, nil)
