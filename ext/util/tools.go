package util

import (
	"github.com/d5/tengo/v2"
)

const (
	LegoAnnotation = "//lgo:(%s) (%s)"
)

func MergeMap[K comparable, V any](maps ...map[K]V) map[K]V {
	newMap := make(map[K]V)
	for _, m := range maps {
		if m != nil {
			for k, v := range m {
				newMap[k] = v
			}
		}
	}
	return newMap
}

func ToArray[T any](values ...T) (*tengo.Array, error) {
	objs, err := ToObjectSlice(values...)
	if err != nil {
		return nil, err
	}
	return &tengo.Array{Value: objs}, nil
}

func ToImmutableArray[T any](values ...T) (*tengo.ImmutableArray, error) {
	objs, err := ToObjectSlice(values...)
	if err != nil {
		return nil, err
	}
	return &tengo.ImmutableArray{Value: objs}, nil
}

func ToObjectSlice[T any](values ...T) ([]tengo.Object, error) {
	var ret = make([]tengo.Object, len(values))
	for idx, v := range values {
		v, err := tengo.FromInterface(v)
		if err != nil {
			return nil, err
		}
		ret[idx] = v
	}
	return ret, nil
}

func ToSlice[T any](array []tengo.Object) []T {
	result := make([]T, len(array))
	for idx, v := range array {
		result[idx] = tengo.ToInterface(v).(T)
	}
	return result
}
func ToMap[T any](m map[string]tengo.Object) map[string]T {
	var nullVal T
	if m == nil {
		return nil
	}
	result := make(map[string]T, len(m))
	for k, v := range m {
		vv := tengo.ToInterface(v)
		if vv == nil {
			result[k] = nullVal
		} else {
			result[k] = vv.(T)
		}
	}
	return result
}

//func ToObjectSlice[T any](input []T) ([]tengo.Object, error) {
//	var result []tengo.Object
//	if input == nil {
//		return result, nil
//	}
//	result = make([]tengo.Object, len(input))
//	for idx, v := range input {
//		if value, err := tengo.FromInterface(v); err != nil {
//			return nil, err
//		} else {
//			result[idx] = value
//		}
//	}
//	return result, nil
//}
