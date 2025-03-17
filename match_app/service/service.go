package service

import (
	"BrainBlitz.com/game/adapter/broker"
	errmsg "BrainBlitz.com/game/pkg/err_msg"
	"BrainBlitz.com/game/pkg/richerror"
	"context"
	"fmt"
	"github.com/thoas/go-funk"
	"google.golang.org/protobuf/proto"
	"log/slog"
	"sort"
	"time"
)

type Config struct {
	WaitingTimeout time.Duration `koanf:"waiting_timeout"`
	LeastPresence  time.Duration `koanf:"least_presence"`
}

type Repository interface {
	AddToWaitingList(ctx context.Context, category Category, userId string) error
	GetWaitingListByCategory(ctx context.Context, category Category) ([]WaitingMember, error)
}

type Service struct {
	config     Config
	repository Repository
	broker     broker.Broker
	logger     *slog.Logger
}

func NewService(repository Repository, config Config, broker broker.Broker, logger *slog.Logger) Service {
	return Service{
		config:     config,
		repository: repository,
		broker:     broker,
		logger:     logger,
	}
}

func (svc Service) AddToWaitingList(ctx context.Context, request AddToWaitingListRequest) (AddToWaitingListResponse, error) {
	const op = "matchMakingHandler.AddToWaitingList"

	err := svc.repository.AddToWaitingList(ctx, MapToCategory(request.Category), request.UserId)
	if err != nil {
		svc.logger.WithGroup(op).Error("add to waiting list failed", "request.UserId", request.UserId, "error", err.Error())
		return AddToWaitingListResponse{},
			richerror.New(op).WithKind(richerror.KindUnexpected).WithError(err).WithMessage(errmsg.SomeThingWentWrong)
	}
	resp := AddToWaitingListResponse{
		Timeout: svc.config.WaitingTimeout,
	}

	return resp, nil
}

func (svc Service) MatchWaitUsers(ctx context.Context, req MatchWaitedUsersRequest) (MatchWaitedUsersResponse, error) {
	const op = "matchMakingHandler.MatchWaitUsers"
	var rErr error = nil
	var readyUsers []MatchedUsers
	var finalUsers []MatchedUsers
	var waitingMembers []WaitingMember
	for _, category := range GetCategories() {
		result, err := svc.repository.GetWaitingListByCategory(ctx, category)
		for _, res := range result {
			//todo we should implement presenceClient
			//if presenceRes, err := svc.presenceClient.GetPresenceByUserID(ctx, strconv.Itoa(int(res.UserId))); err != nil {
			//	fmt.Println(op, res, err)
			//} else {
			//	if time.Now().Add(svc.config.LeastPresence).UnixMilli() <= presenceRes {
			waitingMembers = append(waitingMembers, WaitingMember{
				UserId:    res.UserId,
				TimeStamp: res.TimeStamp,
				Category:  category,
			})
			//}
			//}
		}
		if err != nil {
			rErr = richerror.New(op).WithError(err)
		}
	}
	sort.Slice(waitingMembers, func(i, j int) bool {
		return waitingMembers[i].TimeStamp < waitingMembers[j].TimeStamp
	})
	for _, member := range waitingMembers {
		index := funk.IndexOf(readyUsers, func(users MatchedUsers) bool {
			for _, category := range users.Category {
				if category.String() == member.Category.String() {
					return true
				}
			}
			return false
		})
		if index != -1 {
			readyUsers[index].UserId = append(readyUsers[index].UserId, uint64(member.UserId))
		} else {
			readyUsers = append(readyUsers, MatchedUsers{
				Category: []Category{member.Category},
				UserId:   []uint64{uint64(member.UserId)},
			})
		}
	}
	for _, readyUser := range readyUsers {
		r := len(readyUser.UserId)
		if r < 2 {
			continue
		}
		if r%2 != 0 {
			r--
		}
		finalUsers = append(finalUsers, MatchedUsers{
			Category: readyUser.Category,
			UserId:   readyUser.UserId[:r],
		})
		svc.logger.WithGroup(op).Info("readyUsers for category", "readyUsers", readyUser)
	}

	// todo remove these users from waiting list
	// todo rpc call to create a match for this users
	if len(finalUsers) > 0 {
		for _, user := range finalUsers {
			svc.logger.WithGroup(op).Info("finalUsers for category", "user", fmt.Sprintf("%s", user))
		}
		svc.publishFinalUsers(finalUsers)
	}
	return MatchWaitedUsersResponse{}, rErr
}

func (svc Service) publishFinalUsers(users []MatchedUsers) {
	const op = "matchMakingHandler.publishFinalUsers"
	matchMakingTopic := "matchMaking_v1_matchUsers"

	buff, err := proto.Marshal(MapFromEntityToProtoMessage(users))
	if err != nil {
		//todo update metrics
		svc.logger.WithGroup(op).Error("error in marshaling match message", "error", err.Error())
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err = svc.broker.Publish(ctx, matchMakingTopic, buff)
	if err != nil {
		svc.logger.Error("error in producing message.", "topic", matchMakingTopic, "error", err)
	}

	svc.logger.Info(op, "message", "publishing message...")
}
