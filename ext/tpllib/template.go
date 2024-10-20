package tpllib

import (
	"bytes"
	"encoding/base64"
	"errors"
	"github.com/d5/tengo/v2"
	hp "html/template"
	"lightbox/ext/util"
	"strings"
	tp "text/template"
)

type htmlTemplate struct {
	tpl *hp.Template
}

//Render
//snippet:name=tpl.render;prefix=render;body=render(${1:object or map});
func (t *htmlTemplate) Render(data interface{}) (string, error) {
	if t == nil || t.tpl == nil {
		return "", errors.New("template is nil")
	}
	buf := bytes.NewBuffer([]byte{})
	err := t.tpl.Execute(buf, data)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}
func (t *htmlTemplate) wrap() *tengo.ImmutableMap {
	return &tengo.ImmutableMap{Value: map[string]tengo.Object{
		"render": &tengo.UserFunction{Name: "render", Value: func(args ...tengo.Object) (tengo.Object, error) {
			if len(args) == 0 {
				return nil, tengo.ErrWrongNumArguments
			}
			var (
				result string
				err    error
			)
			switch arg := args[0].(type) {
			case *tengo.ImmutableMap, *tengo.Map:
				data := tengo.ToInterface(arg)
				result, err = t.Render(data)
				if err != nil {
					return util.Error(err), nil
				}
			default:
				result = ""
				err = tengo.ErrInvalidArgumentType{Name: "data", Expected: "Map", Found: arg.TypeName()}
			}
			return &tengo.String{Value: result}, err
		}},
	}}
}

func NewHtmlTemplate(tpl string) (t *htmlTemplate, err error) {
	htpl := htmlTemplate{}
	if strings.HasPrefix(tpl, "@") {
		htpl.tpl, err = hp.ParseFiles(tpl[1:])
	} else {
		htpl.tpl, err = hp.New(base64.StdEncoding.EncodeToString([]byte(tpl))).Parse(tpl)
	}
	return &htpl, err
}

type TextTemplate struct {
	tpl *tp.Template
}

func (t *TextTemplate) Render(data interface{}) (string, error) {
	if t == nil || t.tpl == nil {
		return "", errors.New("template is nil")
	}
	buf := bytes.NewBuffer([]byte{})
	err := t.tpl.Execute(buf, data)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}
func (t *TextTemplate) wrap() *tengo.ImmutableMap {
	return &tengo.ImmutableMap{Value: map[string]tengo.Object{
		"render": &tengo.UserFunction{Name: "render", Value: func(args ...tengo.Object) (tengo.Object, error) {
			if len(args) == 0 {
				return nil, tengo.ErrWrongNumArguments
			}
			var (
				result string
				err    error
			)
			switch arg := args[0].(type) {
			case *tengo.ImmutableMap, *tengo.Map:
				data := tengo.ToInterface(arg)
				result, err = t.Render(data)
				if err != nil {
					return util.Error(err), nil
				}
			default:
				result = ""
				err = tengo.ErrInvalidArgumentType{Name: "data", Expected: "Map", Found: arg.TypeName()}
			}

			return &tengo.String{Value: result}, err
		}},
	}}
}

func NewTextTemplate(tpl string) (t *TextTemplate, err error) {
	ttpl := TextTemplate{}
	if strings.HasPrefix(tpl, "@") {
		ttpl.tpl, err = tp.ParseFiles(tpl[1:])
	} else {
		ttpl.tpl, err = tp.New(base64.StdEncoding.EncodeToString([]byte(tpl))).Parse(tpl)
	}
	return &ttpl, err
}
