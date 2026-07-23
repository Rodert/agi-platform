package service

import (
	"context"

	"github.com/javapub/agi-platform-backend/internal/model"
	"github.com/javapub/agi-platform-backend/internal/queue"
)

// publishTask 发布任务到队列
func (s *CreationService) publishTask(task *model.Task, params map[string]interface{}) error {
	ctx := context.Background()
	msg := &queue.TaskMessage{
		TaskID:           task.ID,
		UserID:           task.UserID,
		Type:             task.Type,
		Prompt:           task.Prompt,
		ModelName:        task.ModelName,
		Params:           params,
		MaxRetryAttempts: task.MaxRetryAttempts,
	}
	return s.queueProducer.Publish(ctx, msg)
}
