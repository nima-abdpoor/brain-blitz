package service

type CreateRefreshTokenRequest struct {
	Data []CreateTokenRequest `json:"data"`
}

type CreateRefreshTokenResponse struct {
	RefreshToken string `json:"refresh_token"`
	ExpireTime   int64  `json:"expire_time"`
}

type CreateAccessTokenRequest struct {
	Data []CreateTokenRequest `json:"data" required:"true"`
}

type CreateAccessTokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpireTime  int64  `json:"expire_time"`
}

type ValidateTokenRequest struct {
	Token string   `json:"token" required:"true"`
	Data  []string `json:"data" required:"true"`
}

type ValidateTokenResponse struct {
	Valid          bool                 `json:"valid"`
	AdditionalData []CreateTokenRequest `json:"data"`
}

type CreateTokenRequest struct {
	Key   string `json:"key" required:"true"`
	Value string `json:"value" required:"true"`
}
