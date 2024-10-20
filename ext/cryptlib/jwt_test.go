package cryptlib

import (
	"encoding/base64"
	"fmt"
	"github.com/golang-jwt/jwt"
	"testing"
)

func TestJWTParseWithBase64Key(t *testing.T) {
	tokenString := "eyJhbGciOiJIUzI1NiJ9.eyJob3RhbElkIjoxMDAwMDAwMDAwMDQsImV4cCI6MTY1MzE1OTc5NiwidXNlcklkIjoxMDAwMDAwMDAwNTUsImlhdCI6MTY1MjU1NDk5NiwiYWNjb3VudCI6ImxuQGxhbm5pYW8uY29tIiwicGxhdGZvcm0iOiJob3RhbCJ9.b-qVZhHdb1SJzHEEdKV7ed8CmfH2L0onmMD6mV2yGy0"
	claims, err := JWTParseWithBase64Key(tokenString, "mySecret")
	fmt.Println(claims, err)
}
func TestJwt(t *testing.T) {
	tokenString := "eyJhbGciOiJIUzI1NiJ9.eyJob3RhbElkIjoxMDAwMDAwMDAwMDQsImV4cCI6MTY1MzE1OTc5NiwidXNlcklkIjoxMDAwMDAwMDAwNTUsImlhdCI6MTY1MjU1NDk5NiwiYWNjb3VudCI6ImxuQGxhbm5pYW8uY29tIiwicGxhdGZvcm0iOiJob3RhbCJ9.b-qVZhHdb1SJzHEEdKV7ed8CmfH2L0onmMD6mV2yGy0"
	claims := jwt.MapClaims{}
	k, _ := base64.StdEncoding.DecodeString("mySecret")

	withClaims, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return k, nil
	})
	if err != nil {
		fmt.Println(err)
		t.FailNow()
		return
	}
	fmt.Println(withClaims.Valid)
	r, _ := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return k, nil
	})

	fmt.Println(r.Valid)
}
