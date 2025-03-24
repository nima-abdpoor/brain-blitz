package request

type UpsertPresenceRequest struct {
	UserID string
}

type GetPresenceRequest struct {
	UserID []string
}
