package middleware

import (
	"BrainBlitz.com/game/internal/core/port/service"
	auth "BrainBlitz.com/game/internal/core/service"
	middlewareConsts "BrainBlitz.com/game/internal/middleware/constants"
	errmsg "BrainBlitz.com/game/pkg/err_msg"
	"errors"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAuth(t *testing.T) {
	e := echo.New()

	createRequest := func(method, target string, headers map[string]string) *http.Request {
		req := httptest.NewRequest(method, target, strings.NewReader(""))
		for key, value := range headers {
			req.Header.Set(key, value)
		}
		return req
	}

	tests := []struct {
		name               string
		idParam            string
		token              string
		validateTokenRes   bool
		validateTokenData  map[string]interface{}
		validateTokenError error
		expectedStatusCode int
		expectedResponse   string
		authGenerator      service.AuthGenerator
		expectUserId       string
		expectRole         string
	}{
		{
			name:               "Invalid Id",
			idParam:            "0",
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   "Invalid Id",
		},
		{
			name:               "Invalid Id",
			idParam:            "",
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   "Invalid Id",
		},
		{
			name:               "Missing Authorization Header",
			idParam:            "123",
			expectedStatusCode: http.StatusForbidden,
			expectedResponse:   errmsg.InvalidAuthentication,
		},
		{
			name:               "Invalid Token",
			idParam:            "123",
			token:              "invalid-token",
			validateTokenError: errors.New("invalid token"),
			expectedStatusCode: http.StatusForbidden,
			expectedResponse:   errmsg.InvalidAuthentication,
		},
		{
			name:               "Valid Token but Invalid User",
			idParam:            "123",
			token:              "valid-token",
			validateTokenRes:   true,
			validateTokenData:  map[string]interface{}{"user": "456", "role": "admin"},
			expectedStatusCode: http.StatusForbidden,
			expectedResponse:   errmsg.AccessDenied,
		},
		{
			name:               "Valid Token and Valid User",
			idParam:            "123",
			token:              "valid-token",
			validateTokenRes:   true,
			validateTokenData:  map[string]interface{}{"user": "123", "role": "admin"},
			expectedStatusCode: http.StatusOK,
			expectedResponse:   "next",
			expectUserId:       "123",
			expectRole:         "admin",
		},
		{
			name:               "Valid Token and Valid User",
			idParam:            "123",
			token:              "valid-token",
			validateTokenRes:   true,
			validateTokenData:  map[string]interface{}{"user": "123", "role": "admin"},
			expectedStatusCode: http.StatusOK,
			expectedResponse:   "next",
			expectUserId:       "123",
			expectRole:         "admin",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := createRequest(http.MethodGet, "/", map[string]string{
				"Authorization": tt.token,
			})
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("id")
			c.SetParamValues(tt.idParam)

			next := func(c echo.Context) error {
				return c.String(http.StatusOK, "next")
			}
			middleware := Auth(getMockedAuthGenerator(tt.validateTokenRes, tt.validateTokenError, tt.validateTokenData))(next)
			err := middleware(c)
			if err != nil {
				assert.Error(t, err)
			}
			assert.Equal(t, tt.expectedStatusCode, rec.Code)
			assert.Contains(t, rec.Body.String(), tt.expectedResponse)
			if tt.expectedStatusCode == http.StatusOK {
				userClaim := c.Get(middlewareConsts.UserId).(auth.Claim)
				assert.Equal(t, tt.expectUserId, userClaim.UserId)
				assert.Equal(t, tt.expectRole, userClaim.Role)
			}
		})
	}
}

func getMockedAuthGenerator(shouldTokenBeValid bool, err error, data map[string]interface{}) service.AuthGenerator {
	return service.NewMockAuthGenerator(
		func(d []string, token string) (bool, map[string]interface{}, error) {
			return shouldTokenBeValid, data, err
		}, func(data map[string]string) (string, error) {
			return "", nil
		}, func(data map[string]string) (string, error) {
			return "", nil
		})
}
