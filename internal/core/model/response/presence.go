package response

type UpsertPresenceResponse struct {
}

type GetPresenceResponse struct {
	UserIdToTimestamp map[string]int64
}
