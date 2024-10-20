package xlslib

import (
	"github.com/d5/tengo/v2"
	"lightbox/ext/util"
	"lightbox/sandbox"
)

var module = map[string]tengo.Object{
	"open": &tengo.UserFunction{
		Name:  "open",
		Value: util.FuncASROE(open),
	},
	"new": &tengo.UserFunction{
		Name:  "new",
		Value: util.FuncARO(newXls),
	},
}

var Entry = sandbox.NewRegistry("xls", module, nil)
