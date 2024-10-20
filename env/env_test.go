package env

import (
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
	"io"
	"os"
	"strings"
	"testing"
)

func TestBatchSet(t *testing.T) {
	type args struct {
		all map[string]interface{}
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "test1",
			args: args{all: map[string]interface{}{
				"a": 1,
				"b": 2,
			}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
		})
	}
}
func TestJson(t *testing.T) {
	jsonData := `
{
	"appkey":"this is a app key",
	"appscret":"this is a secret",
	"values":{
	"employee":1,
	"orgid":2
}
}
`
	var bc interface{}
	if v, ok := bc.(string); ok {
		fmt.Println(v)
	} else {
		fmt.Println("error")
	}
	m := map[string]interface{}{}
	err := json.Unmarshal([]byte(jsonData), &m)
	fmt.Println(err)
	fmt.Println(m)

}

func TestLogus(t *testing.T) {
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetLevel(logrus.DebugLevel)
	logrus.SetOutput(io.MultiWriter(os.Stdout, &lumberjack.Logger{Filename: "./log10.log", MaxSize: 1}))
	logrus.Info("a", "b")
	logrus.Error("tenant1", "b")
	for i := 0; i < 10000; i++ {
		logrus.WithField("sandbox", "{abc}").Error("123\n12", strings.Repeat("this is a long long string", 2000))
	}
	logrus.Debug("this is a debug info")
	logrus.SetOutput(os.Stderr)
}
