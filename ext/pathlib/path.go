package pathlib

import (
	"github.com/d5/tengo/v2"
	"github.com/d5/tengo/v2/stdlib"
	"lightbox/ext/util"
	"lightbox/sandbox"
	"path"
	"path/filepath"
)

var module = map[string]tengo.Object{
	"ext": &tengo.UserFunction{
		Name:  "ext",
		Value: stdlib.FuncASRS(path.Ext),
	},
	"base": &tengo.UserFunction{
		Name:  "base",
		Value: stdlib.FuncASRS(path.Base),
	},
	"dir": &tengo.UserFunction{
		Name:  "dir",
		Value: stdlib.FuncASRS(path.Dir),
	},
	"clean": &tengo.UserFunction{
		Name:  "clean",
		Value: stdlib.FuncASRS(path.Clean),
	},
	"split": &tengo.UserFunction{
		Name:  "split",
		Value: util.FuncASRSS(path.Split),
	},
	"join": &tengo.UserFunction{
		Name:  "join",
		Value: util.FuncASsRS(path.Join),
	},
	"is_abs": &tengo.UserFunction{
		Name:  "is_abs",
		Value: util.FuncASRB(path.IsAbs),
	},
	"match": &tengo.UserFunction{
		Name:  "match",
		Value: util.FuncASRBE(path.Match),
	},
}

var filePathModule = map[string]tengo.Object{
	"ext": &tengo.UserFunction{
		Name:  "ext",
		Value: stdlib.FuncASRS(filepath.Ext),
	},
	"base": &tengo.UserFunction{
		Name:  "base",
		Value: stdlib.FuncASRS(filepath.Base),
	},
	"dir": &tengo.UserFunction{
		Name:  "dir",
		Value: stdlib.FuncASRS(filepath.Dir),
	},
	"clean": &tengo.UserFunction{
		Name:  "clean",
		Value: stdlib.FuncASRS(filepath.Clean),
	},
	"split": &tengo.UserFunction{
		Name:  "split",
		Value: util.FuncASRSS(filepath.Split),
	},
	"join": &tengo.UserFunction{
		Name:  "join",
		Value: util.FuncASsRS(filepath.Join),
	},
	"abs": &tengo.UserFunction{
		Name:  "abs",
		Value: stdlib.FuncASRSE(filepath.Abs),
	},
	"match": &tengo.UserFunction{
		Name:  "match",
		Value: util.FuncASRBE(filepath.Match),
	},
}

var PathEntry = sandbox.NewRegistry("path", module, nil)
var FilePathEntry = sandbox.NewRegistry("filepath", filePathModule, nil)
