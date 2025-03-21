package service

import (
	"fmt"
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

var (
	ErrValidationDataRequired = "category is required"
	ErrInvalidCategory        = "invalid category"
)

func ValidateAddToWaitingListRequest(req AddToWaitingListRequest) error {
	err := validation.ValidateStruct(&req,
		validation.Field(
			&req.Category,
			validation.Required.Error(ErrValidationDataRequired),
		),
	)

	if err != nil {
		return err
	}

	if MapToCategory(req.Category) == 0 {
		return fmt.Errorf(ErrInvalidCategory)
	}

	return nil
}
