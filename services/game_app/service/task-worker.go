package service

import (
	taskqueue "BrainBlitz.com/game/adapter/task-queue"
	"BrainBlitz.com/game/pkg/logger"
	"context"
)

type TaskWorker struct {
	TaskProcessor taskqueue.TaskProcessor
	service       Service
	logger        logger.Logger
}

func NewTaskWorker(logger logger.Logger, service Service, processor taskqueue.TaskProcessor) TaskWorker {
	return TaskWorker{
		TaskProcessor: processor,
		service:       service,
		logger:        logger,
	}
}

func (tw TaskWorker) Process() {
	op := ""
	err := tw.TaskProcessor.Process(context.Background(), map[string]taskqueue.HandlerFunc{
		"game:completed": tw.service.ProcessGameCompletion,
	})
	if err != nil {
		tw.logger.Error(op, "error in processing the task", "error", err.Error())
	}
}
