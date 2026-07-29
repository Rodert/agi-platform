package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/javapub/agi-platform-backend/internal/worker"
	"github.com/javapub/agi-platform-backend/pkg/logger"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// Consumer 任务消费者
type Consumer struct {
	redis         *redis.Client
	streamName    string
	consumerGroup string
	consumerName  string
	concurrency   int
	processor     worker.Processor
	stopChan      chan struct{}
}

func NewConsumer(
	rdb *redis.Client,
	streamName string,
	consumerGroup string,
	consumerName string,
	concurrency int,
	processor worker.Processor,
) *Consumer {
	if concurrency < 1 {
		concurrency = 1
	}
	return &Consumer{
		redis:         rdb,
		streamName:    streamName,
		consumerGroup: consumerGroup,
		consumerName:  consumerName,
		concurrency:   concurrency,
		processor:     processor,
		stopChan:      make(chan struct{}),
	}
}

// Start 开始消费
func (c *Consumer) Start(ctx context.Context) error {
	// 创建消费者组（如果不存在）
	err := c.redis.XGroupCreateMkStream(ctx, c.streamName, c.consumerGroup, "0").Err()
	if err != nil && err.Error() != "BUSYGROUP Consumer Group name already exists" {
		return fmt.Errorf("创建消费者组失败: %w", err)
	}

	logger.Info(fmt.Sprintf("🚀 Worker 开始消费队列: %s，并发数: %d", c.streamName, c.concurrency))

	for i := 0; i < c.concurrency; i++ {
		consumerName := fmt.Sprintf("%s-%d", c.consumerName, i+1)
		go c.consumeLoop(ctx, consumerName)
	}

	return nil
}

// Stop 停止消费
func (c *Consumer) Stop() {
	close(c.stopChan)
	logger.Info("🛑 Worker 停止消费")
}

// consumeLoop 消费循环
func (c *Consumer) consumeLoop(ctx context.Context, consumerName string) {
	for {
		select {
		case <-c.stopChan:
			return
		case <-ctx.Done():
			return
		default:
			// 读取消息
			streams, err := c.redis.XReadGroup(ctx, &redis.XReadGroupArgs{
				Group:    c.consumerGroup,
				Consumer: consumerName,
				Streams:  []string{c.streamName, ">"},
				Count:    1,
				Block:    5 * time.Second,
			}).Result()

			if err != nil {
				if err == redis.Nil {
					// 没有新消息，继续等待
					continue
				}
				logger.Error("读取队列失败", zap.Error(err))
				time.Sleep(1 * time.Second)
				continue
			}

			// 处理消息
			for _, stream := range streams {
				for _, message := range stream.Messages {
					c.handleMessage(ctx, message)
				}
			}
		}
	}
}

// handleMessage 处理单个消息
func (c *Consumer) handleMessage(ctx context.Context, msg redis.XMessage) {
	// 解析任务数据
	dataStr, ok := msg.Values["data"].(string)
	if !ok {
		logger.Error("消息格式错误", zap.String("id", msg.ID))
		c.ackMessage(ctx, msg.ID)
		return
	}

	var queueTask TaskMessage
	if err := json.Unmarshal([]byte(dataStr), &queueTask); err != nil {
		logger.Error("解析任务失败", zap.Error(err))
		c.ackMessage(ctx, msg.ID)
		return
	}

	taskMsg := &worker.TaskMessage{TaskID: queueTask.TaskID, UserID: queueTask.UserID, Type: queueTask.Type, Prompt: queueTask.Prompt, ModelName: queueTask.ModelName, Params: queueTask.Params}
	logger.Info(fmt.Sprintf("⚙️  开始处理任务: TaskID=%d, Type=%s, Attempt=%d", taskMsg.TaskID, taskMsg.Type, queueTask.Attempt))

	// 调用处理器
	if err := c.processor.Process(ctx, taskMsg); err != nil {
		logger.Error("处理任务失败",
			zap.Int64("task_id", taskMsg.TaskID),
			zap.Error(err),
		)
		canRetry := true
		if decider, ok := c.processor.(worker.RetryDecider); ok {
			canRetry = decider.ShouldRetry(taskMsg.TaskID)
		}
		if canRetry && c.retryTask(ctx, &queueTask) {
			if processor, ok := c.processor.(worker.RetryProcessor); ok {
				if markErr := processor.MarkRetrying(queueTask.TaskID); markErr != nil {
					logger.Error("记录重试状态失败", zap.Int64("task_id", queueTask.TaskID), zap.Error(markErr))
				}
			}
			logger.Info(fmt.Sprintf("🔁 任务已重新入队: TaskID=%d, Attempt=%d", queueTask.TaskID, queueTask.Attempt))
		}
	} else {
		logger.Info(fmt.Sprintf("✅ 任务处理完成: TaskID=%d", taskMsg.TaskID))
	}

	// 确认消息
	c.ackMessage(ctx, msg.ID)
}

// retryTask uses the task's creation-time retry budget. A retry count of zero
// deliberately preserves a single execution.
func (c *Consumer) retryTask(ctx context.Context, task *TaskMessage) bool {
	if task.Attempt >= task.MaxRetryAttempts {
		return false
	}
	task.Attempt++
	data, err := json.Marshal(task)
	if err != nil {
		logger.Error("序列化重试任务失败", zap.Error(err))
		return false
	}
	if err := c.redis.XAdd(ctx, &redis.XAddArgs{Stream: c.streamName, Values: map[string]interface{}{"task_id": task.TaskID, "user_id": task.UserID, "type": task.Type, "data": string(data), "created_at": time.Now().Unix()}}).Err(); err != nil {
		logger.Error("重新投递任务失败", zap.Int64("task_id", task.TaskID), zap.Error(err))
		return false
	}
	return true
}

// ackMessage 确认消息
func (c *Consumer) ackMessage(ctx context.Context, messageID string) {
	err := c.redis.XAck(ctx, c.streamName, c.consumerGroup, messageID).Err()
	if err != nil {
		logger.Error("确认消息失败", zap.String("message_id", messageID), zap.Error(err))
	}
}
