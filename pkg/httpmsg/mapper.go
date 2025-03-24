package httpmsg

import (
	errmsg "BrainBlitz.com/game/pkg/err_msg"
	"BrainBlitz.com/game/pkg/richerror"
	"net/http"
)

func Error(err error) (message string, code int) {
	switch err.(type) {
	case richerror.RichError:
		re := err.(richerror.RichError)
		msg := re.Message()
		code := mapKindToHTTPStatusCode(re.Kind())
		if code >= 500 {
			msg = errmsg.SomeThingWentWrong
		}
		return msg, code
	default:
		return err.Error(), http.StatusBadRequest
	}
}

func mapKindToHTTPStatusCode(kind richerror.Kind) int {
	switch kind {
	case richerror.KindInvalid:
		return http.StatusBadRequest
	case richerror.KindForbidden:
		return http.StatusForbidden
	case richerror.KindNotFound:
		return http.StatusNotFound
	case richerror.KindUnexpected:
		return http.StatusInternalServerError
	default:
		return http.StatusBadRequest
	}
}
