package http

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"BrainBlitz.com/game/pkg/logger"
	"BrainBlitz.com/game/pkg/playercount"
	"BrainBlitz.com/game/services/game_app/service"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
)

func TestGameOptionsReturnsConfiguredInclusiveRange(t *testing.T) {
	svc := service.NewService(service.Config{Players: playercount.Limits{Min: 3, Max: 4}}, nil, nil, nil, nil, logger.SlogAdapter{})
	h := NewHandler(svc, logger.SlogAdapter{})
	recorder := httptest.NewRecorder()
	ctx := echo.New().NewContext(httptest.NewRequest(http.MethodGet, "/api/v1/game-options", nil), recorder)
	require.NoError(t, h.GameOptions(ctx))
	require.Equal(t, http.StatusOK, recorder.Code)
	var options service.GameOptions
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &options))
	require.Equal(t, 3, options.MinPlayers)
	require.Equal(t, 4, options.MaxPlayers)
	require.Equal(t, []int{3, 4}, options.NumberOfPlayers)
}
