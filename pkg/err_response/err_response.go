package err_response

import (
	"BrainBlitz.com/game/pkg/errors"
	"encoding/json"
	"net/http"
)

func respondWithError(w http.ResponseWriter, err *errors.AppError) {
	w.WriteHeader(err.HTTPStatus)
	error := json.NewEncoder(w).Encode(map[string]string{
		"error":   err.Code,
		"message": err.Message,
	})
}
