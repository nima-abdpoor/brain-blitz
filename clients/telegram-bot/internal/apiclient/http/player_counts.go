package http

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type PlayerCountsClient struct{ client *Client }

func NewPlayerCountsClient(client *Client) *PlayerCountsClient { return &PlayerCountsClient{client} }

func (c *PlayerCountsClient) PlayerCounts(ctx context.Context, token string) ([]int, error) {
	resp, err := c.client.Do(ctx, http.MethodGet, "/api/v1/game-options", nil, http.Header{"Authorization": []string{"Bearer " + token}})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("game options status %d", resp.StatusCode)
	}
	var options struct {
		Min    int   `json:"minPlayers"`
		Max    int   `json:"maxPlayers"`
		Counts []int `json:"numberOfPlayers"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 64*1024)).Decode(&options); err != nil {
		return nil, err
	}
	if options.Min < 2 || options.Max < options.Min || len(options.Counts) == 0 || options.Max-options.Min != len(options.Counts)-1 {
		return nil, fmt.Errorf("invalid player count options")
	}
	for i, count := range options.Counts {
		if count != options.Min+i {
			return nil, fmt.Errorf("invalid player count options")
		}
	}
	return options.Counts, nil
}
