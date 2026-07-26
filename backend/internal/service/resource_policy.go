package service

import (
	"github.com/javapub/agi-platform-backend/internal/model"
	"github.com/javapub/agi-platform-backend/internal/repository"
	"github.com/javapub/agi-platform-backend/pkg/errors"
	"gorm.io/gorm"
)

var generatedResourceTypes = map[string]struct{}{
	"image":     {},
	"video":     {},
	"thumbnail": {},
}

const generatedAssetRetentionDays = 7

type ResourcePolicyService struct {
	repo *repository.ResourcePolicyRepository
}

func NewResourcePolicyService(repo *repository.ResourcePolicyRepository) *ResourcePolicyService {
	return &ResourcePolicyService{repo: repo}
}

func (s *ResourcePolicyService) List() ([]*model.ResourcePolicy, error) { return s.repo.List() }

func (s *ResourcePolicyService) Update(resourceType string, input *model.ResourcePolicy) (*model.ResourcePolicy, error) {
	policy, err := s.repo.FindByType(resourceType)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New(errors.ErrCodeNotFound, "资源策略不存在")
		}
		return nil, err
	}
	if input.KeyPrefix == "" || input.RetentionDays < 0 || input.CacheMaxAge < 0 || input.MaxSizeMB < 1 {
		return nil, errors.New(errors.ErrCodeBadRequest, "资源策略参数无效")
	}
	if !input.IsPublic {
		return nil, errors.New(errors.ErrCodeBadRequest, "当前资源类型必须开启公开访问")
	}
	if _, temporary := generatedResourceTypes[resourceType]; temporary && input.RetentionDays != generatedAssetRetentionDays {
		return nil, errors.New(errors.ErrCodeBadRequest, "图片、视频及缩略图固定保留 7 天")
	}
	policy.KeyPrefix, policy.RetentionDays, policy.IsPublic, policy.CacheMaxAge, policy.MaxSizeMB = input.KeyPrefix, input.RetentionDays, input.IsPublic, input.CacheMaxAge, input.MaxSizeMB
	if err := s.repo.Update(policy); err != nil {
		return nil, err
	}
	return policy, nil
}
