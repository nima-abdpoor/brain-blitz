package service

import (
	"BrainBlitz.com/game/contract/event"
	errmsg "BrainBlitz.com/game/pkg/err_msg"
	"BrainBlitz.com/game/services/game_app/service"
	"context"
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
	"time"
)

type MockLogger struct{}

func (m MockLogger) Info(msg string, args ...any)  {}
func (m MockLogger) Warn(msg string, args ...any)  {}
func (m MockLogger) Debug(msg string, args ...any) {}
func (m MockLogger) Error(msg string, args ...any) {}

type MockRepo struct {
	mock.Mock
}

func (m *MockRepo) RemoveWaitingMember(ctx context.Context, members []WaitingMember) error {
	args := m.Called(ctx, members)
	return args.Error(0)
}

func (m *MockRepo) AddToWaitingList(ctx context.Context, category Category, userId string) error {
	args := m.Called(ctx, category, userId)
	return args.Error(0)
}

func (m *MockRepo) GetWaitingListByCategory(ctx context.Context, category Category) ([]WaitingMember, error) {
	args := m.Called(ctx, category)
	return args.Get(0).([]WaitingMember), args.Error(1)
}

type BrokerMock struct {
	mock.Mock
}

func (m *BrokerMock) Publish(ctx context.Context, topic string, message []byte) error {
	args := m.Called(ctx, topic, message)
	return args.Error(0)
}

func (m *BrokerMock) Consume(ctx context.Context, topic string, handler func([]byte, context.Context) error) error {
	args := m.Called(ctx, topic, handler)
	return args.Error(0)
}

func TestAddToWaitingList_Success(t *testing.T) {
	mockRepo := new(MockRepo)
	mockBroker := new(BrokerMock)

	config := Config{
		WaitingTimeout: 5 * time.Second,
		LeastPresence:  1 * time.Second,
	}

	svc := NewService(mockRepo, config, mockBroker, MockLogger{})

	req := AddToWaitingListRequest{
		Category: string(service.CategoryTypeMusic),
		UserId:   "123",
	}

	mockRepo.On("AddToWaitingList", mock.Anything, MapToCategory(string(service.CategoryTypeMusic)), "123").Return(nil)

	resp, err := svc.AddToWaitingList(context.Background(), req)

	assert.NoError(t, err)
	assert.Equal(t, config.WaitingTimeout, resp.Timeout)
	mockRepo.AssertExpectations(t)
}

func TestAddToWaitingList_ValidationError(t *testing.T) {
	mockRepo := new(MockRepo)
	mockBroker := new(BrokerMock)
	mockLogger := new(MockLogger)

	svc := NewService(mockRepo, Config{}, mockBroker, mockLogger)

	invalidReq := AddToWaitingListRequest{
		Category: "",
		UserId:   "123",
	}

	_, err := svc.AddToWaitingList(context.Background(), invalidReq)

	assert.Error(t, err)
	assert.Equal(t, err.Error(), errmsg.InvalidInputErrMsg)
}

func TestAddToWaitingList(t *testing.T) {
	mockRepo := new(MockRepo)
	mockBroker := new(BrokerMock)
	mockLogger := new(MockLogger)

	svc := NewService(mockRepo, Config{}, mockBroker, mockLogger)

	t.Run("error-repo-AddToWaitingList", func(t *testing.T) {
		req := AddToWaitingListRequest{
			Category: "music",
			UserId:   "123",
		}

		mockRepo.On("AddToWaitingList", mock.Anything, MapToCategory("music"), "123").Return(errors.New("db error"))

		_, err := svc.AddToWaitingList(context.Background(), req)

		assert.Error(t, err)
		assert.Equal(t, err.Error(), errmsg.InvalidInputErrMsg)
	})
}

func TestMatchWaitUsers_GetWaitingListError(t *testing.T) {
	mockRepo := new(MockRepo)
	mockBroker := new(BrokerMock)
	mockLogger := new(MockLogger)

	svc := NewService(mockRepo, Config{}, mockBroker, mockLogger)

	mockRepo.On("GetWaitingListByCategory", mock.Anything, CategoryTypeMusic).
		Return([]WaitingMember{}, fmt.Errorf("db error"))

	mockRepo.On("GetWaitingListByCategory", mock.Anything, CategoryTypeSport).
		Return([]WaitingMember{}, nil)

	mockRepo.On("GetWaitingListByCategory", mock.Anything, CategoryTypeTech).
		Return([]WaitingMember{}, nil)

	t.Run("error-while-getting-waiting-list", func(t *testing.T) {
		_, err := svc.MatchWaitUsers(context.Background(), MatchWaitedUsersRequest{})

		assert.Error(t, err)
		assert.Equal(t, err.Error(), errmsg.SomeThingWentWrong)
		mockRepo.AssertExpectations(t)
	})
}

func TestMatchWaitUsers_MatchingLogic(t *testing.T) {
	mockRepo := new(MockRepo)
	mockBroker := new(BrokerMock)
	mockLogger := new(MockLogger)

	svc := NewService(mockRepo, Config{}, mockBroker, mockLogger)

	now := time.Now().Unix()
	members := []WaitingMember{
		{UserId: 1, TimeStamp: now, Category: CategoryTypeMusic},
		{UserId: 2, TimeStamp: now + 1, Category: CategoryTypeMusic},
	}
	mockRepo.On("GetWaitingListByCategory", mock.Anything, CategoryTypeMusic).Return(members, nil)

	mockRepo.On("GetWaitingListByCategory", mock.Anything, CategoryTypeSport).Return([]WaitingMember{}, nil)
	mockRepo.On("GetWaitingListByCategory", mock.Anything, CategoryTypeTech).Return([]WaitingMember{}, nil)

	membersThatMustBeRemoved := []WaitingMember{
		{UserId: 1, Category: CategoryTypeMusic},
		{UserId: 2, Category: CategoryTypeMusic},
	}
	mockRepo.On("RemoveWaitingMember", mock.Anything, membersThatMustBeRemoved).Return(nil)
	mockBroker.On("Publish", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("[]uint8")).Return(nil)

	_, err := svc.MatchWaitUsers(context.Background(), MatchWaitedUsersRequest{})
	assert.NoError(t, err)

	mockBroker.AssertCalled(t, "Publish", mock.Anything, event.MATCH_V1_MATCH_USERS, mock.AnythingOfType("[]uint8"))
}

func TestPublishFinalUsers_PublishesMessage(t *testing.T) {
	mockRepo := new(MockRepo)
	mockBroker := new(BrokerMock)
	mockLogger := new(MockLogger)

	svc := NewService(mockRepo, Config{}, mockBroker, mockLogger)

	users := []MatchedUsers{
		{
			Category: []Category{CategoryTypeSport},
			UserId:   []uint64{1, 2},
		},
	}

	membersThatMustBeRemoved := []WaitingMember{
		{UserId: 1, Category: CategoryTypeSport},
		{UserId: 2, Category: CategoryTypeSport},
	}
	mockRepo.On("RemoveWaitingMember", mock.Anything, membersThatMustBeRemoved).Return(nil)

	mockBroker.On("Publish", mock.Anything, event.MATCH_V1_MATCH_USERS, mock.AnythingOfType("[]uint8")).Return(nil)

	svc.publishFinalUsers(users)

	mockBroker.AssertCalled(t, "Publish", mock.Anything, event.MATCH_V1_MATCH_USERS, mock.AnythingOfType("[]uint8"))
}
