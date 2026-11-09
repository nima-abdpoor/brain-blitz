package service

import (
	"context"
	"fmt"
	"testing"

	"BrainBlitz.com/game/contract/event"
	matchproto "BrainBlitz.com/game/contract/match/golang"
	"BrainBlitz.com/game/pkg/playercount"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func matchRosters(t *testing.T, members []WaitingMember, config Config) [][]uint64 {
	t.Helper()
	repo, broker := new(MockRepo), new(BrokerMock)
	for _, c := range GetCategories() {
		var list []WaitingMember
		for _, m := range members {
			if m.Category == c {
				list = append(list, m)
			}
		}
		repo.On("GetWaitingListByCategory", mock.Anything, c).Return(list, nil)
	}
	var rosters [][]uint64
	broker.On("Publish", mock.Anything, event.MATCH_V1_MATCH_USERS, mock.Anything).Run(func(args mock.Arguments) {
		var payload matchproto.AllMatchedUsers
		require.NoError(t, proto.Unmarshal(args.Get(2).([]byte), &payload))
		ids := map[string]bool{}
		for _, match := range payload.Users {
			require.NotEmpty(t, match.MatchId)
			require.False(t, ids[match.MatchId], "each game needs its own ID")
			ids[match.MatchId] = true
			rosters = append(rosters, match.UserId)
		}
	}).Return(nil)
	repo.On("RemoveWaitingMember", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		var want []uint
		for _, roster := range rosters {
			for _, id := range roster {
				want = append(want, uint(id))
			}
		}
		var removed []uint
		for _, m := range args.Get(1).([]WaitingMember) {
			removed = append(removed, m.UserId)
		}
		require.ElementsMatch(t, want, removed, "leave unmatched players queued")
	}).Return(nil)
	_, err := NewService(repo, config, broker, MockLogger{}).MatchWaitUsers(context.Background(), MatchWaitedUsersRequest{})
	require.NoError(t, err)
	return rosters
}

func TestMatchExactSizedGamesAndKeepRemainder(t *testing.T) {
	for count := 2; count <= 15; count++ {
		t.Run(fmt.Sprint(count), func(t *testing.T) {
			members := []WaitingMember{}
			for id := 1; id <= 2*count+count-1; id++ {
				members = append(members, WaitingMember{UserId: uint(id), Category: CategoryTypeTech, Players: count, TimeStamp: int64(id)})
			}
			rosters := matchRosters(t, members, Config{})
			require.Len(t, rosters, 2)
			for i, roster := range rosters {
				require.Len(t, roster, count)
				for j, id := range roster {
					require.EqualValues(t, i*count+j+1, id, "oldest players match first")
				}
			}
		})
	}
}

func TestMatchSeparatesSizeCategoryAndTimer(t *testing.T) {
	members := []WaitingMember{}
	var next uint = 1
	for _, c := range GetCategories() {
		for _, count := range []int{2, 3, 4, 15} {
			for _, seconds := range []int{10, 60} {
				for n := 0; n < count; n++ {
					members = append(members, WaitingMember{UserId: next, Category: c, Players: count, QuestionSeconds: seconds, TimeOptionID: fmt.Sprint(seconds), TimeStamp: int64(next)})
					next++
				}
			}
		}
	}
	rosters := matchRosters(t, members, Config{})
	require.Len(t, rosters, 24)
	for _, roster := range rosters {
		first := members[int(roster[0])-1]
		require.Len(t, roster, first.Players)
		for _, id := range roster {
			m := members[int(id)-1]
			require.Equal(t, first.Category, m.Category)
			require.Equal(t, first.QuestionSeconds, m.QuestionSeconds)
			require.Equal(t, first.Players, m.Players)
		}
	}
}

func TestMatchWaitsForFullRosterAndRespectsConfiguredLimits(t *testing.T) {
	members := []WaitingMember{{UserId: 1, Category: CategoryTypeSport, Players: 3}, {UserId: 2, Category: CategoryTypeSport, Players: 3}}
	require.Empty(t, matchRosters(t, members, Config{}))
	for i := 3; i <= 5; i++ {
		members = append(members, WaitingMember{UserId: uint(i), Category: CategoryTypeSport, Players: 2})
	}
	require.Empty(t, matchRosters(t, members, Config{Players: playercount.Limits{Min: 3, Max: 4}}))
	members = append(members, WaitingMember{UserId: 6, Category: CategoryTypeSport, Players: 3})
	require.Equal(t, [][]uint64{{1, 2, 6}}, matchRosters(t, members, Config{Players: playercount.Limits{Min: 3, Max: 4}}))
}

func TestConsumerRejectsUnsupportedPlayerCounts(t *testing.T) {
	for _, count := range []int32{-1, 1, 16} {
		repo := new(MockRepo)
		data, err := proto.Marshal(&matchproto.AddToWaitingList{Category: "SPORT", UserId: 1, Players: count})
		require.NoError(t, err)
		require.Error(t, NewService(repo, Config{}, nil, MockLogger{}).ConsumeJoinMatchQueueRequest(data, context.Background()))
		repo.AssertNotCalled(t, "AddToWaitingList", mock.Anything, mock.Anything, mock.Anything)
	}
}
