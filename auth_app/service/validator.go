package service

import (
	"fmt"
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

var (
	ErrValidationDataRequired = "data is required"
	ErrValidationRequired     = "invalid data"
)

func ValidateCreateAccessTokenRequest(req CreateAccessTokenRequest) error {
	return validation.ValidateStruct(&req,
		validation.Field(
			&req.Data,
			validation.Required.Error(ErrValidationDataRequired),
			validation.Each(validation.By(func(value interface{}) error {
				dataItem, ok := value.(CreateTokenRequest)
				if !ok {
					return fmt.Errorf("invalid data structure")
				}
				return validation.ValidateStruct(&dataItem,
					validation.Field(&dataItem.Key, validation.Required.Error(ErrValidationRequired)),
					validation.Field(&dataItem.Value, validation.Required.Error(ErrValidationRequired)),
				)
			})),
		),
	)
}

func ValidateCreateRefreshTokenRequest(req CreateRefreshTokenRequest) error {
	return validation.ValidateStruct(&req,
		validation.Field(
			&req.Data,
			validation.Required.Error(ErrValidationDataRequired),
			validation.Each(validation.By(func(value interface{}) error {
				dataItem, ok := value.(CreateTokenRequest)
				if !ok {
					return fmt.Errorf("invalid data structure")
				}
				return validation.ValidateStruct(&dataItem,
					validation.Field(&dataItem.Key, validation.Required.Error(ErrValidationRequired)),
					validation.Field(&dataItem.Value, validation.Required.Error(ErrValidationRequired)),
				)
			})),
		),
	)
}

func ValidateValidateTokenRequest(req ValidateTokenRequest) error {
	return validation.ValidateStruct(&req,
		validation.Field(
			&req.Token,
			validation.Required.Error(ErrValidationDataRequired),
		),
	)
}
