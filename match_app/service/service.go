package service

import (
	"BrainBlitz.com/game/adapter/broker"
	"BrainBlitz.com/game/logger"
	errmsg "BrainBlitz.com/game/pkg/err_msg"
	"BrainBlitz.com/game/pkg/richerror"
	"context"
	"github.com/thoas/go-funk"
	"go.uber.org/zap"
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
	// todo we should separate this
	publisherBroker broker.PublisherBroker
}

func NewService(repository Repository, config Config) Service {
	return Service{
		config:     config,
		repository: repository,
	}
}

func (svc Service) AddToWaitingList(ctx context.Context, request AddToWaitingListRequest) (AddToWaitingListResponse, error) {
	const op = "matchMakingHandler.AddToWaitingList"

	err := svc.repository.AddToWaitingList(ctx, MapToCategory(request.Category), request.UserId)
	if err != nil {
		logger.Logger.Named(op).Error("add to waiting list failed", zap.String("request.UserId", request.UserId), zap.Error(err))
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
			return users.Category.String() == member.Category.String()
		})
		if index != -1 {
			readyUsers[index].UserId = append(readyUsers[index].UserId, uint64(member.UserId))
		} else {
			readyUsers = append(readyUsers, MatchedUsers{
				Category: member.Category,
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
		logger.Logger.Named(op).Info("readyUsers for category", zap.Any("readyUsers", readyUser))
	}

	// todo remove these users from waiting list
	// todo rpc call to create a match for this users
	if len(finalUsers) > 0 {
		for _, user := range finalUsers {
			logger.Logger.Named(op).Info("finalUsers for category", zap.Any("user", user))
		}
		svc.publishFinalUsers(finalUsers)
	}
	return MatchWaitedUsersResponse{}, rErr
}

func (svc Service) publishFinalUsers(users []MatchedUsers) {
	//todo implement me
	const op = "matchMakingHandler.publishFinalUsers"
	//matchMakingTopic := "matchMaking_v1_matchUsers"
	//buff, err := proto.Marshal(MapFromEntityToProtoMessage(users))
	//if err != nil {
	//	//todo update metrics
	//	logger.Logger.Named(op).Error("error in marshaling match message", zap.Error(err))
	//}
	//producer := svc.publisherBroker.Publish(nil)
	//switch producer.(type) {
	//case *kafka.Producer:
	//	{
	//		p := producer.(*kafka.Producer)
	//		defer p.Close()
	//		err := p.Produce(&kafka.Message{
	//			TopicPartition: kafka.TopicPartition{
	//				Topic:     &matchMakingTopic,
	//				Partition: kafka.PartitionAny,
	//			},
	//			Value: buff,
	//		}, nil)
	//		if err != nil {
	//			//todo add metrics
	//			logger.Logger.Named(op).Error("error in producing message.", zap.String("topic", matchMakingTopic), zap.Error(err))
	//		} else {
	//			//todo add metrics
	//			logger.Logger.Named(op).Info("publishing message...", zap.String("time", time.Now().String()))
	//		}
	//	}
	//default:
	//	{
	//		//todo add metrics
	//		logger.Logger.Named(op).Error("Unhandled type of publisherBroker", zap.Any("producer", producer))
	//	}
	//}
}
