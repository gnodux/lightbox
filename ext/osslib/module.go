package osslib

import (
	"github.com/d5/tengo/v2"
	"lightbox/sandbox"
)

var module = map[string]tengo.Object{
	"open": &tengo.UserFunction{Name: "open", Value: newOSS},
}
var Entry = sandbox.NewRegistry("oss", module, nil)
