package service

import (
	"BrainBlitz.com/game/pkg/playerstats"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type StatisticsProvider interface {
	PlayerStatistics(context.Context, string) (playerstats.Statistics, error)
}

type StatisticsClient struct {
	baseURL string
	client  *http.Client
}

func NewStatisticsClient(baseURL string) StatisticsClient {
	return StatisticsClient{strings.TrimRight(baseURL, "/"), &http.Client{Timeout: 5 * time.Second}}
}
func (s Service) WithStatistics(provider StatisticsProvider) Service {
	s.statistics = provider
	return s
}
func (c StatisticsClient) PlayerStatistics(ctx context.Context, userID string) (playerstats.Statistics, error) {
	var stats playerstats.Statistics
	if c.baseURL == "" {
		return stats, fmt.Errorf("game statistics endpoint not configured")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/api/v1/statistics", nil)
	if err != nil {
		return stats, err
	}
	req.Header.Set("X-User-ID", userID)
	resp, err := c.client.Do(req)
	if err != nil {
		return stats, fmt.Errorf("requesting game statistics: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return stats, fmt.Errorf("game statistics status %d", resp.StatusCode)
	}
	err = json.NewDecoder(io.LimitReader(resp.Body, 64*1024)).Decode(&stats)
	return stats, err
}
