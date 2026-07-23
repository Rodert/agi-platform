package repository

import (
	"time"

	"github.com/javapub/agi-platform-backend/internal/model"
	"gorm.io/gorm"
)

type AnnouncementRepository struct{ db *gorm.DB }

func NewAnnouncementRepository(db *gorm.DB) *AnnouncementRepository {
	return &AnnouncementRepository{db: db}
}

func (r *AnnouncementRepository) ListPublished(page, pageSize int) ([]*model.Announcement, int64, error) {
	query := r.db.Model(&model.Announcement{}).Where("is_published = ? AND (published_at IS NULL OR published_at <= ?)", true, time.Now())
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var items []*model.Announcement
	err := query.Order("published_at DESC, id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&items).Error
	return items, total, err
}

func (r *AnnouncementRepository) ListAdmin(page, pageSize int) ([]*model.Announcement, int64, error) {
	query := r.db.Model(&model.Announcement{})
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var items []*model.Announcement
	err := query.Order("created_at DESC, id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&items).Error
	return items, total, err
}

func (r *AnnouncementRepository) FindByID(id int64) (*model.Announcement, error) {
	var item model.Announcement
	if err := r.db.First(&item, id).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *AnnouncementRepository) Save(item *model.Announcement) error { return r.db.Save(item).Error }
func (r *AnnouncementRepository) Create(item *model.Announcement) error {
	return r.db.Create(item).Error
}
func (r *AnnouncementRepository) Delete(id int64) error {
	return r.db.Delete(&model.Announcement{}, id).Error
}
