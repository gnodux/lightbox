package httplib

import (
	"errors"
	"fmt"
	"golang.org/x/sync/singleflight"
	"html/template"
	"io/fs"
	"sync"
)

type HtmlViewConverter struct {
	tpl        *template.Template
	mimeMap    map[string]bool
	defaultExt string
	sync.Once
	g singleflight.Group
}

func (h *HtmlViewConverter) init() {
	h.tpl = template.New("htmlView")
	h.mimeMap = map[string]bool{
		"text/html": true,
	}
	h.defaultExt = ".tpl.html"
}
func (h *HtmlViewConverter) CanWrite(data interface{}, mediaType string) bool {
	h.Do(h.init)
	v, ok := h.mimeMap[mediaType]
	if !ok {
		return ok
	}
	return v
}
func (h *HtmlViewConverter) CanRead(mediaType string, context *HttpContext) bool {
	h.Do(h.init)
	return false
}

func (h *HtmlViewConverter) Write(context *HttpContext, data interface{}) error {
	h.Do(h.init)
	viewName, ok := context.Vars["view"]
	if !ok {
		return fmt.Errorf("viewName %s not found", viewName)
	}
	tp, err, _ := h.g.Do(viewName, func() (interface{}, error) {
		t := h.tpl.Lookup(viewName)
		if t == nil {
			viewFs, err := fs.Sub(context.Server.app, "view")
			if err != nil {
				return nil, err
			}
			viewFileName := viewName + h.defaultExt
			buf, err := fs.ReadFile(viewFs, viewFileName)
			if err != nil {
				return nil, err
			}
			t, err = h.tpl.New(viewName).Parse(string(buf))
			if err != nil {
				return nil, err
			}
		}
		return t.Clone()
	})
	if err != nil {
		return err
	}
	if tp == nil {
		return fmt.Errorf("view %s not found", viewName)
	}
	return tp.(*template.Template).Execute(context.Response, map[string]interface{}{
		"context":  context,
		"response": context.Response,
		"request":  context.Request,
		"view":     viewName,
		"data":     data,
	})
}
func (h *HtmlViewConverter) Read(context *HttpContext) (interface{}, error) {
	h.Do(h.init)
	return nil, errors.New("html view not support unmarshal")
}
