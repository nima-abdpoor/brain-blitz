package service

import (
	"BrainBlitz.com/game/adapter/broker"
	"BrainBlitz.com/game/contract/event"
	"BrainBlitz.com/game/contract/match/golang"
	errApp "BrainBlitz.com/game/pkg/err_app"
	"BrainBlitz.com/game/pkg/logger"
	"BrainBlitz.com/game/pkg/matchqueue"
	"BrainBlitz.com/game/pkg/playercount"
	"context"
	"fmt"
	"github.com/oklog/ulid/v2"
	"google.golang.org/protobuf/proto"
	"log/slog"
	"math/rand"
	"sort"
	"strconv"
	"time"
)

type Config struct {
	Players        playercount.Limits `koanf:"players"`
	WaitingTimeout time.Duration      `koanf:"waiting_timeout"`
	LeastPresence  time.Duration      `koanf:"least_presence"`
}

type Repository interface {
	AddToWaitingList(ctx context.Context, category Category, userId string, choices ...matchqueue.Choice) error
	GetWaitingListByCategory(ctx context.Context, category Category) ([]WaitingMember, error)
	RemoveWaitingMember(ctx context.Context, members []WaitingMember) error
}

type Service struct {
	config     Config
	repository Repository
	broker     broker.Broker
	logger     logger.Logger
}

func NewService(repository Repository, config Config, broker broker.Broker, logger logger.Logger) Service {
	config.Players = config.Players.WithDefaults()
	if err := config.Players.Validate(); err != nil {
		panic(err)
	}
	return Service{
		config:     config,
		repository: repository,
		broker:     broker,
		logger:     logger,
	}
}

func (svc Service) AddToWaitingList(ctx context.Context, request AddToWaitingListRequest) (AddToWaitingListResponse, error) {
	const op = "matchMakingHandler.AddToWaitingList"

	err := ValidateAddToWaitingListRequest(request)
	if err != nil {
		return AddToWaitingListResponse{}, errApp.Wrap(op, err, errApp.ErrInvalidInput, map[string]string{
			"message": " failed to add into waiting list",
			"data":    fmt.Sprint(request),
		}, svc.logger)
	}

	players, err := svc.config.Players.Resolve(request.Players)
	if err != nil {
		return AddToWaitingListResponse{}, errApp.Wrap(op, err, errApp.ErrInvalidInput, nil, svc.logger)
	}
	err = svc.repository.AddToWaitingList(ctx, MapToCategory(request.Category), request.UserId, matchqueue.Choice{Players: players})
	if err != nil {
		return AddToWaitingListResponse{}, errApp.Wrap(op, err, errApp.ErrInternal, map[string]string{
			"message": " failed to add into waiting list",
			"data":    fmt.Sprint(request),
		}, svc.logger)
	}
	resp := AddToWaitingListResponse{
		Timeout: svc.config.WaitingTimeout,
	}

	return resp, nil
}

func (svc Service) MatchWaitUsers(ctx context.Context, req MatchWaitedUsersRequest) (MatchWaitedUsersResponse, error) {
	const op = "matchMakingHandler.MatchWaitUsers"
	var rErr error = nil
	var finalUsers []MatchedUsers
	var waitingMembers []WaitingMember
	for _, category := range GetCategories() {
		result, err := svc.repository.GetWaitingListByCategory(ctx, category)
		if err != nil {
			rErr = errApp.Wrap(op, err, errApp.ErrInternal, map[string]string{"message": "failed to get waiting list"}, svc.logger)
			continue
		}
		for _, member := range result {
			member.Category = category
			waitingMembers = append(waitingMembers, member)
		}
	}
	sort.SliceStable(waitingMembers, func(i, j int) bool { return waitingMembers[i].TimeStamp < waitingMembers[j].TimeStamp })
	type pool struct {
		category Category
		seconds  int
		option   string
		players  int
	}
	pools := map[pool][]uint64{}
	seen := map[uint]bool{}
	for _, member := range waitingMembers {
		players := member.Players
		if players == 0 {
			players = 2
		} // Existing Redis entries predate player selection.
		if _, err := svc.config.Players.Resolve(players); err != nil || seen[member.UserId] {
			continue
		}
		seen[member.UserId] = true
		key := pool{member.Category, member.QuestionSeconds, member.TimeOptionID, players}
		pools[key] = append(pools[key], uint64(member.UserId))
		if len(pools[key]) < players {
			continue
		}
		entropy := ulid.Monotonic(rand.New(rand.NewSource(time.Now().UnixNano())), 0)
		id := ulid.MustNew(ulid.Timestamp(time.Now()), entropy)
		finalUsers = append(finalUsers, MatchedUsers{
			Id: id.String(), Category: []Category{member.Category}, QuestionSeconds: member.QuestionSeconds,
			TimeOptionID: member.TimeOptionID, UserId: pools[key],
		})
		delete(pools, key) // A new roster and match ID for each exact-size group.
	}

	if len(finalUsers) > 0 {
		svc.logger.Info(op, "message", "readyUsers for category", "finalUsers for category", fmt.Sprintf("%v", finalUsers))
		svc.publishFinalUsers(finalUsers)
	}
	return MatchWaitedUsersResponse{}, rErr
}

func (svc Service) publishFinalUsers(users []MatchedUsers) {
	const op = "matchMakingHandler.publishFinalUsers"
	matchMakingTopic := event.MATCH_V1_MATCH_USERS

	buff, err := proto.Marshal(MapFromEntityToProtoMessage(users))
	if err != nil {
		svc.logger.Error(op, "message", "error in marshaling match message", err.Error())
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err = svc.broker.Publish(ctx, matchMakingTopic, buff)
	if err != nil {
		svc.logger.Error("error in producing message.", "topic", matchMakingTopic, "error", err)
		return
	}

	var waitingMembers []WaitingMember
	for _, user := range users {
		for _, userid := range user.UserId {
			waitingMembers = append(waitingMembers, WaitingMember{
				UserId:   uint(userid),
				Category: user.Category[0],
			})
		}
	}
	err = svc.repository.RemoveWaitingMember(ctx, waitingMembers)
	if err != nil {
		svc.logger.Error(op, "message", "error in removing waiting members", err.Error())
	}

	svc.logger.Info(op, "message", "publishing message...")
}

func (svc Service) ConsumeJoinMatchQueueRequest(message []byte, ctx context.Context) error {
	const op = "match.ConsumeJoinMatchQueueRequest"

	addToWaitingListRequestProto := &golang.AddToWaitingList{}
	err := proto.Unmarshal(message, addToWaitingListRequestProto)
	if err != nil {
		svc.logger.Error(op, "error in unmarshalling add to waiting list request message", slog.String("error", err.Error()))
		return err
	}

	userId, category := MapFromAddToWaitingListProtoToEntity(addToWaitingListRequestProto)
	if category == CategoryTypeUnknown {
		svc.logger.Error(op, "unknown category received by consumer")
		return fmt.Errorf("unknown category")
	}
	players, err := svc.config.Players.Resolve(int(addToWaitingListRequestProto.GetPlayers()))
	if err != nil {
		return err
	}
	err = svc.repository.AddToWaitingList(ctx, category, strconv.FormatUint(userId, 10), matchqueue.Choice{Seconds: int(addToWaitingListRequestProto.GetQuestionSeconds()), OptionID: addToWaitingListRequestProto.GetTimeOptionId(), Players: players})
	if err != nil {
		svc.logger.Error(op, "error in storing add to waiting list request message", slog.String("error", err.Error()))
		return err
	}

	svc.logger.Info(op, "message", "consumed message successfully", "userId", userId, "category", string(category))

	return nil
}
