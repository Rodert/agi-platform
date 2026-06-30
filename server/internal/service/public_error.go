package service

import (
	"errors"
	"regexp"
	"strings"

	"agi-platform/server/internal/storage"
)

var (
	urlPattern        = regexp.MustCompile(`https?://[^\s"'，。；;]+`)
	cosHostPattern    = regexp.MustCompile(`[A-Za-z0-9][A-Za-z0-9-]*\.cos\.[A-Za-z0-9-]+\.myqcloud\.com`)
	secretLikePattern = regexp.MustCompile(`(?i)(secret[_-]?id|secret[_-]?key|api[_-]?key|authorization|bearer|token)=?[^\s,，;；]+`)
	requestIDPattern  = regexp.MustCompile(`(?i)(RequestId|TraceId):\s*[A-Za-z0-9+/=_-]+`)
)

func publicErrorMessage(err error) string {
	if err == nil {
		return ""
	}
	if errors.Is(err, storage.ErrUploadFailed) {
		return storage.ErrUploadFailed.Error()
	}
	message := strings.TrimSpace(err.Error())
	if message == "" {
		return "任务处理失败，请稍后重试"
	}
	message = requestIDPattern.ReplaceAllString(message, "")
	message = secretLikePattern.ReplaceAllString(message, "$1=***")
	message = urlPattern.ReplaceAllString(message, "[已隐藏URL]")
	message = cosHostPattern.ReplaceAllString(message, "[已隐藏存储域名]")
	message = strings.Join(strings.Fields(message), " ")
	if strings.TrimSpace(message) == "" {
		return "任务处理失败，请稍后重试"
	}
	if len(message) > 300 {
		return message[:300] + "..."
	}
	return message
}
