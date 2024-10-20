package redislib

import (
	"github.com/d5/tengo/v2"
	"github.com/d5/tengo/v2/stdlib"
	"github.com/go-redis/redis/v8"
	"lightbox/ext/util"
)

func statusCmdConstructor(proxy *util.Proxy[*redis.StatusCmd]) {
	cmd := proxy.Value
	proxy.Props = map[string]tengo.Object{
		"result": util.NewUserFunc(func(args ...tengo.Object) (tengo.Object, error) {
			r, err := cmd.Result()
			if err != nil {
				return util.Error(err), nil
			}
			return &tengo.String{Value: r}, nil
		}),
		"err":       util.NewUserFunc(stdlib.FuncARE(cmd.Err)),
		"val":       util.NewUserFunc(stdlib.FuncARS(cmd.Val)),
		"name":      util.NewUserFunc(stdlib.FuncARS(cmd.Name)),
		"full_name": util.NewUserFunc(stdlib.FuncARS(cmd.FullName)),
	}
}

func NewStatusCmd(cmd *redis.StatusCmd) *util.Proxy[*redis.StatusCmd] {
	return util.NewProxy(cmd).WithConstructor(statusCmdConstructor)
}
