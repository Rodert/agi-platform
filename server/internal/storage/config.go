package storage

import (
	"fmt"
	"strings"

	"agi-platform/server/internal/config"
)

func NewFromConfig(cfg config.StorageConfig) (Store, error) {
	provider := strings.ToLower(strings.TrimSpace(cfg.Provider))
	if provider == "" {
		provider = "cos"
	}
	if provider == "local" {
		return NewLocalStore(cfg.LocalRoot), nil
	}
	if provider != "cos" {
		return nil, fmt.Errorf("unsupported storage provider: %s", cfg.Provider)
	}

	cosCfg := COSConfig{
		SecretID:      cfg.COS.SecretID,
		SecretKey:     cfg.COS.SecretKey,
		Bucket:        cfg.COS.Bucket,
		Region:        cfg.COS.Region,
		PublicBaseURL: cfg.COS.PublicBaseURL,
		UploadPrefix:  cfg.COS.UploadPrefix,
	}
	if cosCfg.SecretID == "" || cosCfg.SecretKey == "" || cosCfg.Bucket == "" || cosCfg.Region == "" || cosCfg.PublicBaseURL == "" {
		return nil, fmt.Errorf("cos storage requires secret_id, secret_key, bucket, region and public_base_url")
	}
	store, err := NewCOSStore(cosCfg)
	if err != nil {
		return nil, fmt.Errorf("init cos storage: %w", err)
	}
	return store, nil
}
