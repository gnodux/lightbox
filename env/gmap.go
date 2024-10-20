package env

import (
	"fmt"
	"sync"
)

type Environment struct {
	sync.Map
}

const (
	typeKeyFmt = "<%s>%s"
)

func GetVal[T any](e *Environment, key string) (T, bool) {
	var r T
	if v, ok := e.Get(key); ok {
		r, ok = v.(T)
		return r, ok
	}
	return r, false
}

func (e *Environment) Get(key string) (any, bool) {
	return e.Load(key)
}

func (e *Environment) Set(key string, value any) {
	e.Store(key, value)
}

func (e *Environment) SetTyped(name, typeName string, value any) {
	k := fmt.Sprintf(typeKeyFmt, typeName, name)
	e.Store(k, value)
}
func (e *Environment) GetTyped(name, typeName string) (any, bool) {
	k := fmt.Sprintf(typeKeyFmt, typeName, name)
	return e.Load(k)
}
func (e *Environment) Parse(src string) (string, error) {
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
			if envVal, ok := e.Get(envName); ok {
				newSrc = append(newSrc, ([]rune)(envVal.(string))...)
			} else {
				newSrc = append(newSrc, ([]rune)("!"+envName)...)
			}
			start = 0
			end = 0
		}
	}
	return string(newSrc), nil
}
