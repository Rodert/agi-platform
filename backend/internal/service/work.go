package service

import (
	"time"

	"github.com/javapub/agi-platform-backend/internal/dto"
	"github.com/javapub/agi-platform-backend/internal/model"
	"github.com/javapub/agi-platform-backend/internal/repository"
	"github.com/javapub/agi-platform-backend/pkg/errors"
	"gorm.io/gorm"
)

type WorkService struct {
	workRepo *repository.WorkRepository
	taskRepo *repository.TaskRepository
	userRepo *repository.UserRepository
}

func NewWorkService(
	workRepo *repository.WorkRepository,
	taskRepo *repository.TaskRepository,
	userRepo *repository.UserRepository,
) *WorkService {
	return &WorkService{
		workRepo: workRepo,
		taskRepo: taskRepo,
		userRepo: userRepo,
	}
}

// PublishWork 发布作品
func (s *WorkService) PublishWork(userID int64, req *dto.PublishWorkRequest) (*dto.WorkResponse, error) {
	// 1. 验证任务是否存在且属于该用户
	task, err := s.taskRepo.FindByID(req.TaskID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrTaskNotFound
		}
		return nil, err
	}

	if task.UserID != userID {
		return nil, errors.ErrForbidden
	}

	// 2. 验证任务是否成功
	if task.Status != "success" {
		return nil, errors.New(errors.ErrCodeBadRequest, "只能发布成功的任务")
	}

	// 3. 检查是否已发布
	existingWork, _ := s.workRepo.FindByTaskID(req.TaskID)
	if existingWork != nil {
		return nil, errors.New(errors.ErrCodeBadRequest, "该任务已发布为作品")
	}

	// 4. 创建作品
	now := time.Now()
	work := &model.Work{
		UserID:      userID,
		TaskID:      req.TaskID,
		Title:       req.Title,
		Prompt:      task.Prompt,
		Category:    req.Category,
		Type:        task.Type,
		ImageURL:    task.ResultURL,
		VideoURL:    task.ResultURL,
		AuditStatus: "pending",
		PublishedAt: &now,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.workRepo.Create(work); err != nil {
		return nil, err
	}

	return s.workToResponse(work, userID), nil
}

// GetWork 获取作品详情
func (s *WorkService) GetWork(workID, userID int64) (*dto.WorkResponse, error) {
	work, err := s.workRepo.FindByID(workID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrWorkNotFound
		}
		return nil, err
	}

	// 增加浏览量
	s.workRepo.IncrementViews(workID)

	// 获取用户信息
	resp := s.workToResponse(work, userID)
	if user, err := s.userRepo.FindByID(work.UserID); err == nil {
		resp.User = &dto.UserInfo{
			ID:     user.ID,
			Name:   user.Name,
			Avatar: user.Avatar,
		}
	}

	return resp, nil
}

// GetWorkList 获取作品列表
func (s *WorkService) GetWorkList(req *dto.WorkListRequest, userID int64) ([]*dto.WorkResponse, int64, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}

	works, total, err := s.workRepo.FindList(req.Category, req.Type, req.UserID, req.Page, req.PageSize)
	if err != nil {
		return nil, 0, err
	}

	responses := make([]*dto.WorkResponse, len(works))
	for i, work := range works {
		responses[i] = s.workToResponse(work, userID)

		// 获取用户信息
		if user, err := s.userRepo.FindByID(work.UserID); err == nil {
			responses[i].User = &dto.UserInfo{
				ID:     user.ID,
				Name:   user.Name,
				Avatar: user.Avatar,
			}
		}
	}

	return responses, total, nil
}

// LikeWork 点赞作品
func (s *WorkService) LikeWork(userID, workID int64) error {
	// 检查作品是否存在
	work, err := s.workRepo.FindByID(workID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.ErrWorkNotFound
		}
		return err
	}

	// 检查是否已点赞
	isLiked, err := s.workRepo.IsLiked(userID, workID)
	if err != nil {
		return err
	}
	if isLiked {
		return errors.ErrWorkAlreadyLiked
	}

	// 不能给自己的作品点赞
	if work.UserID == userID {
		return errors.New(errors.ErrCodeBadRequest, "不能给自己的作品点赞")
	}

	return s.workRepo.Like(userID, workID)
}

// UnlikeWork 取消点赞
func (s *WorkService) UnlikeWork(userID, workID int64) error {
	// 检查是否已点赞
	isLiked, err := s.workRepo.IsLiked(userID, workID)
	if err != nil {
		return err
	}
	if !isLiked {
		return errors.ErrWorkNotLiked
	}

	return s.workRepo.Unlike(userID, workID)
}

// CollectWork 收藏作品
func (s *WorkService) CollectWork(userID, workID int64) error {
	// 检查作品是否存在
	_, err := s.workRepo.FindByID(workID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.ErrWorkNotFound
		}
		return err
	}

	// 检查是否已收藏
	isCollected, err := s.workRepo.IsCollected(userID, workID)
	if err != nil {
		return err
	}
	if isCollected {
		return errors.New(errors.ErrCodeBadRequest, "已收藏")
	}

	return s.workRepo.Collect(userID, workID)
}

// UncollectWork 取消收藏
func (s *WorkService) UncollectWork(userID, workID int64) error {
	// 检查是否已收藏
	isCollected, err := s.workRepo.IsCollected(userID, workID)
	if err != nil {
		return err
	}
	if !isCollected {
		return errors.New(errors.ErrCodeBadRequest, "未收藏")
	}

	return s.workRepo.Uncollect(userID, workID)
}

// workToResponse 转换为响应
func (s *WorkService) workToResponse(work *model.Work, userID int64) *dto.WorkResponse {
	resp := &dto.WorkResponse{
		ID:            work.ID,
		UserID:        work.UserID,
		Title:         work.Title,
		Prompt:        work.Prompt,
		Category:      work.Category,
		Type:          work.Type,
		Ratio:         work.Ratio,
		ImageURL:      work.ImageURL,
		VideoURL:      work.VideoURL,
		AuditStatus:   work.AuditStatus,
		LikesCount:    work.LikesCount,
		CollectsCount: work.CollectsCount,
		ViewsCount:    work.ViewsCount,
		CreatedAt:     work.CreatedAt.Format("2006-01-02 15:04:05"),
	}

	if work.PublishedAt != nil {
		resp.PublishedAt = work.PublishedAt.Format("2006-01-02 15:04:05")
	}

	// 检查是否点赞和收藏（如果已登录）
	if userID > 0 {
		isLiked, _ := s.workRepo.IsLiked(userID, work.ID)
		resp.IsLiked = isLiked

		isCollected, _ := s.workRepo.IsCollected(userID, work.ID)
		resp.IsCollected = isCollected
	}

	return resp
}
