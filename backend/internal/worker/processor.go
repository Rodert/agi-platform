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
