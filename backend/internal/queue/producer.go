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
	TaskID           int64                  `json:"task_id"`
	UserID           int64                  `json:"user_id"`
	Type             string                 `json:"type"` // image/video/product
	Prompt           string                 `json:"prompt"`
	ModelName        string                 `json:"model_name"`
	Params           map[string]interface{} `json:"params"`
	Attempt          int                    `json:"attempt"`
	MaxRetryAttempts int                    `json:"max_retry_attempts"`
}

// Producer 任务生产者
type Producer struct {
	redis        *redis.Client
	imageStream  string
	videoStream  string
	defaultStream string
}

func NewProducer(rdb *redis.Client, imageStream, videoStream, defaultStream string) *Producer {
	return &Producer{redis: rdb, imageStream: imageStream, videoStream: videoStream, defaultStream: defaultStream}
}

// Publish 发布任务到队列
func (p *Producer) Publish(ctx context.Context, msg *TaskMessage) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("序列化任务失败: %w", err)
	}

	streamName := p.streamFor(msg.Type)
	_, err = p.redis.XAdd(ctx, &redis.XAddArgs{
		Stream: streamName,
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

	logger.Info(fmt.Sprintf("📨 任务已发布到队列: TaskID=%d, Type=%s, Stream=%s", msg.TaskID, msg.Type, streamName))
	return nil
}

func (p *Producer) streamFor(taskType string) string {
	if taskType == "video" && p.videoStream != "" {
		return p.videoStream
	}
	if taskType != "video" && p.imageStream != "" {
		return p.imageStream
	}
	return p.defaultStream
}
