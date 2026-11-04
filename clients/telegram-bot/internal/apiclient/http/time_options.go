package http

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/nima-abdpoor/brain-blitz-telegram-bot/internal/core/matchmaking"
	"io"
	"net/http"
)

type TimeOptionsClient struct{ client *Client }

func NewTimeOptionsClient(client *Client) *TimeOptionsClient { return &TimeOptionsClient{client} }
func (c *TimeOptionsClient) TimeOptions(ctx context.Context, token string) ([]matchmaking.TimeOption, error) {
	resp, err := c.client.Do(ctx, http.MethodGet, "/api/v1/question/time-options", nil, http.Header{"Authorization": []string{"Bearer " + token}})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("question time options status %d", resp.StatusCode)
	}
	var options []matchmaking.TimeOption
	if err := json.NewDecoder(io.LimitReader(resp.Body, 64*1024)).Decode(&options); err != nil {
		return nil, err
	}
	result := []matchmaking.TimeOption{}
	for _, o := range options {
		if o.Enabled && o.Seconds >= 5 && o.Seconds <= 3600 && o.ID != "" {
			result = append(result, o)
		}
	}
	return result, nil
}
