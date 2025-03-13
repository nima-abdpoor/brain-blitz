package service

import "github.com/golang-jwt/jwt/v5"

func toJWTClaims(data []CreateTokenRequest) jwt.MapClaims {
	claims := jwt.MapClaims{}
	for _, value := range data {
		claims[value.Key] = value.Value
	}
	return claims
}

func toMapData(data []string, claims jwt.MapClaims) []CreateTokenRequest {
	result := make([]CreateTokenRequest, 0)
	for _, key := range data {
		result = append(result, CreateTokenRequest{
			Key:   key,
			Value: claims[key].(string),
		})
	}
	return result
}
