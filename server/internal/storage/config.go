package storage

import (
	"strings"

	"agi-platform/server/internal/config"
)

func NewFromConfig(cfg config.StorageConfig) Store {
	if strings.EqualFold(cfg.Provider, "local") {
		return NewLocalStore(cfg.LocalRoot)
	}

	cosCfg := COSConfig{
		SecretID:      cfg.COS.SecretID,
		SecretKey:     cfg.COS.SecretKey,
		Bucket:        cfg.COS.Bucket,
		Region:        cfg.COS.Region,
		PublicBaseURL: cfg.COS.PublicBaseURL,
		UploadPrefix:  cfg.COS.UploadPrefix,
	}
	if cosCfg.SecretID != "" && cosCfg.SecretKey != "" && cosCfg.Bucket != "" && cosCfg.Region != "" && cosCfg.PublicBaseURL != "" {
		if store, err := NewCOSStore(cosCfg); err == nil {
			return store
		}
	}
	return NewLocalStore(cfg.LocalRoot)
}
