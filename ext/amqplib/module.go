package amqplib

import (
	"github.com/d5/tengo/v2"
	"lightbox/ext/util"
	"lightbox/sandbox"
)

var Entry = sandbox.NewRegistry(
	"amqp",
	nil,
	map[string]sandbox.UserFunction{
		"dial": func(app *sandbox.Applet, args ...tengo.Object) (tengo.Object, error) {
			url, _ := tengo.ToString(args[0])
			if conn, err := Dial(url, app); err != nil {
				return util.Error(err), nil
			} else {
				return conn, nil
			}
		},
	})
