package worker

import (
	"context"
)

// TaskMessage 任务消息
type TaskMessage struct {
	TaskID    int64                  `json:"task_id"`
	UserID    int64                  `json:"user_id"`
	Type      string                 `json:"type"`
	Prompt    string                 `json:"prompt"`
	ModelName string                 `json:"model_name"`
	Params    map[string]interface{} `json:"params"`
}

// Processor 任务处理器接口
type Processor interface {
	Process(ctx context.Context, msg *TaskMessage) error
}

// RetryProcessor is optional so queue consumers remain independent of a
// concrete generation implementation while still recording retry state.
type RetryProcessor interface {
	Processor
	MarkRetrying(taskID int64) error
}

// RetryDecider lets a processor suppress automatic retries for requests whose
// upstream submission outcome is unknown and therefore unsafe to repeat.
type RetryDecider interface {
	ShouldRetry(taskID int64) bool
}
