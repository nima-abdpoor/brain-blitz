package service

import (
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
	claims := getJWTClaims(data)
	claims[auth.eXPKey] = auth.eXPAccessValue
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	accessToken, err := token.SignedString(auth.secretKey)
	if err != nil {
		return "", err
	}
	return accessToken, nil
}

func (auth JWTAuthGenerator) CreateRefreshToken(data map[string]string) (string, error) {
	claims := getJWTClaims(data)
	claims[auth.eXPKey] = auth.eXPAccessValue
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	refreshToken, err := token.SignedString(auth.secretKey)
	if err != nil {
		return "", err
	}
	return refreshToken, nil
}

func (auth JWTAuthGenerator) ValidateToken(tokenString string) (bool, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return auth.secretKey, nil
	})
	if err != nil {
		return false, err
	}
	return token.Valid, nil
}

func getJWTClaims(data map[string]string) jwt.MapClaims {
	claims := jwt.MapClaims{}
	for key, value := range data {
		claims[key] = value
	}
	return claims
}
