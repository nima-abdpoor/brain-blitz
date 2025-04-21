package http

import (
	"BrainBlitz.com/game/services/auth_app/service"
	"context"
	"encoding/json"
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type MockLogger struct{}

func (m MockLogger) Info(msg string, args ...any)  {}
func (m MockLogger) Warn(msg string, args ...any)  {}
func (m MockLogger) Debug(msg string, args ...any) {}
func (m MockLogger) Error(msg string, keysAndValues ...interface{}) {
	// no op
}

type mockService struct{}

func (m mockService) ValidateToken(ctx context.Context, req service.ValidateTokenRequest) (service.ValidateTokenResponse, error) {
	return service.ValidateTokenResponse{
		Valid: true,
		AdditionalData: []service.CreateTokenRequest{
			{Key: "id", Value: "123"},
			{Key: "role", Value: "admin"},
		},
	}, nil
}

func generateTestJWT(secret string) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":   "123",
		"role": "admin",
	})
	tokenString, _ := token.SignedString([]byte(secret))
	return tokenString
}

func TestValidateToken_Success(t *testing.T) {
	e := echo.New()

	secretKey := "MOCK_SECRET_KEY"

	reqBody := `{"data": ["id", "role"]}`
	req := httptest.NewRequest(http.MethodPost, "/validate", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(echo.HeaderAuthorization, generateTestJWT(secretKey))

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	svc := service.NewService(service.Config{
		SecretKey: secretKey,
	}, MockLogger{})

	h := NewHandler(svc, MockLogger{})

	err := h.ValidateToken(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var resp service.ValidateTokenResponse
	err = json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.True(t, resp.Valid)

	assert.Equal(t, "123", rec.Header().Get("X-User-ID"))
	assert.Equal(t, "admin", rec.Header().Get("X-User-Role"))
	assert.NotEmpty(t, rec.Header().Get("X-Auth-Data"))
}

func TestValidateToken_MissingAuthHeader(t *testing.T) {
	e := echo.New()

	req := httptest.NewRequest(http.MethodPost, "/validate", strings.NewReader(`{}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	svc := service.NewService(service.Config{}, MockLogger{})

	h := NewHandler(svc, MockLogger{})

	err := h.ValidateToken(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)

	var resp service.ValidateTokenResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.False(t, resp.Valid)
}
