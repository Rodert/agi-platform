package repository

import (
	"github.com/javapub/agi-platform-backend/internal/model"
	"gorm.io/gorm"
)

type AIProviderAccountRepository struct{ db *gorm.DB }
func NewAIProviderAccountRepository(db *gorm.DB) *AIProviderAccountRepository { return &AIProviderAccountRepository{db: db} }
func (r *AIProviderAccountRepository) List() ([]*model.AIProviderAccount,error) { var rows []*model.AIProviderAccount; return rows,r.db.Order("provider,id").Find(&rows).Error }
func (r *AIProviderAccountRepository) Find(id int64) (*model.AIProviderAccount,error) { var row model.AIProviderAccount; return &row,r.db.First(&row,id).Error }
func (r *AIProviderAccountRepository) Create(row *model.AIProviderAccount) error { return r.db.Create(row).Error }
func (r *AIProviderAccountRepository) Update(row *model.AIProviderAccount) error { return r.db.Save(row).Error }
func (r *AIProviderAccountRepository) Delete(id int64) error { return r.db.Transaction(func(tx *gorm.DB) error { if err:=tx.Model(&model.AIModel{}).Where("provider_account_id = ?",id).Update("provider_account_id",nil).Error;err!=nil{return err};return tx.Delete(&model.AIProviderAccount{},id).Error }) }
