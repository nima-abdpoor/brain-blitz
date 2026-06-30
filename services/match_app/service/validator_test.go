package service

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateAddToWaitingListRequest(t *testing.T) {
	tests := []struct {
		name        string
		request     AddToWaitingListRequest
		expectError bool
		errorText   string
	}{
		{
			name: "valid request",
			request: AddToWaitingListRequest{
				UserId:   "123",
				Category: "MUSIC",
			},
			expectError: false,
		},
		{
			name: "missing category",
			request: AddToWaitingListRequest{
				UserId:   "123",
				Category: "",
			},
			expectError: true,
			errorText:   ErrValidationDataRequired,
		},
		{
			name: "invalid category",
			request: AddToWaitingListRequest{
				UserId:   "123",
				Category: "unknown-category",
			},
			expectError: true,
			errorText:   ErrInvalidCategory,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := ValidateAddToWaitingListRequest(test.request)

			if test.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), test.errorText)
			} else {
				fmt.Println("erwaasdf--------->", err)
				assert.NoError(t, err)
			}
		})
	}
}
