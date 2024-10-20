package redislib

import (
	"github.com/d5/tengo/v2"
	"github.com/go-redis/redis/v8"
	"lightbox/ext/util"
	"lightbox/sandbox"
)

var module = map[string]tengo.Object{
	"keep_ttl": &tengo.Int{Value: redis.KeepTTL},
	"dial": util.NewUserFunc(func(args ...tengo.Object) (tengo.Object, error) {
		if len(args) != 1 {
			return nil, tengo.ErrWrongNumArguments
		}
		url, _ := tengo.ToString(args[0])
		return NewClient(url)
	}),
}

var Entry = sandbox.NewRegistry("redis", module, nil)
