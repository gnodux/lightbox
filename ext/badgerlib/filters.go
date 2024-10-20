package badgerlib

import (
	"bytes"
	"regexp"
)

type FilterFn func(value, filter []byte) (bool, error)

func ReMatch(value, filter []byte) (bool, error) {
	return regexp.Match(string(filter), value)
}
func Contains(value, filter []byte) (bool, error) {
	return bytes.Contains(value, filter), nil
}
func HasPrefix(value, filter []byte) (bool, error) {
	return bytes.HasPrefix(value, filter), nil
}
func HasSuffix(value, filter []byte) (bool, error) {
	return bytes.HasSuffix(value, filter), nil
}

func And(filters ...FilterFn) FilterFn {
	return func(value, filter []byte) (bool, error) {
		for _, f := range filters {
			if ok, err := f(value, filter); err != nil || !ok {
				return false, err
			}
		}
		return true, nil
	}
}
func Or(filters ...FilterFn) FilterFn {
	return func(value, filter []byte) (bool, error) {
		for _, f := range filters {
			if ok, err := f(value, filter); err != nil || ok {
				return true, err
			}
		}
		return false, nil
	}
}
