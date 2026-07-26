package repository

import (
	"github.com/javapub/agi-platform-backend/internal/model"
	"gorm.io/gorm"
)

type AIProviderAccountRepository struct{ db *gorm.DB }
func NewAIProviderAccountRepository(db *gorm.DB) *AIProviderAccountRepository { return &AIProviderAccountRepository{db: db} }
func (r *AIProviderAccountRepository) List() ([]*model.AIProviderAccount,error) { var rows []*model.AIProviderAccount; return rows,r.db.Preload("ChannelModels").Preload("ChannelModels.Model").Order("priority ASC, provider ASC, id ASC").Find(&rows).Error }
func (r *AIProviderAccountRepository) Find(id int64) (*model.AIProviderAccount,error) { var row model.AIProviderAccount; return &row,r.db.First(&row,id).Error }
func (r *AIProviderAccountRepository) Create(row *model.AIProviderAccount) error { return r.db.Create(row).Error }
func (r *AIProviderAccountRepository) Update(row *model.AIProviderAccount) error { return r.db.Save(row).Error }
func (r *AIProviderAccountRepository) Delete(id int64) error { return r.db.Delete(&model.AIProviderAccount{}, id).Error }
