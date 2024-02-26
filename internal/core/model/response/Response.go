package response

type Response struct {
	Data         interface{} `json:"data"`
	Status       bool        `json:"status"`
	ErrorCode    int         `json:"errorCode"`
	ErrorMessage string      `json:"errorMessage"`
}
