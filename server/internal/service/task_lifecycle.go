package service

import "time"

const (
	taskTimeoutDuration = 24 * time.Hour
	taskTimeoutMessage  = "生成超时 24 小时"
)

func taskTimeoutCutoff(now time.Time) time.Time {
	return now.Add(-taskTimeoutDuration)
}

func taskCreatedAtExpired(createdAt time.Time, now time.Time) bool {
	return !createdAt.IsZero() && createdAt.Before(taskTimeoutCutoff(now))
}
