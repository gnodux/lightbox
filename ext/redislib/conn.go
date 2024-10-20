package redislib

import (
	"context"
	"github.com/d5/tengo/v2"
	"github.com/d5/tengo/v2/stdlib"
	"github.com/go-redis/redis/v8"
	"lightbox/ext/util"
	"time"
)

func newClientConstructor() util.ConstructFn[*util.Proxy[*redis.Client]] {
	defaultTTL := &tengo.Int{Value: int64(1 * time.Hour)}
	return func(proxy *util.Proxy[*redis.Client]) {
		client := proxy.Value
		proxy.Props = map[string]tengo.Object{
			"get": util.NewUserFunc(func(args ...tengo.Object) (tengo.Object, error) {
				if len(args) != 1 {
					return nil, tengo.ErrWrongNumArguments
				}
				key, _ := tengo.ToString(args[0])
				ret := client.Get(context.Background(), key)
				return NewStringCmd(ret), nil
			}),
			"ttl": defaultTTL,
			"set": util.NewUserFunc(func(args ...tengo.Object) (tengo.Object, error) {
				if len(args) < 2 {
					return nil, tengo.ErrWrongNumArguments
				}
				key, _ := tengo.ToString(args[0])
				value := tengo.ToInterface(args[1])
				expire := defaultTTL.Value
				if len(args) == 3 {
					expire, _ = tengo.ToInt64(args[2])
				}
				r := client.Set(context.Background(), key, value, time.Duration(expire))
				return NewStatusCmd(r), nil
			}),
			"setnx": util.NewUserFunc(func(args ...tengo.Object) (tengo.Object, error) {
				if len(args) != 3 {
					return nil, tengo.ErrWrongNumArguments
				}
				key, _ := tengo.ToString(args[0])
				value := tengo.ToInterface(args[1])
				expire := defaultTTL.Value
				if len(args) == 3 {
					expire, _ = tengo.ToInt64(args[2])
				}
				r := client.SetNX(context.Background(), key, value, time.Duration(expire))
				return NewBoolCmd(r), nil
			}),
			"del": util.NewUserFunc(func(args ...tengo.Object) (tengo.Object, error) {
				if len(args) == 0 {
					return nil, tengo.ErrWrongNumArguments
				}
				var keys = make([]string, len(args))
				for idx, arg := range args {
					v, _ := tengo.ToString(arg)
					keys[idx] = v
				}
				r := client.Del(context.Background(), keys...)
				return NewIntCmd(r), nil
			}),
			"exists": util.NewUserFunc(func(args ...tengo.Object) (tengo.Object, error) {
				if len(args) == 0 {
					return nil, tengo.ErrWrongNumArguments
				}
				var keys = make([]string, len(args))
				for idx, arg := range args {
					v, _ := tengo.ToString(arg)
					keys[idx] = v
				}
				r := client.Exists(context.Background(), keys...)
				return NewIntCmd(r), nil
			}),
			"close": util.NewUserFunc(stdlib.FuncARE(client.Close)),
		}
	}
}

func NewClient(url string) (*util.Proxy[*redis.Client], error) {
	opt, err := redis.ParseURL(url)
	if err != nil {
		return nil, err
	}
	client := redis.NewClient(opt)

	return util.NewProxy(client).WithConstructor(newClientConstructor()), nil
}
