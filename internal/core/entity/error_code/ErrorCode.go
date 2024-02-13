package error_code

import "fmt"

type ErrorCode string
type LocalErrorCode int

// error code
const (
	Success        ErrorCode = "SUCCESS"
	InvalidRequest ErrorCode = "INVALID_REQUEST"
	DuplicateUser  ErrorCode = "DUPLICATE_USER"
	InternalError  ErrorCode = "INTERNAL_ERROR"
)

// error message
const (
	SuccessErrMsg        = "success"
	InternalErrMsg       = "internal error"
	InvalidRequestErrMsg = "invalid request"
)

const (
	BcryptErrorHashingPassword LocalErrorCode = iota + 1
)

func GetLocalErrorCode(code LocalErrorCode) string {
	return fmt.Sprintf("ERROR_CODE:%v", code)
}
