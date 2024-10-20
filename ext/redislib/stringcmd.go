package redislib

import (
	"github.com/d5/tengo/v2"
	"github.com/d5/tengo/v2/stdlib"
	"github.com/go-redis/redis/v8"
	"lightbox/ext/util"
)

//type StringCmdWrapper struct {
//	util.Proxy[*redis.StringCmd]
//}
//
//func (s *StringCmdWrapper) Init() map[string]tengo.Object {
//	return map[string]tengo.Object{
//		"int": util.NewUserFunc(func(args ...tengo.Object) (tengo.Object, error) {
//			v, err := s.Value.Int()
//			if err != nil {
//				return util.Error(err), nil
//			}
//			return &tengo.Int{Value: int64(v)}, nil
//		}),
//		"int64": util.NewUserFunc(func(args ...tengo.Object) (tengo.Object, error) {
//			v, err := s.Value.Int64()
//			if err != nil {
//				return util.Error(err), nil
//			}
//			return &tengo.Int{Value: v}, nil
//		}),
//		"uint64": util.NewUserFunc(func(args ...tengo.Object) (tengo.Object, error) {
//			v, err := s.Value.Uint64()
//			if err != nil {
//				return util.Error(err), nil
//			}
//			return &util.UInt{Value: v}, nil
//		}),
//		"float": util.NewUserFunc(func(args ...tengo.Object) (tengo.Object, error) {
//			v, err := s.Value.Float32()
//			if err != nil {
//				return util.Error(err), nil
//			}
//			return &tengo.Float{Value: float64(v)}, nil
//		}),
//		"float64": util.NewUserFunc(func(args ...tengo.Object) (tengo.Object, error) {
//			v, err := s.Value.Float64()
//			if err != nil {
//				return util.Error(err), nil
//			}
//			return &tengo.Float{Value: v}, nil
//		}),
//		"bytes": util.NewUserFunc(func(args ...tengo.Object) (tengo.Object, error) {
//			v, err := s.Value.Bytes()
//			if err != nil {
//				return util.Error(err), nil
//			}
//			return &tengo.Bytes{Value: v}, nil
//		}),
//		"time": util.NewUserFunc(func(args ...tengo.Object) (tengo.Object, error) {
//			v, err := s.Value.Time()
//			if err != nil {
//				return util.Error(err), nil
//			}
//			return &tengo.Time{Value: v}, nil
//		}),
//		"bool": util.NewUserFunc(func(args ...tengo.Object) (tengo.Object, error) {
//			v, err := s.Value.Bool()
//			if err != nil {
//				return util.Error(err), nil
//			}
//			return tengo.FromInterface(v)
//		}),
//		"err":       util.NewUserFunc(stdlib.FuncARE(s.Value.Err)),
//		"val":       util.NewUserFunc(stdlib.FuncARS(s.Value.Val)),
//		"name":      util.NewUserFunc(stdlib.FuncARS(s.Value.Code)),
//		"full_name": util.NewUserFunc(stdlib.FuncARS(s.Value.FullName)),
//	}
//}
//
//func (s *StringCmdWrapper) TypeName() string {
//	return "string-command"
//}
//func (s *StringCmdWrapper) String() string {
//	return s.Value.String()
//}

func stringCmdConstructor(proxy *util.Proxy[*redis.StringCmd]) {
	cmd := proxy.Value
	proxy.Props = map[string]tengo.Object{
		"int": util.NewUserFunc(func(args ...tengo.Object) (tengo.Object, error) {
			v, err := cmd.Int()
			if err != nil {
				return util.Error(err), nil
			}
			return &tengo.Int{Value: int64(v)}, nil
		}),
		"int64": util.NewUserFunc(func(args ...tengo.Object) (tengo.Object, error) {
			v, err := cmd.Int64()
			if err != nil {
				return util.Error(err), nil
			}
			return &tengo.Int{Value: v}, nil
		}),
		"uint64": util.NewUserFunc(func(args ...tengo.Object) (tengo.Object, error) {
			v, err := cmd.Uint64()
			if err != nil {
				return util.Error(err), nil
			}
			return &util.UInt{Value: v}, nil
		}),
		"float": util.NewUserFunc(func(args ...tengo.Object) (tengo.Object, error) {
			v, err := cmd.Float32()
			if err != nil {
				return util.Error(err), nil
			}
			return &tengo.Float{Value: float64(v)}, nil
		}),
		"float64": util.NewUserFunc(func(args ...tengo.Object) (tengo.Object, error) {
			v, err := cmd.Float64()
			if err != nil {
				return util.Error(err), nil
			}
			return &tengo.Float{Value: v}, nil
		}),
		"bytes": util.NewUserFunc(func(args ...tengo.Object) (tengo.Object, error) {
			v, err := cmd.Bytes()
			if err != nil {
				return util.Error(err), nil
			}
			return &tengo.Bytes{Value: v}, nil
		}),
		"time": util.NewUserFunc(func(args ...tengo.Object) (tengo.Object, error) {
			v, err := cmd.Time()
			if err != nil {
				return util.Error(err), nil
			}
			return &tengo.Time{Value: v}, nil
		}),
		"bool": util.NewUserFunc(func(args ...tengo.Object) (tengo.Object, error) {
			v, err := cmd.Bool()
			if err != nil {
				return util.Error(err), nil
			}
			return tengo.FromInterface(v)
		}),
		"err":       util.NewUserFunc(stdlib.FuncARE(cmd.Err)),
		"val":       util.NewUserFunc(stdlib.FuncARS(cmd.Val)),
		"name":      util.NewUserFunc(stdlib.FuncARS(cmd.Name)),
		"full_name": util.NewUserFunc(stdlib.FuncARS(cmd.FullName)),
	}
}

func NewStringCmd(cmd *redis.StringCmd) *util.Proxy[*redis.StringCmd] {
	return util.NewProxy(cmd).WithConstructor(stringCmdConstructor)
}
