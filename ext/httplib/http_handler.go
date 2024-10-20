package httplib

import (
	"errors"
	"github.com/d5/tengo/v2"
	log "github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"lightbox/ext/util"
	"net/http"
	"strings"
)

const (
	defaultMaxMemory = 32 << 20 // 32 MB
)

func WrapResponse(writer http.ResponseWriter) *tengo.ImmutableMap {
	return &tengo.ImmutableMap{Value: map[string]tengo.Object{
		"body": &tengo.UserFunction{
			Name: "body",
			Value: func(args ...tengo.Object) (tengo.Object, error) {
				return writeResponse(args, writer)
			},
		},
		"redirect": &tengo.UserFunction{
			Value: func(args ...tengo.Object) (ret tengo.Object, err error) {
				if len(args) == 0 {
					return nil, tengo.ErrWrongNumArguments
				}
				url, _ := tengo.ToString(args[0])
				writer.Header().Set("Cache-Control", "must-revalidate, no-store")
				writer.Header().Set("Content-Type", " text/html;charset=UTF-8")
				writer.Header().Set("Location", url)
				writer.WriteHeader(http.StatusFound)
				return nil, err
			},
		},
		"set_cookie": &tengo.UserFunction{Value: func(args ...tengo.Object) (ret tengo.Object, err error) {
			for idx, arg := range args {
				if wrapper, ok := arg.(*util.ReflectProxy); ok {
					if wrapper.Self == nil {
						log.Errorf("arg %d self value is nil", idx)
						continue
					}
					if c, ok := wrapper.Self.(*http.Cookie); ok {
						http.SetCookie(writer, c)
					} else {
						log.Error("object is not a cookie:", idx, ",", c)
					}
				} else {
					log.Error("support reflect wrapper only:", idx, ",", arg)
				}
			}
			return nil, nil
		}},
		"status": &tengo.UserFunction{
			Name: "status",
			Value: func(args ...tengo.Object) (tengo.Object, error) {
				if len(args) != 1 {
					return nil, tengo.ErrWrongNumArguments
				}
				if status, ok := tengo.ToInt(args[0]); ok {
					writer.WriteHeader(status)
				} else {
					return nil, tengo.ErrInvalidArgumentType{Name: "status", Expected: "number", Found: args[0].TypeName()}
				}
				return nil, nil
			},
		},
		"header": &tengo.UserFunction{
			Name: "header",
			Value: func(args ...tengo.Object) (ret tengo.Object, err error) {
				if len(args) != 1 {
					return nil, tengo.ErrWrongNumArguments
				}
				switch m := args[0].(type) {
				case *tengo.Map:
					for k, v := range m.Value {
						if hv, ok := tengo.ToString(v); ok {
							writer.Header().Set(k, hv)
						} else {
							return nil, tengo.ErrInvalidArgumentType{Name: k, Expected: "string", Found: v.TypeName()}
						}
					}
				default:
					return nil, errors.New("headers need a map")
				}
				return
			},
		},
		"json": &tengo.UserFunction{
			Name: "json",
			Value: func(args ...tengo.Object) (ret tengo.Object, err error) {
				writer.Header().Set("Content-Type", "application/json")
				return writeResponse(args, writer)
			},
		},
	}}
}

func writeResponse(args []tengo.Object, r http.ResponseWriter) (tengo.Object, error) {
	for _, arg := range args {
		switch c := arg.(type) {
		case *tengo.String:
			_, err := r.Write([]byte(c.Value))
			if err != nil {
				return nil, err
			}
		case *tengo.Bytes:
			if _, err := r.Write(c.Value); err != nil {
				return nil, err
			}
		default:
			if s, ok := tengo.ToString(c); ok {
				if _, err := r.Write([]byte(s)); err != nil {
					return nil, err
				}
			}
		}
	}
	return nil, nil
}
func WrapRequest(r *http.Request) *tengo.ImmutableMap {
	return &tengo.ImmutableMap{
		Value: map[string]tengo.Object{
			"header": &tengo.UserFunction{
				Name: "header",
				Value: func(args ...tengo.Object) (ret tengo.Object, err error) {
					hm := map[string]tengo.Object{}
					for k, v := range r.Header {
						hm[k] = &tengo.String{Value: strings.Join(v, ";")}
					}
					return &tengo.Map{Value: hm}, nil
				},
			},
			"body": &tengo.UserFunction{
				Name: "body",
				Value: func(args ...tengo.Object) (ret tengo.Object, err error) {

					all, err := ioutil.ReadAll(r.Body)
					if err != nil {
						return nil, err
					}
					defer func(Body io.ReadCloser) {
						err := Body.Close()
						if err != nil {
							log.Error(err)
						}
					}(r.Body)
					return &tengo.Bytes{Value: all}, nil
				},
			},
			"method": &tengo.UserFunction{
				Value: func(args ...tengo.Object) (ret tengo.Object, err error) {
					return &tengo.String{Value: r.Method}, nil
				},
			},
			"url": &tengo.UserFunction{
				Value: func(args ...tengo.Object) (ret tengo.Object, err error) {
					return &tengo.String{Value: r.RequestURI}, nil
				},
			},
			"form": &tengo.UserFunction{
				Value: func(args ...tengo.Object) (ret tengo.Object, err error) {
					err = r.ParseForm()
					if err != nil {
						return nil, err
					}
					result := &tengo.ImmutableMap{Value: map[string]tengo.Object{}}
					if len(args) > 0 {
						for _, arg := range args {
							if k, ok := tengo.ToString(arg); ok {
								result.Value[k] = &tengo.String{Value: r.FormValue(k)}
							}
						}
					} else {
						for k, v := range r.Form {
							if v != nil && len(v) > 0 {
								result.Value[k] = &tengo.String{Value: v[0]}
							}
						}
					}

					return result, nil
				},
			},
			"post_form": &tengo.UserFunction{
				Value: func(args ...tengo.Object) (ret tengo.Object, err error) {
					if r.PostForm == nil {
						err = r.ParseMultipartForm(defaultMaxMemory)
					}
					result := &tengo.ImmutableMap{Value: map[string]tengo.Object{}}
					if len(args) > 0 {
						for _, arg := range args {
							if k, ok := tengo.ToString(arg); ok {
								result.Value[k] = &tengo.String{Value: r.PostFormValue(k)}
							}
						}
					} else {
						for k, v := range r.PostForm {
							if v != nil && len(v) > 0 {
								result.Value[k] = &tengo.String{Value: v[0]}
							}
						}
					}

					return result, nil
				},
			},
			"file": &tengo.UserFunction{
				Value: func(args ...tengo.Object) (tengo.Object, error) {
					result := &tengo.ImmutableMap{Value: map[string]tengo.Object{}}
					var err error
					if r.MultipartForm == nil {
						err = r.ParseMultipartForm(defaultMaxMemory)
						if err != nil {
							return nil, err
						}
					}

					for _, arg := range args {
						if k, ok := tengo.ToString(arg); ok {
							err = func() error {
								f, h, err := r.FormFile(k)
								if err != nil {
									return err
								}
								defer f.Close()
								buf, err := ioutil.ReadAll(f)
								if err != nil {
									return err
								}
								result.Value[k] = &tengo.ImmutableMap{Value: map[string]tengo.Object{
									"file_name": &tengo.String{Value: h.Filename},
									"size":      &tengo.Int{Value: h.Size},
									"header":    headerToTMap(h.Header),
									"data":      &tengo.Bytes{Value: buf},
								}}
								return nil
							}()
							if err != nil {
								return nil, err
							}
						}
					}

					return result, nil
				},
			},
			//snippet: name=all_file;prefix=all_file;description=get all file form request;
			"all_file": &tengo.UserFunction{
				Value: func(args ...tengo.Object) (tengo.Object, error) {
					if r.MultipartForm == nil {
						err := r.ParseMultipartForm(defaultMaxMemory)
						if err != nil {
							return nil, err
						}
					}
					ret := &tengo.ImmutableMap{Value: map[string]tengo.Object{}}
					if r.MultipartForm != nil && r.MultipartForm.File != nil {
						for k, fhs := range r.MultipartForm.File {
							if len(fhs) > 0 {
								fhs0 := fhs[0]
								f, err := fhs0.Open()
								if err != nil {
									return nil, err
								}
								data, err := ioutil.ReadAll(f)
								if err != nil {
									return nil, err
								}
								ret.Value[k] = &tengo.ImmutableMap{Value: map[string]tengo.Object{
									"file_name": &tengo.String{Value: fhs0.Filename},
									"size":      &tengo.Int{Value: fhs0.Size},
									"header":    headerToTMap(fhs0.Header),
									"data":      &tengo.Bytes{Value: data},
								}}

							}
						}
					}
					return ret, nil
				},
			},
			"query": &tengo.UserFunction{
				Name: "query",
				Value: func(args ...tengo.Object) (ret tengo.Object, err error) {
					queries := map[string]tengo.Object{}
					values := r.URL.Query()
					for k, vv := range values {
						if len(vv) == 1 {
							queries[k] = &tengo.String{Value: vv[0]}
						} else {
							var ar []tengo.Object
							for _, v := range vv {
								ar = append(ar, &tengo.String{Value: v})
							}
							queries[k] = &tengo.Array{Value: ar}
						}
					}
					ret = &tengo.Map{Value: queries}
					return
				},
			},
		},
	}
}

func headerToTMap(header map[string][]string) *tengo.ImmutableMap {
	ret := &tengo.ImmutableMap{Value: map[string]tengo.Object{}}
	for k, v := range header {
		ret.Value[k] = &tengo.String{Value: strings.Join(v, ";")}
	}
	return ret
}
