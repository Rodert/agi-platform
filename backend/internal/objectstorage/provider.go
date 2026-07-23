package objectstorage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/javapub/agi-platform-backend/internal/model"
)

func newProvider(config *model.StorageConfig) (provider, error) {
	switch config.Type {
	case "local":
		return &localProvider{basePath: config.LocalPath}, nil
	case "cloudflare":
		return newR2Provider(config)
	default:
		return nil, fmt.Errorf("存储类型 %s 尚未实现", config.Type)
	}
}

type localProvider struct{ basePath string }

func (p *localProvider) Upload(_ context.Context, input uploadInput) error {
	basePath := p.basePath
	if basePath == "" {
		basePath = "./uploads"
	}
	filename := filepath.Join(basePath, filepath.FromSlash(input.Key))
	if err := os.MkdirAll(filepath.Dir(filename), 0o755); err != nil {
		return err
	}
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = io.Copy(file, input.Body)
	return err
}

func (p *localProvider) Download(_ context.Context, key string) (io.ReadCloser, error) {
	basePath := p.basePath
	if basePath == "" {
		basePath = "./uploads"
	}
	return os.Open(filepath.Join(basePath, filepath.FromSlash(key)))
}

func (p *localProvider) PresignGet(_ context.Context, _ string, _ time.Duration) (string, error) {
	return "", fmt.Errorf("本地私有存储不支持向上游提供临时读取地址")
}
