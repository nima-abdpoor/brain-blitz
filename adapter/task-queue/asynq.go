package task_queue

import (
	"BrainBlitz.com/game/pkg/logger"
	"context"
	"encoding/json"
	"fmt"
	"github.com/hibiken/asynq"
	"time"
)

type PublisherConfig struct {
	RedisConfig RedisConfig `koanf:"redis_config"`
}

type WorkerConfig struct {
	RedisConfig RedisConfig `koanf:"redis_config"`
}

type RedisConfig struct {
	Host string `koanf:"host"`
	Port string `koanf:"port"`
}

type AsynqTaskQueuePublisher struct {
	Logger logger.Logger
	Config PublisherConfig
	client *asynq.Client
}

type AsynqTaskQueueWorker struct {
	Logger logger.Logger
	Config WorkerConfig
	server *asynq.Server
}

func NewPublisherAsynq(logger logger.Logger, config PublisherConfig) AsynqTaskQueuePublisher {
	client := asynq.NewClient(asynq.RedisClientOpt{Addr: fmt.Sprintf("%s:%s", config.RedisConfig.Host, config.RedisConfig.Port)})

	return AsynqTaskQueuePublisher{
		Logger: logger,
		Config: config,
		client: client,
	}
}

func NewWorkerAsynq(logger logger.Logger, config WorkerConfig) AsynqTaskQueueWorker {
	server := asynq.NewServer(
		asynq.RedisClientOpt{Addr: fmt.Sprintf("%s:%s", config.RedisConfig.Host, config.RedisConfig.Port)},
		asynq.Config{
			Concurrency: 10,
			Queues: map[string]int{
				"default":  6,
				"critical": 3,
				"low":      1,
			},
		},
	)

	return AsynqTaskQueueWorker{
		Logger: logger,
		Config: config,
		server: server,
	}
}

func (a *AsynqTaskQueuePublisher) Publish(ctx context.Context, taskType string, payload any, options ...Option) (error, string) {
	data, err := json.Marshal(payload)
	if err != nil {
		return err, ""
	}

	var asynqOptions []asynq.Option

	for _, opt := range options {
		switch opt.Type() {
		case MaxRetryOpt:
			{
				asynqOptions = append(asynqOptions, asynq.MaxRetry(opt.Value().(int)))
			}
		case ProcessAtOpt:
			{
				asynqOptions = append(asynqOptions, asynq.ProcessAt(opt.Value().(time.Time)))
			}
		case ProcessInOpt:
			{
				asynqOptions = append(asynqOptions, asynq.ProcessIn(opt.Value().(time.Duration)))
			}
		default:
			return fmt.Errorf("unknow option type"), ""
		}
	}

	task := asynq.NewTask(taskType, data, asynqOptions...)
	info, err := a.client.EnqueueContext(ctx, task)
	if err != nil {
		return err, ""
	}

	return nil, info.ID
}

func (a *AsynqTaskQueueWorker) Process(ctx context.Context, handlers map[string]HandlerFunc) error {
	mux := asynq.NewServeMux()

	for taskType, handler := range handlers {
		mux.HandleFunc(taskType, func(c context.Context, t *asynq.Task) error {
			var payload map[string]interface{}
			if err := json.Unmarshal(t.Payload(), &payload); err != nil {
				return err
			}
			return handler(c, payload)
		})
	}

	return a.server.Run(mux)
}
