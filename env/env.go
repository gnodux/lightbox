package env

import (
	"go/types"
	"os"
	"strings"
	"sync"
)

var (
	global  = &Environment{}
	envOnce = &sync.Once{}
)

const (
	Profile = "profile"
)

func Global() *Environment {
	envOnce.Do(parseEnv)
	return global
}
func parseEnv() {
	for _, s := range os.Environ() {
		idx := strings.Index(s, "=")
		if idx > 0 {
			global.Store(s[0:idx], s[idx+1:])
		} else {
			global.Store(s, true)
		}
	}
}

func Set(key any, value any) {
	envOnce.Do(parseEnv)
	global.Store(key, value)
}
func BatchSet(all map[any]any) {
	envOnce.Do(parseEnv)
	for k, v := range all {
		global.Store(k, v)
	}
}
func Get[T any | types.Nil](key any) (T, bool) {
	envOnce.Do(parseEnv)
	var ret T
	if t, ok := global.Load(key); ok {
		ret, ok = t.(T)
		return ret, ok
	}
	return ret, false
}

func GetWithDefault[T any | types.Nil](key any, defaultValue T) T {
	envOnce.Do(parseEnv)
	if v, ok := Get[T](key); ok {
		return v
	} else {
		return defaultValue
	}
}

func All() map[string]interface{} {
	envOnce.Do(parseEnv)
	result := map[string]interface{}{}
	global.Range(func(key, value interface{}) bool {
		result[key.(string)] = value
		return true
	})

	return result
}

func Parse(src string, getter func(any) (string, bool)) (string, error) {
	envOnce.Do(parseEnv)
	var newSrc []rune
	var start int
	var end int
	srcRune := []rune(src)
	for idx := 0; idx < len(srcRune); idx++ {
		r := srcRune[idx]
		if r == '{' {
			start = idx + 1
		} else if r == '}' {
			end = idx
		} else {
			if start == 0 {
				newSrc = append(newSrc, r)
			}
		}
		if end > start {
			envName := src[start:end]
			if envVal, ok := getter(envName); ok {
				newSrc = append(newSrc, ([]rune)(envVal)...)
			} else {
				newSrc = append(newSrc, ([]rune)("!"+envName)...)
			}
			start = 0
			end = 0
		}
	}
	return string(newSrc), nil
}
