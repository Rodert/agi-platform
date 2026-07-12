package service

import (
	"time"

	"github.com/javapub/agi-platform-backend/internal/dto"
	"github.com/javapub/agi-platform-backend/internal/model"
	"github.com/javapub/agi-platform-backend/internal/repository"
	"github.com/javapub/agi-platform-backend/pkg/errors"
	"gorm.io/gorm"
)

type StorageConfigService struct {
	storageRepo *repository.StorageConfigRepository
}

func NewStorageConfigService(storageRepo *repository.StorageConfigRepository) *StorageConfigService {
	return &StorageConfigService{
		storageRepo: storageRepo,
	}
}

// GetStorageConfigs 获取所有存储配置
func (s *StorageConfigService) GetStorageConfigs() ([]*dto.StorageConfigResponse, error) {
	configs, err := s.storageRepo.GetStorageConfigs()
	if err != nil {
		return nil, err
	}

	responses := make([]*dto.StorageConfigResponse, len(configs))
	for i, cfg := range configs {
		responses[i] = s.toResponse(cfg)
		// 脱敏 SecretKey
		if cfg.SecretKey != "" {
			responses[i].SecretKey = "***" + cfg.SecretKey[len(cfg.SecretKey)-4:]
		}
	}

	return responses, nil
}

// CreateStorageConfig 创建存储配置
func (s *StorageConfigService) CreateStorageConfig(req *dto.StorageConfigRequest) (*dto.StorageConfigResponse, error) {
	config := &model.StorageConfig{
		Name:      req.Name,
		Type:      req.Type,
		LocalPath: req.LocalPath,
		Endpoint:  req.Endpoint,
		AccessKey: req.AccessKey,
		SecretKey: req.SecretKey,
		Bucket:    req.Bucket,
		Region:    req.Region,
		Domain:    req.Domain,
		IsEnabled: false,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.storageRepo.Create(config); err != nil {
		return nil, err
	}

	return s.toResponse(config), nil
}

// UpdateStorageConfig 更新存储配置
func (s *StorageConfigService) UpdateStorageConfig(id int64, req *dto.StorageConfigRequest) (*dto.StorageConfigResponse, error) {
	config, err := s.storageRepo.FindByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New(errors.ErrCodeNotFound, "配置不存在")
		}
		return nil, err
	}

	config.Name = req.Name
	config.Type = req.Type
	config.LocalPath = req.LocalPath
	config.Endpoint = req.Endpoint
	config.AccessKey = req.AccessKey
	if req.SecretKey != "" {
		config.SecretKey = req.SecretKey
	}
	config.Bucket = req.Bucket
	config.Region = req.Region
	config.Domain = req.Domain
	config.UpdatedAt = time.Now()

	if err := s.storageRepo.Update(config); err != nil {
		return nil, err
	}

	return s.toResponse(config), nil
}

// EnableStorageConfig 启用存储配置
func (s *StorageConfigService) EnableStorageConfig(id int64) error {
	return s.storageRepo.EnableConfig(id)
}

// DeleteStorageConfig 删除存储配置
func (s *StorageConfigService) DeleteStorageConfig(id int64) error {
	config, err := s.storageRepo.FindByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.New(errors.ErrCodeNotFound, "配置不存在")
		}
		return err
	}

	if config.IsEnabled {
		return errors.New(errors.ErrCodeBadRequest, "不能删除启用中的配置")
	}

	return s.storageRepo.Delete(id)
}

// toResponse 转换为响应
func (s *StorageConfigService) toResponse(cfg *model.StorageConfig) *dto.StorageConfigResponse {
	return &dto.StorageConfigResponse{
		ID:        cfg.ID,
		Name:      cfg.Name,
		Type:      cfg.Type,
		LocalPath: cfg.LocalPath,
		Endpoint:  cfg.Endpoint,
		AccessKey: cfg.AccessKey,
		SecretKey: cfg.SecretKey,
		Bucket:    cfg.Bucket,
		Region:    cfg.Region,
		Domain:    cfg.Domain,
		IsEnabled: cfg.IsEnabled,
		CreatedAt: cfg.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt: cfg.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}
