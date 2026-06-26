package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"agi-platform/server/internal/model"
	"agi-platform/server/internal/provider"
	"agi-platform/server/internal/repository"

	"gorm.io/datatypes"
)

var (
	ErrInsufficientCredits = errors.New("insufficient credits")
	ErrInvalidRequest      = errors.New("invalid request")
)

type ImageTaskService interface {
	Generate(ctx context.Context, req GenerateImageRequest) (*GenerateImageResult, error)
	Get(ctx context.Context, taskNo string) (*ImageTaskDetail, error)
	GetForUser(ctx context.Context, userID uint64, taskNo string) (*ImageTaskDetail, error)
	List(ctx context.Context, limit int, offset int) ([]model.ImageTask, error)
}

type GenerateImageRequest struct {
	UserID         uint64
	APIKeyID       *uint64
	Source         string
	Model          string
	Prompt         string
	NegativePrompt string
	Size           string
	N              int
}

type GenerateImageResult struct {
	Task   model.ImageTask    `json:"task"`
	Images []model.ImageAsset `json:"images"`
}

type ImageTaskDetail struct {
	Task   model.ImageTask    `json:"task"`
	Images []model.ImageAsset `json:"images"`
}

type imageTaskService struct {
	repos       repository.Repositories
	providerHub *provider.Registry
}

func NewImageTaskService(repos repository.Repositories, providerHub *provider.Registry) ImageTaskService {
	return &imageTaskService{
		repos:       repos,
		providerHub: providerHub,
	}
}

func (s *imageTaskService) Generate(ctx context.Context, req GenerateImageRequest) (*GenerateImageResult, error) {
	if req.UserID == 0 || req.Model == "" || req.Prompt == "" || req.Size == "" {
		return nil, ErrInvalidRequest
	}
	if req.N <= 0 {
		req.N = 1
	}
	if req.Source == "" {
		req.Source = "web"
	}

	imageModel, err := s.repos.ImageModels.FindByCode(ctx, req.Model)
	if err != nil {
		return nil, err
	}
	if req.N > imageModel.MaxImagesPerRequest {
		return nil, fmt.Errorf("%w: num_images exceeds model limit", ErrInvalidRequest)
	}

	route, err := s.repos.ImageModels.PickRoute(ctx, imageModel.ID)
	if err != nil {
		return nil, err
	}

	upstream, err := s.repos.Providers.FindByID(ctx, route.ProviderID)
	if err != nil {
		return nil, err
	}

	adapter, err := s.providerHub.Get(upstream.Type)
	if err != nil {
		return nil, err
	}

	credits := imageModel.PriceCredits * int64(req.N)
	task := &model.ImageTask{
		TaskNo:       newTaskNo(),
		UserID:       req.UserID,
		APIKeyID:     req.APIKeyID,
		ModelID:      imageModel.ID,
		RouteID:      &route.ID,
		ProviderID:   &upstream.ID,
		Source:       req.Source,
		Prompt:       req.Prompt,
		Size:         req.Size,
		NumImages:    req.N,
		Status:       model.ImageTaskStatusPending,
		Progress:     0,
		CreditsUsed:  credits,
		RefundStatus: "none",
	}
	if req.NegativePrompt != "" {
		task.NegativePrompt = &req.NegativePrompt
	}

	if err := s.createTaskAndDeductCredits(ctx, task, credits); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrInsufficientCredits
		}
		return nil, err
	}

	now := time.Now()
	_ = s.repos.ImageTasks.UpdateStatus(ctx, nil, task.ID, model.ImageTaskStatusRunning, 10, map[string]interface{}{
		"started_at": now,
	})

	result, err := adapter.Generate(ctx, provider.ImageRequest{
		Model:          route.ProviderModelName,
		Prompt:         req.Prompt,
		NegativePrompt: req.NegativePrompt,
		Size:           req.Size,
		N:              req.N,
		UserID:         req.UserID,
	})
	if err != nil {
		_ = s.markFailedAndRefund(ctx, task, imageModel.AutoRefundOnFailure, err)
		return nil, err
	}

	assets, err := s.markSucceeded(ctx, task, imageModel.ID, result)
	if err != nil {
		return nil, err
	}

	task.Status = model.ImageTaskStatusSucceeded
	task.Progress = 100
	task.ProviderRequestID = result.ProviderTaskID
	task.ProviderResponse = datatypes.JSON([]byte(result.RawResponse))
	task.CompletedAt = &now

	return &GenerateImageResult{
		Task:   *task,
		Images: assets,
	}, nil
}

func (s *imageTaskService) Get(ctx context.Context, taskNo string) (*ImageTaskDetail, error) {
	task, err := s.repos.ImageTasks.FindByTaskNo(ctx, taskNo)
	if err != nil {
		return nil, err
	}

	assets, err := s.repos.ImageTasks.ListAssetsByTaskID(ctx, task.ID)
	if err != nil {
		return nil, err
	}

	return &ImageTaskDetail{
		Task:   *task,
		Images: assets,
	}, nil
}

func (s *imageTaskService) GetForUser(ctx context.Context, userID uint64, taskNo string) (*ImageTaskDetail, error) {
	detail, err := s.Get(ctx, taskNo)
	if err != nil {
		return nil, err
	}
	if detail.Task.UserID != userID {
		return nil, repository.ErrNotFound
	}
	return detail, nil
}

func (s *imageTaskService) List(ctx context.Context, limit int, offset int) ([]model.ImageTask, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	return s.repos.ImageTasks.List(ctx, limit, offset)
}

func (s *imageTaskService) createTaskAndDeductCredits(ctx context.Context, task *model.ImageTask, credits int64) error {
	return s.repos.Tx.Transaction(ctx, func(tx repository.Tx) error {
		before, err := s.repos.Users.FindByID(ctx, task.UserID)
		if err != nil {
			return err
		}
		if before.Credits < credits {
			return repository.ErrNotFound
		}

		after, err := s.repos.Users.DeductCredits(ctx, tx, task.UserID, credits)
		if err != nil {
			return err
		}
		if err := s.repos.ImageTasks.Create(ctx, tx, task); err != nil {
			return err
		}

		relatedID := task.ID
		return s.repos.Wallets.CreateLog(ctx, tx, &model.WalletLog{
			UserID:        task.UserID,
			Type:          model.WalletLogTypeConsume,
			Amount:        -credits,
			BalanceBefore: before.Credits,
			BalanceAfter:  after.Credits,
			RelatedType:   "image_task",
			RelatedID:     &relatedID,
			Remark:        "image generation",
			OperatorType:  "system",
		})
	})
}

func (s *imageTaskService) markSucceeded(ctx context.Context, task *model.ImageTask, modelID uint64, result *provider.ImageResult) ([]model.ImageAsset, error) {
	raw := datatypes.JSON([]byte(result.RawResponse))
	if !json.Valid(raw) {
		raw = datatypes.JSON([]byte(`{}`))
	}

	var assets []model.ImageAsset
	err := s.repos.Tx.Transaction(ctx, func(tx repository.Tx) error {
		now := time.Now()
		if err := s.repos.ImageTasks.UpdateStatus(ctx, tx, task.ID, model.ImageTaskStatusSucceeded, 100, map[string]interface{}{
			"provider_request_id": result.ProviderTaskID,
			"provider_response":   raw,
			"completed_at":        now,
		}); err != nil {
			return err
		}

		for _, image := range result.Images {
			width := nullableInt(image.Width)
			height := nullableInt(image.Height)
			prompt := task.Prompt
			asset := model.ImageAsset{
				TaskID:          task.ID,
				UserID:          task.UserID,
				ModelID:         modelID,
				URL:             image.URL,
				Width:           width,
				Height:          height,
				MimeType:        "image/png",
				Prompt:          &prompt,
				Status:          "active",
				ViolationStatus: "normal",
			}
			if err := s.repos.ImageTasks.CreateAsset(ctx, tx, &asset); err != nil {
				return err
			}
			assets = append(assets, asset)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return assets, nil
}

func (s *imageTaskService) markFailedAndRefund(ctx context.Context, task *model.ImageTask, shouldRefund bool, cause error) error {
	return s.repos.Tx.Transaction(ctx, func(tx repository.Tx) error {
		now := time.Now()
		values := map[string]interface{}{
			"error_message": cause.Error(),
			"completed_at":  now,
		}

		if shouldRefund {
			before, err := s.repos.Users.FindByID(ctx, task.UserID)
			if err != nil {
				return err
			}
			after, err := s.repos.Users.AddCredits(ctx, tx, task.UserID, task.CreditsUsed)
			if err != nil {
				return err
			}
			relatedID := task.ID
			if err := s.repos.Wallets.CreateLog(ctx, tx, &model.WalletLog{
				UserID:        task.UserID,
				Type:          model.WalletLogTypeRefund,
				Amount:        task.CreditsUsed,
				BalanceBefore: before.Credits,
				BalanceAfter:  after.Credits,
				RelatedType:   "image_task",
				RelatedID:     &relatedID,
				Remark:        "image generation failed refund",
				OperatorType:  "system",
			}); err != nil {
				return err
			}
			values["refund_status"] = "refunded"
		}

		return s.repos.ImageTasks.UpdateStatus(ctx, tx, task.ID, model.ImageTaskStatusFailed, 100, values)
	})
}

func newTaskNo() string {
	return fmt.Sprintf("task_%d", time.Now().UnixNano())
}

func nullableInt(value int) *int {
	if value <= 0 {
		return nil
	}
	return &value
}
