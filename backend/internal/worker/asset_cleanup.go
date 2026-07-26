package worker

import (
	"context"
	"fmt"
	"time"

	"github.com/javapub/agi-platform-backend/internal/objectstorage"
	"github.com/javapub/agi-platform-backend/internal/repository"
	"github.com/javapub/agi-platform-backend/pkg/logger"
	"go.uber.org/zap"
)

const expiredAssetBatchSize = 100

// AssetCleaner removes generated assets after the retention period recorded at upload time.
// Published assets use an independent policy without an expiry and are never selected here.
type AssetCleaner struct {
	assetRepo *repository.MediaAssetRepository
	storage   *objectstorage.Manager
}

func NewAssetCleaner(assetRepo *repository.MediaAssetRepository, storage *objectstorage.Manager) *AssetCleaner {
	return &AssetCleaner{assetRepo: assetRepo, storage: storage}
}

func (c *AssetCleaner) Run(ctx context.Context) error {
	for {
		assets, err := c.assetRepo.FindExpired(time.Now(), expiredAssetBatchSize)
		if err != nil {
			return err
		}
		for _, asset := range assets {
			if err := c.storage.Delete(ctx, asset); err != nil {
				logger.Error("清理过期生成资源失败", zap.Int64("asset_id", asset.ID), zap.Error(err))
				continue
			}
			if err := c.assetRepo.Delete(asset.ID); err != nil {
				return fmt.Errorf("删除过期资源记录 %d: %w", asset.ID, err)
			}
		}
		if len(assets) < expiredAssetBatchSize {
			return nil
		}
	}
}
