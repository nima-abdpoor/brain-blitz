package service

type AuthGenerator interface {
	CreateAccessToken(data map[string]string) (string, error)
	CreateRefreshToken(data map[string]string) (string, error)
	ValidateToken(token string) (bool, error)
}
