package json

import (
	"encoding/json"
	"errors"
	"io"
)

func DecodeJson(r io.Reader, v any) error {
	decoder := json.NewDecoder(r)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(v); err != nil {
		if err.Error() == "EOF" {
			return errors.New("EXPECTED JSON")
		}
		return err
	}

	return nil
}
