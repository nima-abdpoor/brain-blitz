package request

type StartMatchCreatorRequest struct {
}

type StartMatchCreatorResponse struct {
}

type PublishMatchCreatedRequest struct {
	UserId  []uint64
	MatchId string
}

type PublishMatchCreatedResponse struct {
}
