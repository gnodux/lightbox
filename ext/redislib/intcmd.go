package redislib

import (
	"github.com/d5/tengo/v2"
	"github.com/d5/tengo/v2/stdlib"
	"github.com/go-redis/redis/v8"
	"lightbox/ext/util"
)

func intCmdConstructor(proxy *util.Proxy[*redis.IntCmd]) {
	cmd := proxy.Value
	proxy.Props = map[string]tengo.Object{
		"result": util.NewUserFunc(func(args ...tengo.Object) (tengo.Object, error) {
			r, err := cmd.Result()
			if err != nil {
				return util.Error(err), nil
			}
			return &tengo.Int{Value: r}, nil
		}),
		"err":       util.NewUserFunc(stdlib.FuncARE(cmd.Err)),
		"val":       util.NewUserFunc(stdlib.FuncARI64(cmd.Val)),
		"name":      util.NewUserFunc(stdlib.FuncARS(cmd.Name)),
		"full_name": util.NewUserFunc(stdlib.FuncARS(cmd.FullName)),
	}
}

func NewIntCmd(cmd *redis.IntCmd) *util.Proxy[*redis.IntCmd] {
	return util.NewProxy(cmd).WithConstructor(intCmdConstructor)
}
