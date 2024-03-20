package response

type Response struct {
	Data         interface{} `json:"data"`
	Status       bool        `json:"status"`
	ErrorCode    int         `json:"errorCode"`
	ErrorMessage string      `json:"errorMessage"`
}

type SignInResponse struct {
	Username     string `json:"username"`
	DisplayName  string `json:"displayName"`
	CreatedAt    uint64 `json:"createdAt"`
	UpdatedAt    uint64 `json:"updatedAt"`
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}
