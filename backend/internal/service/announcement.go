package service

import (
	"strings"
	"time"

	"github.com/javapub/agi-platform-backend/internal/model"
	"github.com/javapub/agi-platform-backend/internal/repository"
	"github.com/javapub/agi-platform-backend/pkg/errors"
	"gorm.io/gorm"
)

type AnnouncementService struct {
	repo *repository.AnnouncementRepository
}

func NewAnnouncementService(repo *repository.AnnouncementRepository) *AnnouncementService {
	return &AnnouncementService{repo: repo}
}

type AnnouncementInput struct {
	Title       string `json:"title" binding:"required,max=120"`
	Content     string `json:"content" binding:"required,max=5000"`
	Category    string `json:"category"`
	IsPublished bool   `json:"is_published"`
}

func (s *AnnouncementService) ListPublished(page, pageSize int) ([]*model.Announcement, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return s.repo.ListPublished(page, pageSize)
}
func (s *AnnouncementService) ListAdmin(page, pageSize int) ([]*model.Announcement, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return s.repo.ListAdmin(page, pageSize)
}
func (s *AnnouncementService) Create(input *AnnouncementInput) (*model.Announcement, error) {
	item, err := announcementFromInput(nil, input)
	if err != nil {
		return nil, err
	}
	if err := s.repo.Create(item); err != nil {
		return nil, err
	}
	return item, nil
}
func (s *AnnouncementService) Update(id int64, input *AnnouncementInput) (*model.Announcement, error) {
	item, err := s.repo.FindByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New(errors.ErrCodeNotFound, "通知不存在")
		}
		return nil, err
	}
	if _, err = announcementFromInput(item, input); err != nil {
		return nil, err
	}
	if err = s.repo.Save(item); err != nil {
		return nil, err
	}
	return item, nil
}
func (s *AnnouncementService) Delete(id int64) error { return s.repo.Delete(id) }
func announcementFromInput(item *model.Announcement, input *AnnouncementInput) (*model.Announcement, error) {
	title, content := strings.TrimSpace(input.Title), strings.TrimSpace(input.Content)
	if title == "" || content == "" {
		return nil, errors.New(errors.ErrCodeBadRequest, "标题和内容不能为空")
	}
	if item == nil {
		item = &model.Announcement{CreatedAt: time.Now()}
	}
	item.Title, item.Content = title, content
	item.Category = strings.TrimSpace(input.Category)
	if item.Category == "" {
		item.Category = "system"
	}
	if input.IsPublished && !item.IsPublished {
		now := time.Now()
		item.PublishedAt = &now
	}
	if !input.IsPublished {
		item.PublishedAt = nil
	}
	item.IsPublished, item.UpdatedAt = input.IsPublished, time.Now()
	return item, nil
}
