package service

import (
	"fmt"
	"github.com/golang-jwt/jwt/v5"
)

type JWTAuthGenerator struct {
	secretKey       []byte
	eXPKey          string
	eXPAccessValue  int64
	eXPRefreshValue int64
}

func NewJWTAuthService(secretKey string, expKey string, accessEXP, refreshEXP int64) JWTAuthGenerator {
	return JWTAuthGenerator{
		secretKey:       []byte(secretKey),
		eXPKey:          expKey,
		eXPAccessValue:  accessEXP,
		eXPRefreshValue: refreshEXP,
	}
}

func (auth JWTAuthGenerator) CreateAccessToken(data map[string]string) (string, error) {
	claims := toJWTClaims(data)
	claims[auth.eXPKey] = auth.eXPAccessValue
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	accessToken, err := token.SignedString(auth.secretKey)
	if err != nil {
		return "", err
	}
	return accessToken, nil
}

func (auth JWTAuthGenerator) CreateRefreshToken(data map[string]string) (string, error) {
	claims := toJWTClaims(data)
	claims[auth.eXPKey] = auth.eXPAccessValue
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	refreshToken, err := token.SignedString(auth.secretKey)
	if err != nil {
		return "", err
	}
	return refreshToken, nil
}

func (auth JWTAuthGenerator) ValidateToken(data []string, tokenString string) (bool, map[string]interface{}, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return auth.secretKey, nil
	})
	if err != nil {
		return false, nil, err
	}
	if claims, ok := token.Claims.(jwt.MapClaims); !ok {
		return false, nil, fmt.Errorf("casting Problem with JWT Claims")
	} else {
		return token.Valid, toMapData(data, claims), nil
	}
}

func toJWTClaims(data map[string]string) jwt.MapClaims {
	claims := jwt.MapClaims{}
	for key, value := range data {
		claims[key] = value
	}
	return claims
}

func toMapData(data []string, claims jwt.MapClaims) map[string]interface{} {
	result := make(map[string]interface{})
	for _, key := range data {
		result[key] = claims[key]
	}
	return result
}

type Claim struct {
	UserId string
	Role   string
}
