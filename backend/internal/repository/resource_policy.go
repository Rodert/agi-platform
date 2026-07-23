package repository

import (
	"time"

	"github.com/javapub/agi-platform-backend/internal/model"
	"gorm.io/gorm"
)

type ResourcePolicyRepository struct{ db *gorm.DB }

func NewResourcePolicyRepository(db *gorm.DB) *ResourcePolicyRepository {
	return &ResourcePolicyRepository{db: db}
}

func (r *ResourcePolicyRepository) List() ([]*model.ResourcePolicy, error) {
	var policies []*model.ResourcePolicy
	err := r.db.Order("id ASC").Find(&policies).Error
	return policies, err
}

func (r *ResourcePolicyRepository) FindByType(resourceType string) (*model.ResourcePolicy, error) {
	var policy model.ResourcePolicy
	err := r.db.Where("resource_type = ?", resourceType).First(&policy).Error
	return &policy, err
}

func (r *ResourcePolicyRepository) Update(policy *model.ResourcePolicy) error {
	policy.UpdatedAt = time.Now()
	return r.db.Save(policy).Error
}
