package contract

import (
	"errors"
	"fmt"
)

func NotNil(value interface{}, msg string) CheckFn {
	return func() error {
		if value == nil {
			return fmt.Errorf(msg)
		}
		return nil
	}
}

func NotEmptyString(v string, msg string) CheckFn {
	return func() error {
		if len(v) == 0 {
			return errors.New(msg)
		}
		return nil
	}

}
func NotEmptySlice[T any](values []T, msg string) CheckFn {
	return func() error {
		if len(values) == 0 {
			return errors.New(msg)
		}
		return nil
	}
}
func NotEqual[T comparable](v1, v2 T, msg string) CheckFn {
	return func() error {
		if v1 != v2 {
			return errors.New(msg)
		}
		return nil
	}
}

func NotExistsInMap[K comparable, V any](m map[K]V, key K, msg string) CheckFn {
	return func() error {
		if _, ok := m[key]; !ok {
			return errors.New(msg)
		}
		return nil
	}
}

type Number interface {
	int8 | int16 | int32 | int64 | int | uint8 | uint16 | uint32 | uint64 | uint | float32 | float64
}

func NotBetween[T Number](v T, start T, end T, msg string) CheckFn {
	return func() error {
		if (start < v) && (v < end) {
			return nil
		}
		return errors.New(msg)
	}
}

func NotGreaterThan[T Number](v1, v2 T, msg string) CheckFn {
	return func() error {
		if v1 > v2 {
			return nil
		}
		return errors.New(msg)
	}
}
