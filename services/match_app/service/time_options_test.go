package service

import (
	"BrainBlitz.com/game/contract/match/golang"
	"context"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"testing"
)

func TestMatchSeparatesQuestionTimes(t *testing.T) {
	repo := new(MockRepo)
	broker := new(BrokerMock)
	repo.On("GetWaitingListByCategory", mock.Anything, CategoryTypeSport).Return([]WaitingMember{{UserId: 1, Category: CategoryTypeSport, QuestionSeconds: 10, TimeOptionID: "10-seconds"}, {UserId: 2, Category: CategoryTypeSport, QuestionSeconds: 60, TimeOptionID: "1-minute"}, {UserId: 3, Category: CategoryTypeSport, QuestionSeconds: 10, TimeOptionID: "10-seconds"}}, nil)
	for _, c := range []Category{CategoryTypeMusic, CategoryTypeTech} {
		repo.On("GetWaitingListByCategory", mock.Anything, c).Return([]WaitingMember{}, nil)
	}
	repo.On("RemoveWaitingMember", mock.Anything, mock.Anything).Return(nil)
	broker.On("Publish", mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		var matches golang.AllMatchedUsers
		require.NoError(t, proto.Unmarshal(args.Get(2).([]byte), &matches))
		require.Len(t, matches.Users, 1)
		require.Equal(t, []uint64{1, 3}, matches.Users[0].UserId)
		require.EqualValues(t, 10, matches.Users[0].QuestionSeconds)
	}).Return(nil)
	_, err := NewService(repo, Config{}, broker, MockLogger{}).MatchWaitUsers(context.Background(), MatchWaitedUsersRequest{})
	require.NoError(t, err)
	broker.AssertExpectations(t)
}
