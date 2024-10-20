package httplib

import (
	"fmt"
	"net/http"
)

type HttpContext struct {
	Server   *httpServer
	Response http.ResponseWriter
	Request  *http.Request
	Vars     map[string]string //path variables
}

func NewHttpContext(server *httpServer, response http.ResponseWriter, request *http.Request, vars map[string]string) *HttpContext {
	return &HttpContext{Server: server, Response: response, Request: request, Vars: vars}
}

type MediaConverterManager struct {
	formatters []MediaConverter
}

func NewConverterManager(converters ...MediaConverter) *MediaConverterManager {
	m := &MediaConverterManager{}
	m.Add(converters...)
	return m
}
func NewDefaultConverterManager(converters ...MediaConverter) *MediaConverterManager {
	m := NewConverterManager(converters...)
	m.Add(&HtmlViewConverter{})
	m.Add(&JsonConverter{})
	return m
}

func (g *MediaConverterManager) Write(message interface{}, ctx *HttpContext, mediaType string) error {
	for _, formatter := range g.formatters {
		if formatter.CanWrite(message, mediaType) {
			return formatter.Write(ctx, message)
		}
	}
	return fmt.Errorf("not found message converter:%s", mediaType)
}
func (g *MediaConverterManager) Read(ctx *HttpContext, mediaType string) (interface{}, error) {
	for _, converter := range g.formatters {
		if converter.CanRead(mediaType, ctx) {
			return converter.Read(ctx)
		}
	}
	return nil, fmt.Errorf("not found message converter:%s", mediaType)
}
func (g *MediaConverterManager) Add(formatters ...MediaConverter) {
	for _, formatter := range formatters {
		g.formatters = append(g.formatters, formatter)
	}
}

type MediaConverter interface {
	Read(context *HttpContext) (interface{}, error)
	Write(context *HttpContext, data interface{}) error
	CanWrite(data interface{}, mediaType string) bool
	CanRead(mediaType string, context *HttpContext) bool
}
