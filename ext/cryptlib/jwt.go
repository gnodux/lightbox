package cryptlib

import (
	"encoding/base64"
	"github.com/golang-jwt/jwt"
)

func JWTParseWithBase64Key(tokenString string, key string) (map[string]interface{}, error) {
	claims := jwt.MapClaims{}
	_, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return base64.StdEncoding.DecodeString(key)
	})
	return claims, err
}
func JWTSigning(method, key string, claims jwt.Claims) (string, error) {
	k, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		return "", err
	}
	var signingMethod jwt.SigningMethod
	switch method {
	default:
		signingMethod = jwt.SigningMethodHS256
	}
	token := jwt.NewWithClaims(signingMethod, claims)
	return token.SignedString(k)
}
