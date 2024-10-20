package transpile

import (
	"github.com/d5/tengo/v2"
	"github.com/d5/tengo/v2/stdlib"
	"io/fs"
	"regexp"
)

func NewReplace(pattern, replace string) (TransFunc, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}
	repl := []byte(replace)
	return func(src []byte) ([]byte, error) {
		ret := re.ReplaceAll(src, repl)
		return ret, nil
	}, nil
}

func NewReplaceFunc(pattern string, fn func([]byte) []byte) (TransFunc, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}
	return func(src []byte) ([]byte, error) {
		ret := re.ReplaceAllFunc(src, fn)
		return ret, nil
	}, nil
}

func NewScriptReplaceFunc(f fs.FS, name string) (TransFunc, error) {
	const (
		input  = "input"
		output = "output"
		empty  = ""
	)
	buf, err := fs.ReadFile(f, name)
	if err != nil {
		return nil, err
	}
	script := tengo.NewScript(buf)
	script.SetImports(stdlib.GetModuleMap(stdlib.AllModuleNames()...))
	err = script.Add(input, empty)
	if err != nil {
		return nil, err
	}
	err = script.Add(output, empty)
	if err != nil {
		return nil, err
	}
	processor, err := script.Compile()
	if err != nil {
		return nil, err
	}
	return func(src []byte) ([]byte, error) {
		p := processor.Clone()
		err := p.Set(input, src)
		if err != nil {
			return src, err
		}
		err = p.Run()
		if err != nil {
			return src, err
		}
		v := p.Get(output)
		err = v.Error()
		return v.Bytes(), err
	}, nil
}
