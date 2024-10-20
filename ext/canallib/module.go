package canallib

import (
	"github.com/d5/tengo/v2"
	"lightbox/sandbox"
)

var Entry = sandbox.NewRegistry(
	"canal",
	map[string]tengo.Object{
		"new_config": &tengo.UserFunction{
			Value: newConfig,
		},
	},
	map[string]sandbox.UserFunction{
		"new": newCanalWithApp,
	})
