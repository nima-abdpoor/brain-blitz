package service

import (
	"BrainBlitz.com/game/adapter/websocket"
	"BrainBlitz.com/game/contract/event"
	errApp "BrainBlitz.com/game/pkg/err_app"
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/labstack/echo/v4"
	"google.golang.org/protobuf/proto"
	"io"
	"log/slog"
	"net"
	"strconv"
	"strings"
)

func (svc Service) ProcessGame(ctx echo.Context, request ProcessGameRequest) (ProcessGameResponse, error) {
	const op = "game.processGame"

	connection, rw, _, err := svc.webSocket.Upgrade(ctx.Request(), ctx.Response())
	if err != nil {
		return ProcessGameResponse{}, errApp.Wrap(op, nil, errApp.ErrInternal, map[string]string{
			"message": "error in initializing websocket",
			"data":    fmt.Sprint(request),
		}, svc.logger)
	}

	id, err := strconv.ParseUint(request.Id, 10, 64)
	if err != nil {
		return ProcessGameResponse{}, errApp.Wrap(op, nil, errApp.ErrInternal, map[string]string{
			"message": "error in converting id to Uint",
			"data":    fmt.Sprint(request),
		}, svc.logger)
	}

	svc.mu.Lock()
	svc.connections[id] = *connection
	svc.mu.Unlock()
	switch svc.getUsersGameStatus(id) {
	case GameStatusInitialized:
		{

		}
	case GameStatusPending:
		{

		}
	case GameStatusCreated:
		{

		}
	case GameStatusStarted:
		{

		}
	case GameStatusFinished, GameStatusUnknown:
		{
			categories := svc.getCategories()
			categoriesByte, err := json.Marshal(categories)
			if err != nil {
				return ProcessGameResponse{}, errApp.Wrap(op, nil, errApp.ErrInternal, map[string]string{
					"message": "error in marshaling json of Categories",
					"data":    fmt.Sprint(request),
				}, svc.logger)
			}
			svc.mu.RLock()
			initialConn := svc.connections[id]
			svc.mu.RUnlock()
			err = svc.webSocket.WriteServerData(initialConn, websocket.OpText, string(categoriesByte))
			if err != nil {
				return ProcessGameResponse{}, errApp.Wrap(op, nil, errApp.ErrInternal, map[string]string{
					"message": "error in returning Categories",
					"data":    fmt.Sprint(request),
				}, svc.logger)
			}
		}
	}

	go func(ctx context.Context, conn *net.Conn, rw *bufio.ReadWriter, userID uint64) {
		defer func() {
			svc.closeConnection(userID, *conn)
			svc.logger.Info("connection closed", "userID", userID)
		}()

		for {
			msg, code, err := svc.webSocket.ReadClientData(rw)
			if err != nil {
				if !errors.Is(err, io.EOF) && !errors.Is(err, net.ErrClosed) {
					svc.logger.Warn("socket read failed", "userID", userID, "error", err)
				}
				break
			}
			err = svc.readMessage(ctx, id, conn, code, msg)
			if err != nil {
				if !errors.Is(err, io.EOF) && !errors.Is(err, net.ErrClosed) {
					svc.logger.Warn("socket read failed", "userID", userID, "error", err)
				}
			}
		}
	}(context.Background(), connection, rw, id)

	return ProcessGameResponse{}, nil
}

func (svc Service) readMessage(ctx context.Context, id uint64, conn *net.Conn, code websocket.OpCode, message string) error {
	op := "game.readMessage"
	svc.logger.Info("received message", "code", code, "message", message)

	var req ProcessGameMessageRequest
	var response ProcessGameMessageResponse
	err := json.Unmarshal([]byte(message), &req)
	if err != nil {
		return err
	}

	switch req.Command {
	case CommandAddToWaitingList:
		{
			svc.logger.Info(op, "adding to waiting list")
			if MapToCategory(req.Category) == CategoryTypeUnknown {
				response = ProcessGameMessageResponse{
					Success: false,
					Event:   Error,
					Message: "invalid category",
				}
				addToWaitingListResponse, err := json.Marshal(response)
				if err != nil {
					return err
				}
				err = svc.webSocket.WriteServerData(*conn, code, string(addToWaitingListResponse))
				if err != nil {
					return err
				}
				return nil
			}
			players, playersErr := svc.config.Players.Resolve(req.NumberOfPlayers)
			if playersErr != nil {
				data, _ := json.Marshal(ProcessGameMessageResponse{Event: Error, Message: playersErr.Error()})
				return svc.webSocket.WriteServerData(*conn, code, string(data))
			}
			choice, choiceErr := svc.resolveTimeOption(ctx, req.TimeOptionID)
			if choiceErr != nil {
				data, _ := json.Marshal(ProcessGameMessageResponse{Event: Error, Message: choiceErr.Error()})
				return svc.webSocket.WriteServerData(*conn, code, string(data))
			}
			choice.Players = players
			svc.mu.Lock()
			name := strings.Join(strings.Fields(req.DisplayName), " ")
			if len([]rune(name)) > 64 {
				name = string([]rune(name)[:64])
			}
			svc.playerNames[id] = name
			svc.mu.Unlock()
			if err := svc.joinQueue(ctx, id, req.Category, choice); err != nil {
				return err
			}
			go svc.saveUsersGameStatus([]uint64{id}, GameStatusInitialized)

			brokerCtx, cancel := context.WithTimeout(context.Background(), svc.config.PublishUserToWaitingListTimeOut)
			defer cancel()
			buff, err := proto.Marshal(MapWaitingListRequestToProtoMessage(id, req.Category, choice))
			if err != nil {
				//todo update metrics
				svc.logger.Error(op, "message", "error in marshaling waiting list request message", err.Error())
			}

			err = svc.broker.Publish(brokerCtx, event.GAME_V1_JOIN_MATCH_QUEUE_REQUESTED, buff)
			if err != nil {
				svc.logger.Error(op, "error in publishing join request message into broker", slog.String("error", err.Error()))
				response = ProcessGameMessageResponse{
					Success: false,
					Event:   Error,
					Message: "internal server error",
				}
				addToWaitingListResponse, err := json.Marshal(response)
				if err != nil {
					return err
				}
				err = svc.webSocket.WriteServerData(*conn, code, string(addToWaitingListResponse))
				if err != nil {
					return err
				}
			}
			response = ProcessGameMessageResponse{
				Success: true,
				Event:   AddedToWaitingList,
				Message: "added to waiting list successfully",
			}
			addToWaitingListResponse, err := json.Marshal(response)
			if err != nil {
				return err
			}

			err = svc.webSocket.WriteServerData(*conn, code, string(addToWaitingListResponse))
			if err != nil {
				return err
			}
		}
	case CommandReady:
		{
			// check user if their status is just initialized
			isGameReady, readyErr := svc.markReady(ctx, req.GameId, id)
			if readyErr != nil {
				data, _ := json.Marshal(ProcessGameMessageResponse{Event: Error, Message: readyErr.Error()})
				return svc.webSocket.WriteServerData(*conn, code, string(data))
			}
			if isGameReady {
				err := svc.startRounds(ctx, req.GameId)
				if err != nil {
					response = ProcessGameMessageResponse{
						Success: false,
						Event:   Error,
						Message: "internal server error",
					}
					readyResponse, err := json.Marshal(response)
					if err != nil {
						return err
					}
					err = svc.webSocket.WriteServerData(*conn, code, string(readyResponse))
					if err != nil {
						return err
					}
					return nil
				}

			}
		}
	case CommandGetCategories:
		{
			fmt.Println("get categories")
		}
	case CommandAnswer:
		{
			mu := svc.roundMutex(req.GameAnswer.GameId)
			mu.Lock()
			defer mu.Unlock()
			playerResponse, allAnswered, err := svc.acceptRoundAnswer(ctx, id, req.GameAnswer)
			if err != nil {
				svc.logger.Error(op, "error in saving answer of a player", slog.String("error", err.Error()))
				response = ProcessGameMessageResponse{
					Success: false,
					Event:   Error,
					Message: err.Error(),
				}
				errorResponse, err := json.Marshal(response)
				if err != nil {
					return err
				}
				err = svc.webSocket.WriteServerData(*conn, code, string(errorResponse))
				if err != nil {
					return err
				}
				return nil
			}

			for i := range playerResponse.PlayerPoint {
				playerResponse.PlayerPoint[i].Answers = nil
			}
			response.Event = AnswerAccepted
			response.Message = "answer accepted"
			response.Success = true
			response.MetaData = ProcessGameMetaDataResponse{
				GameId: req.GameAnswer.GameId,
				Answer: svc.addPlayerNames(playerResponse),
			}
			answerAcceptedJson, err := json.Marshal(response)
			if err != nil {
				return err
			}

			err = svc.webSocket.WriteServerData(*conn, code, string(answerAcceptedJson))
			if err != nil {
				return err
			}
			if allAnswered {
				q, err := svc.repository.GetQuestionsByGameId(ctx, req.GameAnswer.GameId)
				if err != nil {
					return err
				}
				return svc.openRound(ctx, req.GameAnswer.GameId, q.CurrentQuestionIndex+1)
			}
		}
	default:
		{
			svc.logger.Error(op, "invalid command", "command", req.Command, req.MatchId)
			response = ProcessGameMessageResponse{
				Success: false,
				Event:   Error,
				Message: "invalid command",
			}
			addToWaitingListResponse, err := json.Marshal(response)
			if err != nil {
				return err
			}
			err = svc.webSocket.WriteServerData(*conn, code, string(addToWaitingListResponse))
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (svc Service) writeMessage(ids []uint64, msg ProcessGameMessageResponse) error {
	const op = "game.service.writeMessage"
	jsonMsg, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	var deliveryErr error

	for _, id := range ids {
		svc.mu.RLock()
		connection, exists := svc.connections[id]
		svc.mu.RUnlock()
		if !exists {
			deliveryErr = errors.Join(deliveryErr, fmt.Errorf("id: %d not found", id))
			continue
		}
		err = svc.webSocket.WriteServerData(connection, websocket.OpText, string(jsonMsg))
		if err != nil {
			deliveryErr = errors.Join(deliveryErr, err)
			svc.logger.Error(op, "message", "error writing message", slog.String("error", err.Error()))
		}
	}
	return deliveryErr
}

func (svc Service) getCategories() GameInitResponse {
	var categories []string

	for _, category := range GetCategories() {
		categories = append(categories, string(category))
	}

	users := svc.config.Players.Options()

	result := GameInitResponse{
		Categories:      categories,
		NumberOfPlayers: users,
	}

	return result
}
