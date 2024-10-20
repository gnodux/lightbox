package httplib

import (
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/d5/tengo/v2"
	"lightbox/ext/util"
	"lightbox/sandbox"
	"net/http"
	"net/url"
	"reflect"
	"time"
)

var module = map[string]tengo.Object{
	"new_cookie": &tengo.UserFunction{Value: newHttpCookie},
	"encode_url": &tengo.UserFunction{Value: urlEncode},
	"decode_url": &tengo.UserFunction{Value: urlDecode},
	"status": &tengo.ImmutableMap{Value: map[string]tengo.Object{
		"continue":           &tengo.Int{Value: http.StatusContinue},
		"switchingprotocols": &tengo.Int{Value: http.StatusSwitchingProtocols},
		"processing":         &tengo.Int{Value: http.StatusProcessing},
		"earlyhints":         &tengo.Int{Value: http.StatusEarlyHints},

		"ok":                   &tengo.Int{Value: http.StatusOK},
		"created":              &tengo.Int{Value: http.StatusCreated},
		"accepted":             &tengo.Int{Value: http.StatusAccepted},
		"nonauthoritativeinfo": &tengo.Int{Value: http.StatusNonAuthoritativeInfo},
		"nocontent":            &tengo.Int{Value: http.StatusNoContent},
		"resetcontent":         &tengo.Int{Value: http.StatusResetContent},
		"partialcontent":       &tengo.Int{Value: http.StatusPartialContent},
		"multistatus":          &tengo.Int{Value: http.StatusMultiStatus},
		"alreadyreported":      &tengo.Int{Value: http.StatusAlreadyReported},
		"imused":               &tengo.Int{Value: http.StatusIMUsed},

		"multiplechoices":  &tengo.Int{Value: http.StatusMultipleChoices},
		"movedpermanently": &tengo.Int{Value: http.StatusMovedPermanently},
		"found":            &tengo.Int{Value: http.StatusFound},
		"seeother":         &tengo.Int{Value: http.StatusSeeOther},
		"notmodified":      &tengo.Int{Value: http.StatusNotModified},
		"useproxy":         &tengo.Int{Value: http.StatusUseProxy},

		"temporaryredirect": &tengo.Int{Value: http.StatusTemporaryRedirect},
		"permanentredirect": &tengo.Int{Value: http.StatusPermanentRedirect},

		"badrequest":                   &tengo.Int{Value: http.StatusBadRequest},
		"unauthorized":                 &tengo.Int{Value: http.StatusUnauthorized},
		"paymentrequired":              &tengo.Int{Value: http.StatusPaymentRequired},
		"forbidden":                    &tengo.Int{Value: http.StatusForbidden},
		"notfound":                     &tengo.Int{Value: http.StatusNotFound},
		"methodnotallowed":             &tengo.Int{Value: http.StatusMethodNotAllowed},
		"notacceptable":                &tengo.Int{Value: http.StatusNotAcceptable},
		"proxyauthrequired":            &tengo.Int{Value: http.StatusProxyAuthRequired},
		"requesttimeout":               &tengo.Int{Value: http.StatusRequestTimeout},
		"conflict":                     &tengo.Int{Value: http.StatusConflict},
		"gone":                         &tengo.Int{Value: http.StatusGone},
		"lengthrequired":               &tengo.Int{Value: http.StatusLengthRequired},
		"preconditionfailed":           &tengo.Int{Value: http.StatusPreconditionFailed},
		"requestentitytoolarge":        &tengo.Int{Value: http.StatusRequestEntityTooLarge},
		"requesturitoolong":            &tengo.Int{Value: http.StatusRequestURITooLong},
		"unsupportedmediatype":         &tengo.Int{Value: http.StatusUnsupportedMediaType},
		"requestedrangenotsatisfiable": &tengo.Int{Value: http.StatusRequestedRangeNotSatisfiable},
		"expectationfailed":            &tengo.Int{Value: http.StatusExpectationFailed},
		"teapot":                       &tengo.Int{Value: http.StatusTeapot},
		"misdirectedrequest":           &tengo.Int{Value: http.StatusMisdirectedRequest},
		"unprocessableentity":          &tengo.Int{Value: http.StatusUnprocessableEntity},
		"locked":                       &tengo.Int{Value: http.StatusLocked},
		"faileddependency":             &tengo.Int{Value: http.StatusFailedDependency},
		"tooearly":                     &tengo.Int{Value: http.StatusTooEarly},
		"upgraderequired":              &tengo.Int{Value: http.StatusUpgradeRequired},
		"preconditionrequired":         &tengo.Int{Value: http.StatusPreconditionRequired},
		"toomanyrequests":              &tengo.Int{Value: http.StatusTooManyRequests},
		"requestheaderfieldstoolarge":  &tengo.Int{Value: http.StatusRequestHeaderFieldsTooLarge},
		"unavailableforlegalreasons":   &tengo.Int{Value: http.StatusUnavailableForLegalReasons},

		"internalservererror":           &tengo.Int{Value: http.StatusInternalServerError},
		"notimplemented":                &tengo.Int{Value: http.StatusNotImplemented},
		"badgateway":                    &tengo.Int{Value: http.StatusBadGateway},
		"serviceunavailable":            &tengo.Int{Value: http.StatusServiceUnavailable},
		"gatewaytimeout":                &tengo.Int{Value: http.StatusGatewayTimeout},
		"httpversionnotsupported":       &tengo.Int{Value: http.StatusHTTPVersionNotSupported},
		"variantalsonegotiates":         &tengo.Int{Value: http.StatusVariantAlsoNegotiates},
		"insufficientstorage":           &tengo.Int{Value: http.StatusInsufficientStorage},
		"loopdetected":                  &tengo.Int{Value: http.StatusLoopDetected},
		"notextended":                   &tengo.Int{Value: http.StatusNotExtended},
		"networkauthenticationrequired": &tengo.Int{Value: http.StatusNetworkAuthenticationRequired},
	}},
}
var appModule = map[string]sandbox.UserFunction{
	"server":  newHttpServer,
	"get":     httpGet,
	"post":    httpPost,
	"request": httpRequest,
	"client":  getHttpClient,
	"set_proxy": func(app *sandbox.Applet, args ...tengo.Object) (tengo.Object, error) {
		if len(args) != 1 {
			return nil, tengo.ErrWrongNumArguments
		}
		arg, ok := tengo.ToString(args[0])
		if !ok {
			return tengo.FromInterface(errors.New("require a url string"))
		}
		proxyUrl, err := url.Parse(arg)
		if err != nil {
			return tengo.FromInterface(err)
		}
		client, ok := app.Context.Get(ClientKey)
		if !ok || client == nil {
			return tengo.FromInterface(errors.New("client not found"))
		}
		c, ok := client.(*http.Client)
		if !ok {
			return tengo.FromInterface(errors.New("client is not a http client"))
		}

		t, ok := c.Transport.(*http.Transport)
		if !ok {
			return tengo.FromInterface(errors.New("client transport is not a http.Transport"))
		}
		t.Proxy = http.ProxyURL(proxyUrl)

		return tengo.FromInterface(arg)
	},
	"set_timeout": func(app *sandbox.Applet, args ...tengo.Object) (ret tengo.Object, err error) {
		if len(args) != 1 {
			return nil, tengo.ErrWrongNumArguments
		}
		client, ok := app.Context.Get(ClientKey)
		if !ok || client == nil {
			return tengo.FromInterface(errors.New("client not found"))
		}
		c, ok := client.(*http.Client)
		if !ok {
			return tengo.FromInterface(errors.New("client is not a http client"))
		}
		arg := tengo.ToInterface(args[0])
		switch d := arg.(type) {
		case string:
			c.Timeout, err = time.ParseDuration(d)
		case int:
			c.Timeout = time.Duration(d)
		case int64:
			c.Timeout = time.Duration(d)
		case nil:
			err = errors.New("timeout is nil")
		default:
			err = fmt.Errorf("timeout<%s> is not a duration", reflect.TypeOf(arg))
		}
		return tengo.FromInterface(err)
	},
}

const (
	// ClientKey client_key
	ClientKey = "http_default_client"
)

func getHttpClient(applet *sandbox.Applet, args ...tengo.Object) (tengo.Object, error) {
	client, ok := applet.Context.Get(ClientKey)
	if !ok {
		return nil, nil
	}
	return util.NewReflectProxy(client), nil
}

var Entry = sandbox.NewRegistry("http", module, appModule).
	WithHook(sandbox.NewHook(sandbox.SigInitialized, func(applet *sandbox.Applet) error {
		//初始化默认的http client，跳过证书验证
		applet.Context.Set(ClientKey, &http.Client{
			Timeout: time.Second * 5,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		})
		return nil
	}))
