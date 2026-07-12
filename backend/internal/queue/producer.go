package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/javapub/agi-platform-backend/pkg/logger"
	"github.com/redis/go-redis/v9"
)

// TaskMessage 任务消息
type TaskMessage struct {
	TaskID    int64                  `json:"task_id"`
	UserID    int64                  `json:"user_id"`
	Type      string                 `json:"type"` // image/video/product
	Prompt    string                 `json:"prompt"`
	ModelName string                 `json:"model_name"`
	Params    map[string]interface{} `json:"params"`
}

// Producer 任务生产者
type Producer struct {
	redis      *redis.Client
	streamName string
}

func NewProducer(rdb *redis.Client, streamName string) *Producer {
	return &Producer{
		redis:      rdb,
		streamName: streamName,
	}
}

// Publish 发布任务到队列
func (p *Producer) Publish(ctx context.Context, msg *TaskMessage) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("序列化任务失败: %w", err)
	}

	// 发布到 Redis Stream
	_, err = p.redis.XAdd(ctx, &redis.XAddArgs{
		Stream: p.streamName,
		Values: map[string]interface{}{
			"task_id":    msg.TaskID,
			"user_id":    msg.UserID,
			"type":       msg.Type,
			"data":       string(data),
			"created_at": time.Now().Unix(),
		},
	}).Result()

	if err != nil {
		return fmt.Errorf("发布任务失败: %w", err)
	}

	logger.Info(fmt.Sprintf("📨 任务已发布到队列: TaskID=%d, Type=%s", msg.TaskID, msg.Type))
	return nil
}
