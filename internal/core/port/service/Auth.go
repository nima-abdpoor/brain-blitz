package service

// AuthGenerator TODO make this as a separate service (RPC call with Traefic ex)
type AuthGenerator interface {
	CreateAccessToken(data map[string]string) (string, error)
	CreateRefreshToken(data map[string]string) (string, error)
	ValidateToken(data []string, token string) (bool, map[string]interface{}, error)
}
