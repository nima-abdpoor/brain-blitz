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

type HTTPErrMessage struct {
	Message string `json:"message"`
	Error   string `json:"error"`
}

type GRPCErrMessage struct {
	Message string `json:"message"`
	Error   string `json:"error"`
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
		err = errors.New(appErr.Message)
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

func ToHTTPJson(err error) (message interface{}, code int) {
	var appErr *AppError
	if errors.As(err, &appErr) {
		msg := HTTPErrMessage{
			Message: appErr.Message,
			Error:   appErr.Message,
		}
		return msg, appErr.HTTPStatus
	}
	return err.Error(), http.StatusInternalServerError
}

func ToGRPCJson(err error) (message string, code codes.Code) {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Message, appErr.GRPCStatus
	}
	return err.Error(), codes.Internal
}

var (
	ErrNotFound     = New("default", "NOT_FOUND", "Resource not found", http.StatusNotFound, codes.NotFound, nil)
	ErrInternal     = New("default", "INTERNAL_ERROR", "Internal server error", http.StatusInternalServerError, codes.Internal, nil)
	ErrInvalidInput = New("default", "INVALID_INPUT", "Invalid input", http.StatusBadRequest, codes.InvalidArgument, nil)
	ErrUnauthorized = New("default", "UNAUTHORIZED", "Unauthorized access", http.StatusUnauthorized, codes.Unauthenticated, nil)
	ErrInvalidLOGIN = New("default", "INVALID_LOGIN", errmsg.InvalidUserNameOrPasswordErrMsg, http.StatusForbidden, codes.PermissionDenied, nil)
)
