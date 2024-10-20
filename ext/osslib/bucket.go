package osslib

import (
	"bytes"
	"errors"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/d5/tengo/v2"
	"io/ioutil"
	"lightbox/ext/util"
	"strings"
)

const (
	bucketTypeName = "oss-bucket"
)

func bucketConstructor(bucket *util.Proxy[*oss.Bucket]) {
	b := bucket.Value
	bucket.Props = map[string]tengo.Object{
		//snippet:name=bucket.list({prefix:...,max:...});body=list({$1});prefix=list;
		"list": &tengo.UserFunction{
			Name: "list",
			Value: func(args ...tengo.Object) (tengo.Object, error) {
				var options []oss.Option
				marker := oss.Marker("")
				if len(args) == 1 {
					argMap, ok := args[0].(*tengo.Map)
					if !ok {
						return nil, &tengo.ErrInvalidArgumentType{
							Name:     "arg",
							Expected: "map",
							Found:    args[0].TypeName(),
						}
					}
					for k, v := range argMap.Value {
						switch k {
						case "prefix":
							p, _ := tengo.ToString(v)
							options = append(options, oss.Prefix(p))
						case "marker":
							p, _ := tengo.ToString(v)
							marker = oss.Marker(p)
						case "max":
							p, _ := tengo.ToInt(v)
							options = append(options, oss.MaxKeys(p))
						}
					}
				}
				var files []tengo.Object
				for {
					lor, err := b.ListObjects(append(options, marker)...)
					if err != nil {
						return tengo.FromInterface(err)
					}
					marker = oss.Marker(lor.NextMarker)
					//fmt.Println("next marker",lor.NextMarker)

					for _, f := range lor.Objects {
						files = append(files, &tengo.ImmutableMap{Value: map[string]tengo.Object{
							"name":          &tengo.String{Value: f.Key},
							"storage_class": &tengo.String{Value: f.StorageClass},
							"size":          &tengo.Int{Value: f.Size},
							"type":          &tengo.String{Value: f.Type},
							"etag":          &tengo.String{Value: f.ETag},
							"last_modify":   &tengo.Time{Value: f.LastModified},
						}})
					}
					if !lor.IsTruncated {
						break
					}
				}
				return &tengo.Array{Value: files}, nil
			}},
		//snippet:name=bucket.get(key);body=get($1);prefix=get;
		"get": &tengo.UserFunction{
			Name: "get",
			Value: func(args ...tengo.Object) (tengo.Object, error) {
				if len(args) != 1 {
					return nil, tengo.ErrWrongNumArguments
				}
				key, ok := tengo.ToString(args[0])
				if !ok {
					return tengo.FromInterface(errors.New("key not a string"))
				}

				object, err := b.GetObject(key)
				if err != nil {
					return tengo.FromInterface(err)
				}

				all, err := ioutil.ReadAll(object)
				if err != nil {
					return tengo.FromInterface(err)
				}
				return &tengo.Bytes{Value: all}, nil
			}},
		//snippet:name=bucket.get_meta(key);body=get_meta($1);prefix=get_meta;
		"get_meta": &tengo.UserFunction{
			Name: "get_meta",
			Value: func(args ...tengo.Object) (tengo.Object, error) {
				if len(args) != 1 {
					return nil, tengo.ErrWrongNumArguments
				}
				key, ok := tengo.ToString(args[0])
				if !ok {
					return tengo.FromInterface(errors.New("key not a string"))
				}

				object, err := b.GetObjectDetailedMeta(key)
				if err != nil {
					return tengo.FromInterface(err)
				}
				metaValue := map[string]tengo.Object{}
				for k, v := range object {
					vv, e := tengo.FromInterface(strings.Join(v, ";"))
					if e != nil {
						return tengo.FromInterface(e)
					}
					metaValue[k] = vv
				}
				return &tengo.ImmutableMap{Value: metaValue}, nil
			}},
		//snippet:name=bucket.download;prefix=download;body=download(${1:key},${2:localfile});
		"download": &tengo.UserFunction{
			Name: "download", Value: func(args ...tengo.Object) (tengo.Object, error) {
				if len(args) != 2 {
					return nil, tengo.ErrWrongNumArguments
				}
				key, ok := tengo.ToString(args[0])
				if !ok {
					return tengo.FromInterface(errors.New("key not a string"))
				}
				target, ok := tengo.ToString(args[1])
				if !ok {
					return tengo.FromInterface(errors.New("target not a string"))
				}

				object, err := b.GetObject(key)
				if err != nil {
					return tengo.FromInterface(err)
				}

				all, err := ioutil.ReadAll(object)
				if err != nil {
					return tengo.FromInterface(err)
				}
				err = ioutil.WriteFile(target, all, 0644)
				if err != nil {
					return tengo.FromInterface(err)
				}
				return tengo.TrueValue, nil
			}},
		//snippet:name=bucket.put;prefix=put;body=put(${1:key},${2:body});
		"put": &tengo.UserFunction{
			Name: "put",
			Value: func(args ...tengo.Object) (tengo.Object, error) {
				if len(args) != 2 {
					return nil, tengo.ErrWrongNumArguments
				}
				key, ok := tengo.ToString(args[0])
				if !ok {
					return tengo.FromInterface(errors.New("key not a string"))
				}
				var (
					err error
				)
				switch a := args[1].(type) {
				case *tengo.String:
					err = b.PutObjectFromFile(key, a.Value)
					if err != nil {
						return tengo.FromInterface(err)
					}
				case *tengo.Bytes:
					err = b.PutObject(key, bytes.NewBuffer(a.Value))
					if err != nil {
						return tengo.FromInterface(err)
					}
				default:
					return tengo.FromInterface(errors.New("unknown payload type"))
				}
				return tengo.TrueValue, nil
			},
		},
		//snippet:name=bucket.del(...keys);prefix=del;body=del($1);
		"del": &tengo.UserFunction{
			Name: "del",
			Value: func(args ...tengo.Object) (tengo.Object, error) {
				var keys []string
				for _, arg := range args {
					k, ok := tengo.ToString(arg)
					if !ok {
						return tengo.FromInterface(errors.New("key is not a string"))
					}
					keys = append(keys, k)
				}
				deleteObjects, err := b.DeleteObjects(keys)
				if err != nil {
					return tengo.FromInterface(err)
				}
				var results []tengo.Object
				for _, delObj := range deleteObjects.DeletedObjects {
					results = append(results, &tengo.String{Value: delObj})
				}
				return &tengo.ImmutableArray{Value: results}, nil
			},
		},
	}
}
