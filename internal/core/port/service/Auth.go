package service

type AuthGenerator interface {
	CreateAccessToken(data map[string]string) (string, error)
	CreateRefreshToken(data map[string]string) (string, error)
	ValidateToken(data []string, token string) (bool, map[string]interface{}, error)
}
