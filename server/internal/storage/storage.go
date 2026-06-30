package storage

import (
	"context"
	"errors"
	"io"
	"mime"
	"path/filepath"
	"strings"
	"time"
)

var ErrUploadFailed = errors.New("文件存储上传失败，请稍后重试")

type Object struct {
	Key       string
	AppURL    string
	PublicURL string
	Provider  string
	MimeType  string
	Size      int64
}

type Store interface {
	Put(ctx context.Context, key string, body io.Reader, size int64, mimeType string) (Object, error)
	PublicURL(key string) (string, bool)
}

func DatedKey(parts ...string) string {
	values := []string{time.Now().Format("20060102")}
	values = append(values, parts...)
	return filepath.ToSlash(filepath.Join(values...))
}

func MimeTypeFromFilename(filename string, fallback string) string {
	if value := mime.TypeByExtension(filepath.Ext(filename)); value != "" {
		return value
	}
	if strings.TrimSpace(fallback) != "" {
		return fallback
	}
	return "application/octet-stream"
}

func CleanAssetKey(key string) (string, bool) {
	clean := filepath.ToSlash(filepath.Clean(strings.TrimPrefix(key, "/")))
	return clean, clean != "." && !strings.HasPrefix(clean, "../") && !filepath.IsAbs(clean)
}
