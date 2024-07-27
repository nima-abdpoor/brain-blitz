package matchMakingHandler

import (
	"BrainBlitz.com/game/contract/golang/match"
	entity "BrainBlitz.com/game/entity/game"
	"BrainBlitz.com/game/internal/core/model/request"
	"BrainBlitz.com/game/internal/core/model/response"
	"BrainBlitz.com/game/pkg/richerror"
	"context"
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/thoas/go-funk"
	"google.golang.org/protobuf/proto"
	"log"
	"sort"
	"strconv"
	"time"
)

func (s Service) MatchWaitUsers(ctx context.Context, req *request.MatchWaitedUsersRequest) (response.MatchWaitedUsersResponse, error) {
	const op = "matchMakingHandler.MatchWaitUsers"
	var rErr error = nil
	var readyUsers []entity.MatchedUsers
	var finalUsers []entity.MatchedUsers
	var waitingMembers []entity.WaitingMember
	for _, category := range entity.GetCategories() {
		result, err := s.repo.GetWaitingListByCategory(ctx, category)
		for _, res := range result {
			if presenceRes, err := s.presenceClient.GetPresenceByUserID(ctx, strconv.Itoa(int(res.UserId))); err != nil {
				fmt.Println(op, res, err)
			} else {
				if time.Now().Add(s.config.LeastPresence).UnixMilli() <= presenceRes {
					waitingMembers = append(waitingMembers, entity.WaitingMember{
						UserId:    res.UserId,
						TimeStamp: res.TimeStamp,
						Category:  category,
					})
				}
			}
		}
		if err != nil {
			rErr = richerror.New(op).WithError(err)
		}
	}
	sort.Slice(waitingMembers, func(i, j int) bool {
		return waitingMembers[i].TimeStamp < waitingMembers[j].TimeStamp
	})
	for _, member := range waitingMembers {
		index := funk.IndexOf(readyUsers, func(users entity.MatchedUsers) bool {
			return users.Category.String() == member.Category.String()
		})
		if index != -1 {
			readyUsers[index].UserId = append(readyUsers[index].UserId, uint64(member.UserId))
		} else {
			readyUsers = append(readyUsers, entity.MatchedUsers{
				Category: member.Category,
				UserId:   []uint64{uint64(member.UserId)},
			})
		}
	}
	for _, readyUser := range readyUsers {
		r := len(readyUser.UserId)
		if r%2 != 0 {
			r--
		}
		finalUsers = append(finalUsers, entity.MatchedUsers{
			Category: readyUser.Category,
			UserId:   readyUser.UserId[:r],
		})
		fmt.Println(op, "readyUsers for category:", readyUser)
	}

	// todo remove these users from waiting list
	// todo rpc call to create a match for this users
	if len(finalUsers) >= 2 {
		for _, user := range finalUsers {
			fmt.Println(op, "finalUsers for category:", user)
		}
		s.publishFinalUsers(finalUsers)
	}
	return response.MatchWaitedUsersResponse{}, rErr
}

func (s Service) publishFinalUsers(users []entity.MatchedUsers) {
	const op = "matchMakingHandler.publishFinalUsers"
	matchMakingTopic := "matchMaking_v1_matchUsers"
	buff, err := proto.Marshal(match.MapFromEntityToProtoMessage(users))
	if err != nil {
		//todo update metrics
		//todo put logs
		log.Printf("%s, error in marshaling match message %v", op, err)
	}
	producer := s.publisherBroker.Publish(nil)
	switch producer.(type) {
	case *kafka.Producer:
		{
			p := producer.(*kafka.Producer)
			defer p.Close()
			err := p.Produce(&kafka.Message{
				TopicPartition: kafka.TopicPartition{
					Topic:     &matchMakingTopic,
					Partition: kafka.PartitionAny,
				},
				Value: buff,
			}, nil)
			if err != nil {
				//todo add metrics
				//todo add logger
				log.Printf("error in producing message for topic:%s with error:%v", matchMakingTopic, err)
			} else {
				//todo add metrics
				log.Printf("publishing message... %s %s\n", buff, time.Now())
			}
		}
	default:
		{
			//todo add metrics
			//todo add logger
			log.Printf("Unhandled type of publisherBroker %s", producer)
		}
	}
}
