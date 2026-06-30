package service

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"strings"
	"time"

	"agi-platform/server/internal/model"
	"agi-platform/server/internal/provider"
	"agi-platform/server/internal/repository"
	"agi-platform/server/internal/storage"

	"gorm.io/datatypes"
)

var (
	ErrInsufficientCredits = errors.New("insufficient credits")
	ErrInvalidRequest      = errors.New("invalid request")
)

type ImageTaskService interface {
	Generate(ctx context.Context, req GenerateImageRequest) (*GenerateImageResult, error)
	Submit(ctx context.Context, req GenerateImageRequest) (*GenerateImageResult, error)
	Get(ctx context.Context, taskNo string) (*ImageTaskDetail, error)
	GetForUser(ctx context.Context, userID uint64, taskNo string) (*ImageTaskDetail, error)
	List(ctx context.Context, limit int, offset int) ([]model.ImageTask, error)
	ListFiltered(ctx context.Context, filter repository.TaskFilter, limit int, offset int) ([]model.ImageTask, error)
	ListForUser(ctx context.Context, userID uint64, limit int, offset int) ([]model.ImageTask, error)
}

type GenerateImageRequest struct {
	UserID          uint64
	APIKeyID        *uint64
	Source          string
	Model           string
	Prompt          string
	NegativePrompt  string
	Size            string
	N               int
	ReferenceImages []provider.ReferenceImage
	AppBaseURL      string
}

type GenerateImageResult struct {
	Task   model.ImageTask    `json:"task"`
	Images []model.ImageAsset `json:"images"`
}

type ImageTaskDetail struct {
	Task   model.ImageTask    `json:"task"`
	Images []model.ImageAsset `json:"images"`
}

type preparedImageTask struct {
	task        *model.ImageTask
	imageModel  *model.ImageModel
	route       *model.ImageModelRoute
	upstream    *model.Provider
	upstreamKey *model.ProviderKey
	extraConfig map[string]interface{}
	req         GenerateImageRequest
}

type imageTaskService struct {
	repos       repository.Repositories
	providerHub *provider.Registry
	store       storage.Store
}

func NewImageTaskService(repos repository.Repositories, providerHub *provider.Registry, store storage.Store) ImageTaskService {
	return &imageTaskService{
		repos:       repos,
		providerHub: providerHub,
		store:       store,
	}
}

func (s *imageTaskService) Generate(ctx context.Context, req GenerateImageRequest) (*GenerateImageResult, error) {
	prepared, err := s.prepareAndCreateTask(ctx, req)
	if err != nil {
		return nil, err
	}

	assets, err := s.runPreparedTask(context.Background(), prepared)
	if err != nil {
		return nil, err
	}

	task, err := s.repos.ImageTasks.FindByTaskNo(ctx, prepared.task.TaskNo)
	if err != nil {
		return nil, err
	}
	return &GenerateImageResult{
		Task:   *task,
		Images: assets,
	}, nil
}

func (s *imageTaskService) Submit(ctx context.Context, req GenerateImageRequest) (*GenerateImageResult, error) {
	prepared, err := s.prepareAndCreateTask(ctx, req)
	if err != nil {
		return nil, err
	}

	go func() {
		_, _ = s.runPreparedTask(context.Background(), prepared)
	}()

	return &GenerateImageResult{
		Task:   *prepared.task,
		Images: []model.ImageAsset{},
	}, nil
}

func (s *imageTaskService) prepareAndCreateTask(ctx context.Context, req GenerateImageRequest) (*preparedImageTask, error) {
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
	if !upstream.Enabled {
		return nil, fmt.Errorf("%w: provider is disabled", ErrInvalidRequest)
	}

	if _, err := s.providerHub.Get(upstream.Type); err != nil {
		return nil, err
	}

	var upstreamKey *model.ProviderKey
	if upstream.Type != "mock" {
		if route.ProviderKeyID != nil {
			upstreamKey, err = s.repos.Providers.FindKeyByID(ctx, *route.ProviderKeyID)
			if err == nil && upstreamKey.ProviderID != upstream.ID {
				return nil, fmt.Errorf("%w: route provider key does not belong to provider", ErrInvalidRequest)
			}
		} else {
			upstreamKey, err = s.repos.Providers.PickActiveKey(ctx, upstream.ID)
		}
		if err != nil {
			return nil, err
		}
	}

	extraConfig := map[string]interface{}{}
	if len(route.ExtraConfig) > 0 {
		if err := json.Unmarshal(route.ExtraConfig, &extraConfig); err != nil {
			return nil, fmt.Errorf("%w: invalid route extra_config", ErrInvalidRequest)
		}
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
	if upstreamKey != nil {
		task.ProviderKeyID = &upstreamKey.ID
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

	return &preparedImageTask{
		task:        task,
		imageModel:  imageModel,
		route:       route,
		upstream:    upstream,
		upstreamKey: upstreamKey,
		extraConfig: extraConfig,
		req:         req,
	}, nil
}

func (s *imageTaskService) runPreparedTask(ctx context.Context, prepared *preparedImageTask) ([]model.ImageAsset, error) {
	task := prepared.task
	now := time.Now()
	if taskCreatedAtExpired(task.CreatedAt, now) {
		_ = s.markStaleTasksTimeout(ctx)
		return nil, fmt.Errorf("%w: %s", ErrInvalidRequest, taskTimeoutMessage)
	}
	_ = s.repos.ImageTasks.UpdateStatus(ctx, nil, task.ID, model.ImageTaskStatusRunning, 10, map[string]interface{}{
		"started_at": now,
	})

	adapter, err := s.providerHub.Get(prepared.upstream.Type)
	if err != nil {
		_ = s.markFailedAndRefund(ctx, task, prepared.imageModel.AutoRefundOnFailure, err, provider.ResponseErrorRaw(err))
		return nil, err
	}

	referenceImages, err := s.normalizeImageReferences(prepared.req.ReferenceImages, prepared.req.AppBaseURL)
	if err != nil {
		_ = s.markFailedAndRefund(ctx, task, prepared.imageModel.AutoRefundOnFailure, err)
		return nil, err
	}

	result, err := adapter.Generate(ctx, provider.ImageRequest{
		Model:           prepared.route.ProviderModelName,
		Prompt:          prepared.req.Prompt,
		NegativePrompt:  prepared.req.NegativePrompt,
		Size:            prepared.req.Size,
		N:               prepared.req.N,
		UserID:          prepared.req.UserID,
		BaseURL:         prepared.upstream.BaseURL,
		APIKey:          providerAPIKey(prepared.upstreamKey),
		TimeoutSeconds:  prepared.upstream.TimeoutSeconds,
		ReferenceImages: referenceImages,
		Extra:           prepared.extraConfig,
	})
	if err != nil {
		_ = s.markFailedAndRefund(ctx, task, prepared.imageModel.AutoRefundOnFailure, err)
		return nil, err
	}

	assets, err := s.markSucceeded(ctx, task, prepared.imageModel.ID, result)
	if err != nil {
		return nil, err
	}

	return assets, nil
}

func (s *imageTaskService) Get(ctx context.Context, taskNo string) (*ImageTaskDetail, error) {
	if err := s.markStaleTasksTimeout(ctx); err != nil {
		return nil, err
	}
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
	return s.ListFiltered(ctx, repository.TaskFilter{}, limit, offset)
}

func (s *imageTaskService) ListFiltered(ctx context.Context, filter repository.TaskFilter, limit int, offset int) ([]model.ImageTask, error) {
	if err := s.markStaleTasksTimeout(ctx); err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	return s.repos.ImageTasks.ListFiltered(ctx, filter, limit, offset)
}

func (s *imageTaskService) ListForUser(ctx context.Context, userID uint64, limit int, offset int) ([]model.ImageTask, error) {
	if err := s.markStaleTasksTimeout(ctx); err != nil {
		return nil, err
	}
	if userID == 0 {
		return nil, ErrInvalidRequest
	}
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	return s.repos.ImageTasks.ListByUserID(ctx, userID, limit, offset)
}

func (s *imageTaskService) markStaleTasksTimeout(ctx context.Context) error {
	return s.repos.ImageTasks.MarkStaleRunningTimeout(ctx, taskTimeoutCutoff(time.Now()), taskTimeoutMessage)
}

func (s *imageTaskService) normalizeImageReferences(images []provider.ReferenceImage, appBaseURL string) ([]provider.ReferenceImage, error) {
	normalized := make([]provider.ReferenceImage, 0, len(images))
	for _, image := range images {
		if strings.TrimSpace(image.URL) != "" {
			value, err := normalizeReferenceAssetURL(s.store, appBaseURL, image.URL)
			if err != nil {
				return nil, err
			}
			image.URL = value
		}
		normalized = append(normalized, image)
	}
	return normalized, nil
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

	assets := make([]model.ImageAsset, 0, len(result.Images))
	for _, image := range result.Images {
		stored, err := s.storeImageURL(ctx, task.TaskNo, len(assets)+1, image.URL)
		if err != nil {
			return nil, err
		}
		width := nullableInt(image.Width)
		height := nullableInt(image.Height)
		prompt := task.Prompt
		assets = append(assets, model.ImageAsset{
			TaskID:          task.ID,
			UserID:          task.UserID,
			ModelID:         modelID,
			URL:             stored.URL,
			StorageProvider: stored.Provider,
			StorageKey:      stored.Key,
			Width:           width,
			Height:          height,
			MimeType:        stored.MimeType,
			SizeBytes:       stored.SizeBytes,
			Prompt:          &prompt,
			Status:          "active",
			ViolationStatus: "normal",
		})
	}

	err := s.repos.Tx.Transaction(ctx, func(tx repository.Tx) error {
		now := time.Now()
		if err := s.repos.ImageTasks.UpdateStatus(ctx, tx, task.ID, model.ImageTaskStatusSucceeded, 100, map[string]interface{}{
			"provider_request_id": result.ProviderTaskID,
			"provider_response":   raw,
			"completed_at":        now,
		}); err != nil {
			return err
		}

		for i := range assets {
			if err := s.repos.ImageTasks.CreateAsset(ctx, tx, &assets[i]); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return assets, nil
}

type storedImage struct {
	URL       string
	Provider  string
	Key       string
	MimeType  string
	SizeBytes *int64
}

func (s *imageTaskService) storeImageURL(ctx context.Context, taskNo string, index int, imageURL string) (storedImage, error) {
	mimeType, data, err := decodeDataURL(imageURL)
	if err != nil {
		if !strings.HasPrefix(strings.TrimSpace(imageURL), "data:") {
			mimeType, data, err = downloadImage(ctx, imageURL)
		}
		if err != nil {
			return storedImage{}, err
		}
	}
	extension := extensionForMimeType(mimeType)
	key := storage.DatedKey("images", taskNo, fmt.Sprintf("%02d%s", index, extension))
	object, err := s.store.Put(ctx, key, bytes.NewReader(data), int64(len(data)), mimeType)
	if err != nil {
		return storedImage{}, err
	}

	size := object.Size
	return storedImage{
		URL:       object.AppURL,
		Provider:  object.Provider,
		Key:       object.Key,
		MimeType:  object.MimeType,
		SizeBytes: &size,
	}, nil
}

func downloadImage(ctx context.Context, imageURL string) (string, []byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, strings.TrimSpace(imageURL), nil)
	if err != nil {
		return "", nil, err
	}
	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", nil, fmt.Errorf("%w: download image failed: status=%d", ErrInvalidRequest, resp.StatusCode)
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, 64*1024*1024))
	if err != nil {
		return "", nil, err
	}
	mimeType := resp.Header.Get("Content-Type")
	if parsed, _, err := mime.ParseMediaType(mimeType); err == nil && parsed != "" {
		mimeType = parsed
	}
	if mimeType == "" || !strings.HasPrefix(mimeType, "image/") {
		mimeType = http.DetectContentType(data)
	}
	if !strings.HasPrefix(mimeType, "image/") {
		mimeType = "image/png"
	}
	return mimeType, data, nil
}

func decodeDataURL(value string) (string, []byte, error) {
	parts := strings.SplitN(value, ",", 2)
	if len(parts) != 2 {
		return "", nil, fmt.Errorf("%w: invalid image data url", ErrInvalidRequest)
	}
	meta := strings.TrimPrefix(parts[0], "data:")
	if !strings.Contains(meta, ";base64") {
		return "", nil, fmt.Errorf("%w: image data url must be base64", ErrInvalidRequest)
	}
	mimeType := strings.Split(meta, ";")[0]
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}
	data, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return "", nil, err
	}
	return mimeType, data, nil
}

func extensionForMimeType(mimeType string) string {
	switch strings.ToLower(mimeType) {
	case "image/jpeg", "image/jpg":
		return ".jpg"
	case "image/webp":
		return ".webp"
	case "image/gif":
		return ".gif"
	default:
		return ".png"
	}
}

func (s *imageTaskService) markFailedAndRefund(ctx context.Context, task *model.ImageTask, shouldRefund bool, cause error, rawResponse ...string) error {
	return s.repos.Tx.Transaction(ctx, func(tx repository.Tx) error {
		now := time.Now()
		values := map[string]interface{}{
			"error_message": cause.Error(),
			"completed_at":  now,
		}
		if raw := firstNonEmptyString(rawResponse...); raw != "" {
			values["provider_response"] = datatypes.JSON([]byte(safeJSON(raw)))
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

func providerAPIKey(key *model.ProviderKey) string {
	if key == nil {
		return ""
	}
	return key.APIKeyEncrypted
}
