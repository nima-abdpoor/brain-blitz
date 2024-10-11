package notification

import (
	"BrainBlitz.com/game/adapter/broker"
	"BrainBlitz.com/game/contract/golang/game"
	"BrainBlitz.com/game/internal/core/model/request"
	"BrainBlitz.com/game/internal/core/port/service"
	"BrainBlitz.com/game/logger"
	"BrainBlitz.com/game/pkg/richerror"
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
	"io"
	"net"
	"strconv"
	"time"
)

type Service struct {
	config         Config
	consumerBroker broker.ConsumerBroker
	connections    IdToConnection
}

type IdToConnection map[uint64]net.Conn

type Config struct {
}

func New(config Config, consumerBroker broker.ConsumerBroker, connections IdToConnection) service.Notification {
	return Service{
		config:         config,
		consumerBroker: consumerBroker,
		connections:    connections,
	}
}

func (s Service) StartNotifyMatchCreation(req request.StartNotifyMatchCreationRequest) (request.StartNotifyMatchCreationResponse, error) {
	op := "notification.StartNotifyMatchCreation"
	matchCreationTopic := "matchCreated_v1_matchId"
	matchCreatedGroup := "matchMaking"
	consumer, _ := s.consumerBroker.Consume(map[string]string{
		"group.id":          matchCreatedGroup,
		"auto.offset.reset": "smallest",
	})

	switch c := consumer.(type) {
	case *kafka.Consumer:
		{
			run := true
			defer c.Close()
			err := c.SubscribeTopics([]string{matchCreationTopic}, nil)
			if err != nil {
				logger.Logger.Named(op).Error("Error in subscribing to topics", zap.String("topic", matchCreationTopic), zap.Error(err))
			}

			for run == true {
				ev := c.Poll(100)
				switch e := ev.(type) {
				case *kafka.Message:
					gameInfo := &game.GameCreationInfo{}
					err := proto.Unmarshal(e.Value, gameInfo)
					if err != nil {
						//todo update metrics
						logger.Logger.Named(op).Error("Error in unmarshalling message", zap.Error(err))
					}
					// send gameId to the user
					logger.Logger.Named(op).Info(
						"consumer received message",
						zap.String("message", fmt.Sprint(gameInfo)),
						zap.String("time", time.Now().String()))

					//todo we can wrap gameInfo id around json Object
					err = writeMessage(s.connections, gameInfo.UserId, gameInfo.Id)
					if err != nil {
						logger.Logger.Named(op).Error("Error in writing message to client", zap.Error(err))
					}

				case kafka.Error:
					logger.Logger.Named(op).Error("Error in consuming message", zap.Error(e))
					run = false
				default:
					//logger.Logger.Named(op).Error("Unknown message type", zap.Any("message", e))
				}
			}
		}
	default:
		{
			//todo add metrics
			logger.Logger.Named(op).Error("Unhandled type of consumerBroker", zap.Any("consumer", consumer))
		}
	}
	return request.StartNotifyMatchCreationResponse{}, nil
}

func (s Service) InitGame(ctx echo.Context, req *request.InitGameRequest) (request.InitGameResponse, error) {
	const op = "notification.InitGame"

	connection, _, _, err := ws.UpgradeHTTP(ctx.Request(), ctx.Response())
	if err != nil {
		logger.Logger.Named(op).Error("error in initializing websocket", zap.Error(err))
		return request.InitGameResponse{}, richerror.New(op).WithKind(richerror.KindUnexpected).WithError(err)
	}
	if id, err := strconv.ParseUint(req.Id, 10, 64); err != nil {
		logger.Logger.Named(op).Error("error in converting id to Uint", zap.Error(err), zap.String("id", req.Id))
		return request.InitGameResponse{}, err
	} else {
		s.connections[id] = connection
		fmt.Println("idToConnections:", s.connections)
		return request.InitGameResponse{}, nil
	}
}

func readMessage(rw io.ReadWriter) (message string, opCode int, err error) {
	op := "notification.readMessage"
	for {
		msg, opCode, err := wsutil.ReadClientData(rw)
		if err != nil {
			logger.Logger.Named(op).Error("error in readingMessage websocket", zap.Error(err))
		}

		fmt.Println("message:", string(msg), "opCODE:", opCode)
	}
}

func writeMessage(connections IdToConnection, ids []uint64, msg string) error {
	const op = "notification.writeMessage"

	for _, id := range ids {
		if connection, exists := connections[id]; !exists {
			return richerror.New(op).WithMessage(fmt.Sprintf("id: %s not found", id))
		} else {
			err := wsutil.WriteServerMessage(connection, ws.OpText, []byte(msg))
			if err != nil {
				logger.Logger.Named(op).Error("Error in writing message to client", zap.Error(err))
			}
		}
	}
	return nil
}
