package service

import (
	"BrainBlitz.com/game/pkg/cache_manager"
	"BrainBlitz.com/game/pkg/playerstats"
	"context"
	"fmt"
	"github.com/stretchr/testify/mock"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestStatisticsClient(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/statistics" || r.Header.Get("X-User-ID") != "13" {
			t.Errorf("incorrect identity or path")
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"gamesPlayed":3,"totalPoints":44,"rank":2,"rankedPlayers":5}`))
	}))
	defer server.Close()
	stats, err := NewStatisticsClient(server.URL).PlayerStatistics(context.Background(), "13")
	if err != nil || stats.GamesPlayed != 3 || stats.TotalPoints != 44 || stats.Rank != 2 {
		t.Fatalf("stats=%+v err=%v", stats, err)
	}
}
func TestStatisticsClientUnavailable(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusServiceUnavailable) }))
	defer server.Close()
	if _, err := NewStatisticsClient(server.URL).PlayerStatistics(context.Background(), "13"); err == nil {
		t.Fatal("unavailable statistics must not become zero statistics")
	}
}

type profileStatisticsStub struct{ err error }

func (s profileStatisticsStub) PlayerStatistics(_ context.Context, id string) (playerstats.Statistics, error) {
	if id != "13" {
		return playerstats.Statistics{}, fmt.Errorf("unexpected identity: %s", id)
	}
	return playerstats.Statistics{GamesPlayed: 2, TotalPoints: 94, Rank: 1}, s.err
}
func TestProfileIncludesStatisticsAndSurvivesOutage(t *testing.T) {
	for _, unavailable := range []bool{false, true} {
		t.Run(fmt.Sprint(unavailable), func(t *testing.T) {
			repo := new(MockRepository)
			repo.On("GetUserById", mock.Anything, "13").Return(User{ID: 13, Username: "player"}, nil)
			provider := profileStatisticsStub{}
			if unavailable {
				provider.err = fmt.Errorf("statistics offline")
			}
			svc := NewService(repo, cachemanager.CacheManager{}, nil, MockLogger{}).WithStatistics(provider)
			result, err := svc.Profile(context.Background(), ProfileRequest{ID: "13"})
			if err != nil || result.ID != "13" {
				t.Fatalf("account profile must remain available: %+v %v", result, err)
			}
			if unavailable {
				if result.Statistics != nil {
					t.Fatal("outage must return null stats")
				}
			} else if result.Statistics == nil || result.Statistics.TotalPoints != 94 {
				t.Fatalf("missing stats: %+v", result.Statistics)
			}
		})
	}
}
