package service

import (
	"errors"
	"strings"
	"testing"

	"agi-platform/server/internal/storage"
)

func TestPublicErrorMessageRedactsInfrastructureDetails(t *testing.T) {
	raw := `PUT https://agi-platform-dev-1257142189.cos.ap-guangzhou.myqcloud.com/20260630/videos/video_1782819478688029017/result.mp4: 400 UserNetworkTooSlow(Message: User network is too slow., RequestId: abc, TraceId: def)`
	message := publicErrorMessage(errors.New(raw))
	for _, forbidden := range []string{
		"agi-platform-dev-1257142189",
		"myqcloud.com",
		"20260630/videos",
		"RequestId",
		"TraceId",
	} {
		if strings.Contains(message, forbidden) {
			t.Fatalf("public error leaked %q in %q", forbidden, message)
		}
	}
}

func TestPublicErrorMessageStorageUploadFailed(t *testing.T) {
	if got := publicErrorMessage(storage.ErrUploadFailed); got != storage.ErrUploadFailed.Error() {
		t.Fatalf("unexpected storage error message %q", got)
	}
}
