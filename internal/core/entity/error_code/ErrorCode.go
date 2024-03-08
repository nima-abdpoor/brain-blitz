package error_code

import (
	"fmt"
	"net/http"
)

type LocalErrorCode int

// error code
const (
	Success       = http.StatusOK
	BadRequest    = http.StatusBadRequest
	InternalError = http.StatusInternalServerError
)

// error message
const (
	SuccessErrMsg        = "success"
	InternalErrMsg       = "internal error"
	InvalidRequestErrMsg = "invalid request"
	InvalidPasswordMsg   = "invalid password"
)

const (
	BcryptErrorHashingPassword LocalErrorCode = iota + 1
)

func GetLocalErrorCode(code LocalErrorCode) string {
	return fmt.Sprintf("ERROR_CODE:%v", code)
}
