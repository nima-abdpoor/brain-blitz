package auth_adapter

type CreateAccessTokenRequest struct {
	Data []CreateTokenRequest `json:"data" required:"true"`
}

type CreateRefreshTokenRequest struct {
	Data []CreateTokenRequest `json:"data" required:"true"`
}

type CreateRefreshTokenResponse struct {
	RefreshToken string `json:"refresh_token"`
	ExpireTime   int64  `json:"expire_time"`
}

type CreateAccessTokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpireTime  int64  `json:"expire_time"`
}

type CreateTokenRequest struct {
	Key   string `json:"key" required:"true"`
	Value string `json:"value" required:"true"`
}
