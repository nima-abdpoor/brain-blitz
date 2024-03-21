package request

type SignInRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type SignUpRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type ProfileRequest struct {
	ID int64 `uri:"id" binding:"required"`
}
