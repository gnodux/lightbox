package tpllib

import (
	"errors"
	"github.com/d5/tengo/v2"
	"lightbox/sandbox"
)

var module = map[string]tengo.Object{
	//snippet:desc=if text start with '@',load file from path;
	"text": &tengo.UserFunction{Name: "text", Value: func(args ...tengo.Object) (tengo.Object, error) {
		if len(args) != 1 {
			return nil, tengo.ErrWrongNumArguments
		}
		s, ok := tengo.ToString(args[0])
		if !ok {
			return nil, errors.New("except a string")
		}
		template, err := NewTextTemplate(s)
		if err != nil {
			return nil, err
		}
		return template.wrap(), nil
	}},
	//snippet:desc=if text start with '@',load file from path;
	"html": &tengo.UserFunction{Name: "html", Value: func(args ...tengo.Object) (ret tengo.Object, err error) {

		if len(args) != 1 {
			return nil, tengo.ErrWrongNumArguments
		}
		s, ok := tengo.ToString(args[0])
		if !ok {
			return nil, errors.New("except a string")
		}
		template, err := NewHtmlTemplate(s)
		if err != nil {
			return nil, err
		}
		return template.wrap(), nil
	}},
}

var Entry = sandbox.NewRegistry("tpl", module, nil)
