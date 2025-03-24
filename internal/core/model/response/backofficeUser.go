package response

type ListUserResponse struct {
	Users []User `json:"users"`
}

type User struct {
	ID          string `json:"id"`
	Username    string `json:"username"`
	DisplayName string `json:"displayName"`
	Role        string `json:"role"`
	CreatedAt   uint64 `json:"createdAt"`
	UpdatedAt   uint64 `json:"updatedAt"`
}
