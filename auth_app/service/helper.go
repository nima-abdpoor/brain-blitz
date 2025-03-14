package service

import "github.com/golang-jwt/jwt/v5"

func toJWTClaims(data []CreateTokenRequest) jwt.MapClaims {
	claims := jwt.MapClaims{}
	for _, value := range data {
		claims[value.Key] = value.Value
	}
	return claims
}

func toMapData(data map[string]struct{}, claims jwt.MapClaims) []CreateTokenRequest {
	result := make([]CreateTokenRequest, 0)
	for key, _ := range data {
		if value, ok := claims[key].(string); ok {
			result = append(result, CreateTokenRequest{
				Key:   key,
				Value: value,
			})
		}
	}
	return result
}
