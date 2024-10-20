package modman

import (
	"fmt"
	"github.com/d5/tengo/v2"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/singleflight"
	"io/ioutil"
	"net/http"
	"strings"
)

var urlGroup = &singleflight.Group{}

func ImportUrl(url string) tengo.Importable {
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return nil
	}
	result, err, _ := urlGroup.Do(url, func() (interface{}, error) {
		response, err := http.Get(url)
		if err != nil {
			return nil, err
		}
		if response.StatusCode == 200 {
			all, err := ioutil.ReadAll(response.Body)
			if err != nil {
				return nil, err
			}
			return all, nil
		} else {
			all, err := ioutil.ReadAll(response.Body)
			if err != nil {
				return nil, err
			}
			return nil, fmt.Errorf("request failed with %s:%s", response.Status, string(all))
		}
	})
	if err != nil {
		log.Error("request url failed", err)
		return nil
	}
	if result != nil {
		if script, ok := result.([]byte); ok {

			return &tengo.SourceModule{Src: script}
		}
	}
	return nil
}
