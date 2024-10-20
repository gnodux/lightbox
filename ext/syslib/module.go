package syslib

import (
	"github.com/d5/tengo/v2"
	"lightbox/sandbox"
	"regexp"
)

var module = map[string]tengo.Object{
	"must": &tengo.UserFunction{
		Name:  "must",
		Value: sysMust,
	},
	"require": &tengo.UserFunction{
		Value: sysRequire,
	},
	"wait": &tengo.UserFunction{
		Name:  "wait",
		Value: sysWait,
	},
	"exit": &tengo.UserFunction{
		Name:  "exit",
		Value: sysExit,
	},
	"args": &tengo.UserFunction{
		Name:  "args",
		Value: sysArgs,
	},
}

var appModule = map[string]sandbox.UserFunction{
	"exec":           sysExec,
	"fork":           sysFork,
	"env":            sysEnv,
	"config":         sysConfig,
	"get_config":     sysGetConfig,
	"props":          sysConfig,
	"prop":           sysGetConfig,
	"get":            sysGetEnv,
	"set":            sysSet,
	"get_env":        sysGetEnv,
	"set_env":        sysSet,
	"add_transpiler": sysAddTranspiler,
}

var (
	allRegex = map[string]*regexp.Regexp{
		`sys.get_env("$1")||$2`: regexp.MustCompile("%{(.*?)\\|(.*?)}"),
		`sys.get_env("$1")`:     regexp.MustCompile("%{(.*?)}"),
		`sys.prop("$1")||$2`:    regexp.MustCompile("\\$\\${(.*?)\\|(.*?)}"),
		`sys.prop("$1")`:        regexp.MustCompile("\\$\\${(.*?)}"),
	}
)

//variableProcessor
//snippet:name=$$(name);prefix=$$;body=$${$1};
//snippet:name="%{envname};prefix=%;body=%{$1};
func variableProcessor(input []byte) (out []byte, err error) {
	out = input
	for repl, re := range allRegex {
		out = re.ReplaceAll(out, []byte(repl))
	}
	//fmt.Println(string(out))
	return
}

var Entry = sandbox.NewRegistry("sys", module, appModule).WithTranspiler(variableProcessor)
