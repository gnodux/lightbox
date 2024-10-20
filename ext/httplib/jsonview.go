package httplib

import (
	"encoding/json"
	"io/ioutil"
	"strings"
)

type JsonConverter struct {
}

func (j *JsonConverter) Read(context *HttpContext) (interface{}, error) {
	var out map[string]interface{}
	buf, err := ioutil.ReadAll(context.Request.Body)
	if err != nil {
		return nil, err
	}
	if buf == nil || len(buf) == 0 {
		return nil, nil
	}
	err = json.Unmarshal(buf, &out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (j *JsonConverter) Write(context *HttpContext, data interface{}) error {
	var (
		bytes []byte
		err   error
		ok    bool
	)
	if bytes, ok = data.([]byte); !ok {
		bytes, err = json.Marshal(data)
		if err != nil {
			return err
		}
	}
	context.Response.Header().Set("Content-Type", "application/json")
	_, err = context.Response.Write(bytes)
	return err
}

func (j *JsonConverter) CanWrite(data interface{}, mediaType string) bool {
	return strings.Contains(mediaType, "application/json")
}

func (j *JsonConverter) CanRead(mediaType string, context *HttpContext) bool {
	return strings.Contains(mediaType, "application/json")
}
