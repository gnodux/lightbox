package osslib

import (
	"fmt"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/d5/tengo/v2"
	"lightbox/ext/util"
)

const (
	clientTypeName = "oss-client"
)

func clientConstructor(proxy *util.Proxy[*oss.Client]) {
	o := proxy.Value
	proxy.Props = map[string]tengo.Object{
		//snippet:name=oss.bucket;prefix=bucket;body=bucket(${1:bucketName});
		"bucket": &tengo.UserFunction{
			Name: "bucket",
			Value: func(args ...tengo.Object) (ret tengo.Object, err error) {
				if len(args) != 1 {
					err = tengo.ErrWrongNumArguments
					return
				}
				bucketName, ok := tengo.ToString(args[0])
				if !ok {
					err = &tengo.ErrInvalidArgumentType{
						Name:     "bucketName",
						Expected: "string",
						Found:    args[0].TypeName(),
					}
				}
				var bucket *oss.Bucket
				bucket, err = o.Bucket(bucketName)
				if err != nil {
					return tengo.FromInterface(err)
				}
				//ob := &ossBucket{Bucket: bucket}
				//ob.init()
				ob := util.NewProxy(bucket).
					WithConstructor(bucketConstructor).
					WithStringer(func() string {
						return fmt.Sprintf("bucket<%s>", bucket.BucketName)
					}).WithTypeName("oss-bucket")
				return ob, nil
			}},
		//snippet:name=oss.list;prefix=list;body=list();
		"list": &tengo.UserFunction{
			Name: "list",
			Value: func(args ...tengo.Object) (tengo.Object, error) {

				buckets, err := o.ListBuckets()
				if err != nil {
					return tengo.FromInterface(err)
				}
				results := &tengo.ImmutableMap{Value: map[string]tengo.Object{}}
				for _, b := range buckets.Buckets {
					results.Value[b.Name] = &tengo.ImmutableMap{Value: map[string]tengo.Object{
						"name":          &tengo.String{Value: b.Name},
						"storage_class": &tengo.String{Value: b.StorageClass},
						"location":      &tengo.String{Value: b.Location},
						"create_time":   &tengo.Time{Value: b.CreationDate},
					}}
				}
				return results, nil
			},
		},
	}
}

//newOSS
//snippet:name=oss.open;prefix=open;body=open(${1:endpoint},${2:accessKeyId},${3:accessSecret});
func newOSS(args ...tengo.Object) (tengo.Object, error) {
	if len(args) != 3 {
		return nil, tengo.ErrWrongNumArguments
	}
	var (
		endpoint     string
		accessId     string
		accessSecret string
		ok           bool
	)
	endpoint, ok = tengo.ToString(args[0])
	if !ok {
		return nil, &tengo.ErrInvalidArgumentType{
			Name:     "endpoint",
			Expected: "string",
			Found:    args[0].TypeName(),
		}
	}
	accessId, ok = tengo.ToString(args[1])
	if !ok {
		return nil, &tengo.ErrInvalidArgumentType{
			Name:     "accessKey",
			Expected: "string",
			Found:    args[1].TypeName(),
		}
	}
	accessSecret, ok = tengo.ToString(args[2])
	if !ok {
		return nil, &tengo.ErrInvalidArgumentType{
			Name:     "accessSecret",
			Expected: "string",
			Found:    args[2].TypeName(),
		}
	}
	client, err := oss.New(endpoint, accessId, accessSecret)
	if err != nil {
		return tengo.FromInterface(err)
	}
	oc := util.NewProxy(client).WithConstructor(clientConstructor).WithStringer(func() string {
		return fmt.Sprintf("oss-client<%s>", endpoint)
	}).WithTypeName(clientTypeName)
	return oc, nil
}
