package cryptlib

import (
	"errors"
	"github.com/speps/go-hashids"
)

func hashIDEncode(salt string, ids interface{}) (string, error) {
	data := hashids.NewData()
	data.Salt = salt
	hashID, err := hashids.NewWithData(data)
	if err != nil {
		return "", err
	}
	switch iid := ids.(type) {
	case []int:
		return hashID.Encode(iid)
	case []int64:
		return hashID.EncodeInt64(iid)
	case string:
		return hashID.EncodeHex(iid)
	default:
		return "", errors.New("except []int|[]int64|string")
	}
}

func decodeHex(salt, input string) (string, error) {
	data := hashids.NewData()
	data.Salt = salt
	hashID, err := hashids.NewWithData(data)
	if err != nil {
		return "", err
	}
	return hashID.DecodeHex(input)
}
func decodeInt(salt, input string) ([]int, error) {
	data := hashids.NewData()
	data.Salt = salt
	hashID, err := hashids.NewWithData(data)
	if err != nil {
		return nil, err
	}
	return hashID.DecodeWithError(input)
}

func decodeInt64(salt, input string) ([]int64, error) {
	data := hashids.NewData()
	data.Salt = salt
	hashID, err := hashids.NewWithData(data)
	if err != nil {
		return nil, err
	}
	return hashID.DecodeInt64WithError(input)
}
