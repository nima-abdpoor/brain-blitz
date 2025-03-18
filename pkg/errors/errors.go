package errApp

import (
	errmsg "BrainBlitz.com/game/pkg/err_msg"
	"errors"
	"google.golang.org/grpc/codes"
	"net/http"
)

type AppError struct {
	OP         string
	Code       string `json:"code"`
	Message    string `json:"message"`
	HTTPStatus int
	GRPCStatus codes.Code
	Data       map[string]string
}

func New(op, code, message string, httpStatus int, grpcStatus codes.Code, data map[string]string) *AppError {
	return &AppError{
		OP:         op,
		Code:       code,
		Message:    message,
		HTTPStatus: httpStatus,
		GRPCStatus: grpcStatus,
		Data:       data,
	}
}

func (e *AppError) Error() string {
	return e.Message
}

func Wrap(op string, err error, appErr *AppError, data map[string]string) *AppError {
	if err == nil {
		return nil
	}
	return &AppError{
		OP:         op,
		Code:       appErr.Code,
		Message:    err.Error(),
		HTTPStatus: appErr.HTTPStatus,
		GRPCStatus: appErr.GRPCStatus,
	}
}

func Normalize(err error) *AppError {
	if err == nil {
		return nil
	}
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr
	}
	return Wrap("NORMALIZED", err, ErrInternal, nil)
}

var (
	ErrNotFound     = New("default", "NOT_FOUND", "Resource not found", http.StatusNotFound, codes.NotFound, nil)
	ErrInternal     = New("default", "INTERNAL_ERROR", "Internal server error", http.StatusInternalServerError, codes.Internal, nil)
	ErrInvalidInput = New("default", "INVALID_INPUT", "Invalid input", http.StatusBadRequest, codes.InvalidArgument, nil)
	ErrUnauthorized = New("default", "UNAUTHORIZED", "Unauthorized access", http.StatusUnauthorized, codes.Unauthenticated, nil)
	ErrInvalidLOGIN = New("default", "INVALID_LOGIN", errmsg.InvalidUserNameOrPasswordErrMsg, http.StatusForbidden, codes.PermissionDenied, nil)
)
