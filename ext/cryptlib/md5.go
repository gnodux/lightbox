package cryptlib

import (
	"crypto/md5"
	"github.com/d5/tengo/v2"
	"io"
	"lightbox/ext/util"
	"os"
	"strings"
)

func sumMD5(args ...tengo.Object) (tengo.Object, error) {
	if len(args) != 1 {
		return util.Error(tengo.ErrWrongNumArguments), nil
	}
	var data []byte
	switch v := args[0].(type) {
	case *tengo.String:
		if strings.HasPrefix(v.Value, "@") {
			if sumValue, err := sumFileMD5(v.Value[1:]); err != nil {
				return util.Error(err), nil
			} else {
				return &tengo.Bytes{Value: sumValue}, nil
			}
		}
		data = []byte(v.Value)
	case *tengo.Bytes:
		data = v.Value
	default:
		return util.Error(&tengo.ErrInvalidArgumentType{Name: "argo", Expected: "String/bytes", Found: args[0].TypeName()}), nil
	}
	sumValue := md5.Sum(data)
	return &tengo.Bytes{Value: sumValue[:]}, nil
}
func sumFileMD5(f string) ([]byte, error) {

	fd, err := os.Open(f)
	if err != nil {
		return nil, err
	}
	defer fd.Close()
	m := md5.New()
	buf := make([]byte, 1024)
	for {
		n, err := fd.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return nil, err
			}
		}
		m.Write(buf[0:n])
	}
	return m.Sum(nil), nil
}
