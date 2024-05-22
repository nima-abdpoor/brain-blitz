package matchMakingHandler

import (
	entity "BrainBlitz.com/game/entity/game"
	"BrainBlitz.com/game/internal/core/model/request"
	"BrainBlitz.com/game/internal/core/model/response"
	"BrainBlitz.com/game/pkg/richerror"
	"context"
	"fmt"
	"github.com/thoas/go-funk"
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
			readyUsers[index].UserId = append(readyUsers[index].UserId, member.UserId)
		} else {
			readyUsers = append(readyUsers, entity.MatchedUsers{
				Category: member.Category,
				UserId:   []uint{member.UserId},
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

	// todo remove this users from waiting list
	// todo rpc call to create a match for this users
	for _, user := range finalUsers {
		fmt.Println(op, "finalUsers for category:", user)
	}
	return response.MatchWaitedUsersResponse{}, rErr
}
