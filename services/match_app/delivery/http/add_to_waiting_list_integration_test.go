package http

import (
	"BrainBlitz.com/game/services/match_app/service"
	"bytes"
	"context"
	"encoding/json"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"
)

type MockLogger struct{}

func (m MockLogger) Info(msg string, args ...any)  {}
func (m MockLogger) Warn(msg string, args ...any)  {}
func (m MockLogger) Debug(msg string, args ...any) {}
func (m MockLogger) Error(msg string, keysAndValues ...interface{}) {
	// no op
}

type mockRepo struct {
	waitingList map[service.Category][]string
}

func (m *mockRepo) AddToWaitingList(ctx context.Context, category service.Category, userId string) error {
	if m.waitingList == nil {
		m.waitingList = make(map[service.Category][]string)
	}
	m.waitingList[category] = append(m.waitingList[category], userId)
	return nil
}

func (m *mockRepo) GetWaitingListByCategory(ctx context.Context, category service.Category) ([]service.WaitingMember, error) {
	users := m.waitingList[category]
	members := make([]service.WaitingMember, len(users))
	for i, id := range users {
		userId, _ := strconv.Atoi(id)
		members[i] = service.WaitingMember{UserId: uint(userId)}
	}
	return members, nil
}

func TestAddToWaitingListIntegration(t *testing.T) {
	e := echo.New()
	repo := &mockRepo{}
	svc := service.NewService(repo, service.Config{
		WaitingTimeout: 30 * time.Second,
	}, nil, MockLogger{})

	handler := NewHandler(svc, MockLogger{})

	t.Run("success", func(t *testing.T) {
		e.POST("/waiting-list", handler.addToWaitingList)

		body := service.AddToWaitingListRequest{
			UserId:   "1",
			Category: service.Music,
		}
		bodyJSON, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/waiting-list", bytes.NewReader(bodyJSON))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set("X-User-ID", "1")

		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := handler.addToWaitingList(c)
		require.NoError(t, err)

		assert.Equal(t, http.StatusOK, rec.Code)

		var resp service.AddToWaitingListResponse
		err = json.Unmarshal(rec.Body.Bytes(), &resp)
		require.NoError(t, err)

		assert.Equal(t, int64(30_000), resp.Timeout.Milliseconds())
		assert.Contains(t, repo.waitingList[service.CategoryTypeMusic], "1")
	})
}
