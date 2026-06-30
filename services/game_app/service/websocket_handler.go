package service

import (
	taskqueue "BrainBlitz.com/game/adapter/task-queue"
	"BrainBlitz.com/game/adapter/websocket"
	"BrainBlitz.com/game/contract/event"
	errApp "BrainBlitz.com/game/pkg/err_app"
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"github.com/labstack/echo/v4"
	"google.golang.org/protobuf/proto"
	"log/slog"
	"net"
	"strconv"
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
	case GameStatusFinished:
		{
			//err = svc.saveUsersGameStatus(id, GameStatusInitialized)
			//if err != nil {
			//	svc.logger.Error(op, "error in storing users GameStatus", slog.String("error", err.Error()))
			//}
		}
	case GameStatusUnknown:
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
			(*conn).Close()
			svc.mu.Lock()
			delete(svc.connections, userID)
			svc.mu.Unlock()
			svc.logger.Info("connection closed", "userID", userID)
		}()

		for {
			msg, code, err := svc.webSocket.ReadClientData(rw)
			if err != nil {
				svc.logger.Error("read failed", "userID", userID, "error", err)
				break
			}
			err = svc.readMessage(ctx, id, conn, code, msg)
			if err != nil {
				svc.logger.Error("read failed", "userID", userID, "error", err)
			}
		}
	}(ctx.Request().Context(), connection, rw, id)

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
			}
			go svc.saveUsersGameStatus([]uint64{id}, GameStatusInitialized)

			brokerCtx, cancel := context.WithTimeout(context.Background(), svc.config.PublishUserToWaitingListTimeOut)
			defer cancel()
			buff, err := proto.Marshal(MapWaitingListRequestToProtoMessage(id, req.Category))
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
			isGameReady := svc.saveGameStatus(req.GameId, &id, nil)
			if isGameReady {
				finalTTL, err := svc.repository.SetValidAnswerTimeForQuestions(context.Background(), req.GameId)
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
				}

				err = svc.sendQuestionToPlayer(ctx, req.GameId)

				taskId, err := svc.taskPublisher.Publish(context.Background(), "game:completed", struct {
					GameId string `json:"gameId"`
				}{
					GameId: req.GameId,
				}, taskqueue.MaxRetry(3), taskqueue.ProcessIn(finalTTL))
				if err != nil {
					svc.logger.Error(op, "error in publishing task", "error", err.Error())
				}

				svc.logger.Info(op, "game completion task published", "taskId", taskId, "gameId", req.GameId)
			}
		}
	case CommandGetCategories:
		{
			fmt.Println("get categories")
		}
	case CommandAnswer:
		{
			playerResponse, err := svc.savePlayerAnswer(ctx, id, req.GameAnswer)
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

			response.Event = AnswerAccepted
			response.Message = "answer accepted"
			response.Success = true
			response.MetaData = ProcessGameMetaDataResponse{
				GameId: req.GameId,
				Answer: playerResponse,
			}
			answerAcceptedJson, err := json.Marshal(response)
			if err != nil {
				return err
			}

			err = svc.webSocket.WriteServerData(*conn, code, string(answerAcceptedJson))
			if err != nil {
				return err
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

	for _, id := range ids {
		svc.mu.RLock()
		connection, exists := svc.connections[id]
		svc.mu.RUnlock()
		if !exists {
			return fmt.Errorf("id: %d not found", id)
		}
		jsonMsg, err := json.Marshal(msg)
		if err != nil {
			svc.logger.Error(op, "message", "error writing message", slog.String("error", err.Error()))
		}
		err = svc.webSocket.WriteServerData(connection, websocket.OpText, string(jsonMsg))
		if err != nil {
			svc.logger.Error(op, "message", "error writing message", slog.String("error", err.Error()))
		}
	}
	return nil
}

func (svc Service) getCategories() GameInitResponse {
	var categories []string

	for _, category := range GetCategories() {
		categories = append(categories, string(category))
	}

	users := []int{2}

	result := GameInitResponse{
		Categories:      categories,
		NumberOfPlayers: users,
	}

	return result
}
