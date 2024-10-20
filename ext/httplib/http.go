package httplib

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/d5/tengo/v2"
	tjson "github.com/d5/tengo/v2/stdlib/json"
	log "github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"lightbox/env"
	"lightbox/ext/util"
	"lightbox/sandbox"
	"net/http"
	"net/url"
	"strings"
)

func urlEncode(args ...tengo.Object) (tengo.Object, error) {
	if len(args) != 1 {
		return util.Error(tengo.ErrWrongNumArguments), nil
	}
	switch arg := args[0].(type) {
	//case *tengo.String:
	//case *tengo.Map:
	//case *tengo.ImmutableMap:
	default:
		s, _ := tengo.ToString(arg)
		u := url.QueryEscape(s)
		return tengo.FromInterface(u)
	}
	return nil, nil
}

func urlDecode(args ...tengo.Object) (tengo.Object, error) {
	if len(args) != 1 {
		return util.Error(tengo.ErrWrongNumArguments), nil
	}
	switch arg := args[0].(type) {
	//case *tengo.String:
	//case *tengo.Map:
	//case *tengo.ImmutableMap:
	default:
		s, _ := tengo.ToString(arg)
		u, err := url.QueryUnescape(s)
		if err != nil {
			return util.Error(err), nil
		}
		return tengo.FromInterface(u)
	}
	return nil, nil
}

func httpRequestInternal(app *sandbox.Applet, arg *tengo.Map) (tengo.Object, error) {
	client, ok := env.GetVal[*http.Client](app.Context, ClientKey)
	if !ok {
		return nil, errors.New("client not found")
	}
	if client == nil {
		return nil, errors.New("client is nil")
	}
	aUrl, ok := arg.Value["url"]
	if !ok {
		return tengo.FromInterface(errors.New("request require a url argument"))
	}
	url, ok := aUrl.(*tengo.String)
	if !ok {
		return tengo.FromInterface(errors.New("url parameter must be string"))
	}
	aMethod, ok := arg.Value["method"]
	if !ok {
		return tengo.FromInterface(errors.New("request require a method argument"))
	}
	method, ok := aMethod.(*tengo.String)
	if !ok {
		return tengo.FromInterface(errors.New("method parameter must be string"))
	}
	aHeader, ok := arg.Value["header"]
	var header *tengo.Map
	if !ok {
		header = &tengo.Map{Value: map[string]tengo.Object{}}
	} else {
		header, ok = aHeader.(*tengo.Map)
		if !ok {
			return tengo.FromInterface(errors.New("header parameter must be a map"))
		}
	}
	//success := arg.Value["success"]
	//failed := arg.Value["fail"]

	var reader io.Reader
	aPayload, ok := arg.Value["data"]
	if !ok {
		reader = bytes.NewReader([]byte{})
	} else {
		switch payload := aPayload.(type) {
		case *tengo.String:
			reader = bytes.NewBufferString(payload.Value)
			break
		case *tengo.Bytes:
			reader = bytes.NewBuffer(payload.Value)
			break
		default:
			if buf, err := tjson.Encode(payload); err != nil {
				return nil, err
			} else {
				if _, ok := header.Value["Content-Type"]; !ok {
					header.Value["Content-Type"] = &tengo.String{Value: "application/json"}
				}
				reader = bytes.NewReader(buf)
			}
		}
	}

	request, err := http.NewRequest(strings.ToUpper(method.Value), url.Value, reader)

	if err != nil {
		return tengo.FromInterface(err)
	}
	if request.Header == nil {
		request.Header = http.Header{}
	}
	//request.MultipartReader()
	for k, h := range header.Value {
		if headerVal, ok := tengo.ToString(h); ok {
			request.Header.Set(k, headerVal)
		}
	}
	log.WithFields(map[string]interface{}{
		"url":     request.URL,
		"method":  request.Method,
		"header":  request.Header,
		"payload": aPayload,
	}).Trace("start http request")
	response, err := client.Do(request)
	if err != nil {
		return tengo.FromInterface(err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Error("close body error", err)
		}
	}(response.Body)

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return tengo.FromInterface(err)
	}
	responseHeader := make(map[string]interface{}, len(response.Header))
	for k, v := range response.Header {
		responseHeader[k] = strings.Join(v, ";")
	}
	result := map[string]interface{}{}
	result["code"] = response.StatusCode
	result["header"] = responseHeader
	result["body"] = body
	if strings.Contains(response.Header.Get("Content-Type"), "application/json") {
		if jsonObj, err := tjson.Decode(body); err == nil {
			result["data"] = jsonObj
		} else {
			return tengo.FromInterface(err)
		}
	} else {
		result["data"] = body
	}
	log.WithFields(map[string]interface{}{
		"code":   response.StatusCode,
		"header": responseHeader,
		"body":   body,
		"url":    url.Value,
	}).Trace("end http request")
	return tengo.FromInterface(result)
}

func httpRequest(app *sandbox.Applet, args ...tengo.Object) (tengo.Object, error) {
	if len(args) == 0 {
		return tengo.FromInterface(errors.New("require 1 argument"))
	}
	if arg, ok := args[0].(*tengo.Map); ok {
		return httpRequestInternal(app, arg)
	}
	return tengo.FromInterface(errors.New("request require a ImmutableMap arg"))
}

//httpGet
//snippet:name=http.get;prefix=get;body=get(${1:url});
//snippet:name=http.get({url:...,header:{}});prefix=get;body=get({url:${1:url},header:{$2}});
func httpGet(app *sandbox.Applet, args ...tengo.Object) (tengo.Object, error) {
	if len(args) == 0 {
		return tengo.FromInterface(errors.New("require 1 argument"))
	}
	switch a0 := args[0].(type) {
	case *tengo.Map:
		a0.Value["method"] = &tengo.String{Value: "get"}

		return httpRequestInternal(app, a0)
	case *tengo.String:
		return httpRequestInternal(app, &tengo.Map{Value: map[string]tengo.Object{
			"url":    args[0],
			"method": &tengo.String{Value: "GET"},
		}})
	}
	return tengo.FromInterface(fmt.Errorf("unknown arguments %v", args))

}

//httPost
//snippet:name=http.post;prefix=post;body=post({url:${1:url},header:{$2},data:${3:data}});
func httpPost(app *sandbox.Applet, args ...tengo.Object) (tengo.Object, error) {
	if len(args) == 0 {
		return tengo.FromInterface(errors.New("no argument"))
	}
	if a0, ok := args[0].(*tengo.Map); ok {
		if _, ok = a0.Value["method"]; !ok {
			a0.Value["method"] = &tengo.String{Value: "post"}
		}
		return httpRequestInternal(app, a0)
	}
	return httpRequestInternal(app, &tengo.Map{Value: map[string]tengo.Object{
		"url":    args[0],
		"header": &tengo.Map{Value: map[string]tengo.Object{"Content-Type": args[1]}},
		"data":   args[2],
	}})
}

func newHttpCookie(args ...tengo.Object) (tengo.Object, error) {
	c := &http.Cookie{}
	_ = util.StructFromArgs(args, c)
	return util.NewReflectProxy(c), nil
}
