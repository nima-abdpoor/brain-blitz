// Package timeoptions contains database-backed question timer configuration.
package timeoptions

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

type Option struct {
	ID        string `json:"id"`
	Label     string `json:"label"`
	Seconds   int    `json:"seconds"`
	Enabled   bool   `json:"enabled"`
	SortOrder int    `json:"sortOrder"`
}

var validID = regexp.MustCompile(`^[a-zA-Z0-9_-]{1,32}$`)

func (o Option) Validate() error {
	if !validID.MatchString(o.ID) || len(strings.TrimSpace(o.Label)) == 0 || len([]rune(o.Label)) > 50 || o.Seconds < 5 || o.Seconds > 3600 || o.SortOrder < 0 {
		return fmt.Errorf("provide an ID (letters, numbers, _, -), label (1–50 characters), 5–3600 seconds, and a nonnegative sort order")
	}
	return nil
}

type Store struct{ DB *sql.DB }

func (s Store) List(ctx context.Context, all bool) ([]Option, error) {
	rows, err := s.DB.QueryContext(ctx, `SELECT id,label,seconds,enabled,sort_order FROM question_time_options WHERE ($1 OR enabled) ORDER BY sort_order,id`, all)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []Option{}
	for rows.Next() {
		var o Option
		if err := rows.Scan(&o.ID, &o.Label, &o.Seconds, &o.Enabled, &o.SortOrder); err != nil {
			return nil, err
		}
		result = append(result, o)
	}
	return result, rows.Err()
}
func (s Store) Put(ctx context.Context, o Option) error {
	if err := o.Validate(); err != nil {
		return err
	}
	_, err := s.DB.ExecContext(ctx, `INSERT INTO question_time_options(id,label,seconds,enabled,sort_order) VALUES($1,$2,$3,$4,$5) ON CONFLICT(id) DO UPDATE SET label=EXCLUDED.label,seconds=EXCLUDED.seconds,enabled=EXCLUDED.enabled,sort_order=EXCLUDED.sort_order`, o.ID, strings.TrimSpace(o.Label), o.Seconds, o.Enabled, o.SortOrder)
	return err
}

type Provider interface {
	List(context.Context) ([]Option, error)
}
type Client struct {
	URL  string
	HTTP *http.Client
}

func NewClient(url string) Client {
	return Client{strings.TrimRight(url, "/"), &http.Client{Timeout: 4 * time.Second}}
}
func (c Client) List(ctx context.Context) ([]Option, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.URL+"/api/v1/question/time-options", nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("time options status %d", resp.StatusCode)
	}
	var options []Option
	err = json.NewDecoder(io.LimitReader(resp.Body, 64*1024)).Decode(&options)
	return options, err
}
