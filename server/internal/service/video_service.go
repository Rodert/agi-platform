package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"agi-platform/server/internal/model"
	"agi-platform/server/internal/provider"
	"agi-platform/server/internal/repository"
	"agi-platform/server/internal/storage"

	"gorm.io/datatypes"
)

type VideoService interface {
	ListModels(ctx context.Context, limit int, offset int) ([]model.VideoModel, error)
	Submit(ctx context.Context, req SubmitVideoRequest) (*VideoTaskDetail, error)
	Get(ctx context.Context, taskNo string) (*VideoTaskDetail, error)
	GetForUser(ctx context.Context, userID uint64, taskNo string) (*VideoTaskDetail, error)
	ListForUser(ctx context.Context, userID uint64, limit int, offset int) ([]model.VideoTask, error)
	List(ctx context.Context, limit int, offset int) ([]model.VideoTask, error)
	ListFiltered(ctx context.Context, filter repository.TaskFilter, limit int, offset int) ([]model.VideoTask, error)
}

type SubmitVideoRequest struct {
	UserID      uint64
	APIKeyID    *uint64
	Source      string
	Model       string
	Prompt      string
	Seconds     int
	AspectRatio string
	Images      []string
	Videos      []string
	Audios      []string
	AppBaseURL  string
}

type VideoTaskDetail struct {
	Task   model.VideoTask    `json:"task"`
	Videos []model.VideoAsset `json:"videos"`
}

type videoService struct {
	repos       repository.Repositories
	providerHub *provider.Registry
	store       storage.Store
}

func NewVideoService(repos repository.Repositories, providerHub *provider.Registry, store storage.Store) VideoService {
	return &videoService{repos: repos, providerHub: providerHub, store: store}
}

func (s *videoService) ListModels(ctx context.Context, limit int, offset int) ([]model.VideoModel, error) {
	limit, offset = normalizePage(limit, offset)
	return s.repos.Videos.ListModels(ctx, limit, offset)
}

func (s *videoService) Submit(ctx context.Context, req SubmitVideoRequest) (*VideoTaskDetail, error) {
	prepared, err := s.prepareAndCreateVideoTask(ctx, req)
	if err != nil {
		return nil, err
	}
	go s.runVideoTask(context.Background(), prepared)
	return &VideoTaskDetail{Task: *prepared.task, Videos: []model.VideoAsset{}}, nil
}

type preparedVideoTask struct {
	task        *model.VideoTask
	videoModel  *model.VideoModel
	route       *model.VideoModelRoute
	upstream    *model.Provider
	upstreamKey *model.ProviderKey
	extraConfig map[string]interface{}
	req         SubmitVideoRequest
}

func (s *videoService) prepareAndCreateVideoTask(ctx context.Context, req SubmitVideoRequest) (*preparedVideoTask, error) {
	if req.UserID == 0 || req.Model == "" || req.Prompt == "" {
		return nil, ErrInvalidRequest
	}
	if len(req.Images) > 4 || len(req.Videos) > 3 || len(req.Audios) > 1 {
		return nil, fmt.Errorf("%w: reference media exceeds limit", ErrInvalidRequest)
	}
	if req.Source == "" {
		req.Source = "web"
	}
	if req.Seconds <= 0 {
		req.Seconds = 15
	}
	if req.Seconds != 5 && req.Seconds != 10 && req.Seconds != 15 {
		return nil, fmt.Errorf("%w: video seconds must be 5, 10, or 15", ErrInvalidRequest)
	}
	if req.AspectRatio == "" {
		req.AspectRatio = "9:16"
	}
	videoModel, err := s.repos.Videos.FindModelByCode(ctx, req.Model)
	if err != nil {
		return nil, err
	}
	route, err := s.repos.Videos.PickRoute(ctx, videoModel.ID)
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
	if _, err := s.providerHub.GetVideo(upstream.Type); err != nil {
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
	imagesJSON, _ := json.Marshal(nonNilStringSlice(req.Images))
	videosJSON, _ := json.Marshal(nonNilStringSlice(req.Videos))
	audiosJSON, _ := json.Marshal(nonNilStringSlice(req.Audios))
	task := &model.VideoTask{
		TaskNo:       newVideoTaskNo(),
		UserID:       req.UserID,
		APIKeyID:     req.APIKeyID,
		ModelID:      videoModel.ID,
		RouteID:      &route.ID,
		ProviderID:   &upstream.ID,
		Source:       req.Source,
		Prompt:       req.Prompt,
		Seconds:      req.Seconds,
		AspectRatio:  req.AspectRatio,
		Images:       datatypes.JSON(imagesJSON),
		Videos:       datatypes.JSON(videosJSON),
		Audios:       datatypes.JSON(audiosJSON),
		Status:       model.VideoTaskStatusPending,
		CreditsUsed:  videoModel.PriceCredits,
		RefundStatus: "none",
	}
	if upstreamKey != nil {
		task.ProviderKeyID = &upstreamKey.ID
	}
	if err := s.createVideoTaskAndDeductCredits(ctx, task, task.CreditsUsed); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrInsufficientCredits
		}
		return nil, err
	}
	return &preparedVideoTask{task: task, videoModel: videoModel, route: route, upstream: upstream, upstreamKey: upstreamKey, extraConfig: extraConfig, req: req}, nil
}

func (s *videoService) runVideoTask(ctx context.Context, prepared *preparedVideoTask) {
	task := prepared.task
	now := time.Now()
	if taskCreatedAtExpired(task.CreatedAt, now) {
		_ = s.markStaleTasksTimeout(ctx)
		return
	}
	_ = s.repos.Videos.UpdateTaskStatus(ctx, nil, task.ID, model.VideoTaskStatusRunning, 5, map[string]interface{}{"started_at": now})
	adapter, err := s.providerHub.GetVideo(prepared.upstream.Type)
	if err != nil {
		_ = s.markVideoFailedAndRefund(ctx, task, err)
		return
	}
	images, videos, audios, err := s.normalizeVideoReferences(prepared.req.Images, prepared.req.Videos, prepared.req.Audios, prepared.req.AppBaseURL)
	if err != nil {
		_ = s.markVideoFailedAndRefund(ctx, task, err)
		return
	}
	created, err := adapter.CreateVideo(ctx, provider.VideoRequest{
		Model:          prepared.route.ProviderModelName,
		Prompt:         prepared.req.Prompt,
		Seconds:        prepared.req.Seconds,
		AspectRatio:    prepared.req.AspectRatio,
		Images:         images,
		Videos:         videos,
		Audios:         audios,
		UserID:         prepared.req.UserID,
		BaseURL:        prepared.upstream.BaseURL,
		APIKey:         providerAPIKey(prepared.upstreamKey),
		TimeoutSeconds: prepared.upstream.TimeoutSeconds,
		Extra:          prepared.extraConfig,
	})
	if err != nil {
		_ = s.markVideoFailedAndRefund(ctx, task, err)
		return
	}
	task.ProviderTaskID = created.TaskID
	_ = s.repos.Videos.UpdateTaskStatus(ctx, nil, task.ID, model.VideoTaskStatusRunning, 10, map[string]interface{}{
		"provider_task_id": created.TaskID,
		"provider_response": datatypes.JSON(
			[]byte(safeJSON(created.RawResponse)),
		),
	})
	status := created.Status
	var statusResult *provider.VideoStatusResult
	for i := 0; i < 90; i++ {
		if status == "succeeded" || status == "failed" {
			break
		}
		time.Sleep(5 * time.Second)
		statusResult, err = adapter.GetVideo(ctx, provider.VideoStatusRequest{
			TaskID:         created.TaskID,
			BaseURL:        prepared.upstream.BaseURL,
			APIKey:         providerAPIKey(prepared.upstreamKey),
			TimeoutSeconds: prepared.upstream.TimeoutSeconds,
		})
		if err != nil {
			_ = s.markVideoFailedAndRefund(ctx, task, err)
			return
		}
		status = statusResult.Status
		progress := statusResult.Progress
		if progress <= 0 {
			progress = 20
		}
		_ = s.repos.Videos.UpdateTaskStatus(ctx, nil, task.ID, model.VideoTaskStatusRunning, progress, map[string]interface{}{
			"provider_response": datatypes.JSON([]byte(safeJSON(statusResult.RawResponse))),
		})
	}
	if status != "succeeded" {
		_ = s.markVideoFailedAndRefund(ctx, task, videoFailureError(status, statusResult))
		return
	}
	content, mimeType, err := adapter.DownloadVideo(ctx, provider.VideoContentRequest{
		TaskID:         created.TaskID,
		BaseURL:        prepared.upstream.BaseURL,
		APIKey:         providerAPIKey(prepared.upstreamKey),
		TimeoutSeconds: prepared.upstream.TimeoutSeconds,
	})
	if err != nil {
		_ = s.markVideoFailedAndRefund(ctx, task, err)
		return
	}
	if err := s.markVideoSucceeded(ctx, task, prepared.videoModel.ID, content, mimeType, statusResult); err != nil {
		if !errors.Is(err, repository.ErrNotFound) {
			_ = s.markVideoFailedAndRefund(ctx, task, err)
		}
	}
}

func (s *videoService) createVideoTaskAndDeductCredits(ctx context.Context, task *model.VideoTask, credits int64) error {
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
		if err := s.repos.Videos.CreateTask(ctx, tx, task); err != nil {
			return err
		}
		relatedID := task.ID
		return s.repos.Wallets.CreateLog(ctx, tx, &model.WalletLog{
			UserID:        task.UserID,
			Type:          model.WalletLogTypeConsume,
			Amount:        -credits,
			BalanceBefore: before.Credits,
			BalanceAfter:  after.Credits,
			RelatedType:   "video_task",
			RelatedID:     &relatedID,
			Remark:        "video generation",
			OperatorType:  "system",
		})
	})
}

func (s *videoService) markVideoSucceeded(ctx context.Context, task *model.VideoTask, modelID uint64, content []byte, mimeType string, statusResult *provider.VideoStatusResult) error {
	if strings.TrimSpace(mimeType) == "" {
		mimeType = "video/mp4"
	}
	key := storage.DatedKey("videos", task.TaskNo, "result"+extensionForVideoMimeType(mimeType))
	object, err := s.store.Put(ctx, key, bytes.NewReader(content), int64(len(content)), mimeType)
	if err != nil {
		return err
	}
	size := object.Size
	prompt := task.Prompt
	raw := "{}"
	if statusResult != nil {
		raw = safeJSON(statusResult.RawResponse)
	}
	now := time.Now()
	return s.repos.Tx.Transaction(ctx, func(tx repository.Tx) error {
		if err := s.repos.Videos.UpdateTaskStatus(ctx, tx, task.ID, model.VideoTaskStatusSucceeded, 100, map[string]interface{}{
			"provider_response": datatypes.JSON([]byte(raw)),
			"completed_at":      now,
		}); err != nil {
			return err
		}
		return s.repos.Videos.CreateAsset(ctx, tx, &model.VideoAsset{
			TaskID:          task.ID,
			UserID:          task.UserID,
			ModelID:         modelID,
			URL:             object.AppURL,
			StorageProvider: object.Provider,
			StorageKey:      object.Key,
			MimeType:        object.MimeType,
			SizeBytes:       &size,
			DurationSeconds: nullableInt(task.Seconds),
			Prompt:          &prompt,
			Status:          "active",
		})
	})
}

func (s *videoService) normalizeVideoReferences(images []string, videos []string, audios []string, appBaseURL string) ([]string, []string, []string, error) {
	normalizedImages := make([]string, 0, len(images))
	for _, image := range images {
		normalized, err := s.normalizeImageReference(image, appBaseURL)
		if err != nil {
			return nil, nil, nil, err
		}
		normalizedImages = append(normalizedImages, normalized)
	}
	if err := rejectLocalMediaReferences("参考视频", videos); err != nil {
		return nil, nil, nil, err
	}
	if err := rejectLocalMediaReferences("参考音频", audios); err != nil {
		return nil, nil, nil, err
	}
	normalizedVideos, err := s.normalizeAssetReferences(videos, appBaseURL)
	if err != nil {
		return nil, nil, nil, err
	}
	normalizedAudios, err := s.normalizeAssetReferences(audios, appBaseURL)
	if err != nil {
		return nil, nil, nil, err
	}
	return normalizedImages, normalizedVideos, normalizedAudios, nil
}

func (s *videoService) normalizeImageReference(value string, appBaseURL string) (string, error) {
	trimmed := strings.TrimSpace(value)
	if strings.HasPrefix(trimmed, "/api/assets/") {
		return assetURLForUpstream(s.store, appBaseURL, strings.TrimPrefix(trimmed, "/api/assets/"))
	}
	if normalized, err := normalizeReferenceAssetURL(s.store, appBaseURL, trimmed); err != nil {
		return "", err
	} else if normalized != trimmed {
		return normalized, nil
	}
	if strings.HasPrefix(trimmed, "/uploads/") {
		return "", fmt.Errorf("%w: 参考图片需要公网可访问 URL，当前本地上传路径不能被上游访问", ErrInvalidRequest)
	}
	return trimmed, nil
}

func rejectLocalMediaReferences(label string, values []string) error {
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if strings.HasPrefix(trimmed, "/api/assets/") {
			continue
		}
		if strings.HasPrefix(trimmed, "/uploads/") {
			return fmt.Errorf("%w: %s需要公网可访问 URL，当前本地上传路径不能被上游访问", ErrInvalidRequest, label)
		}
	}
	return nil
}

func (s *videoService) normalizeAssetReferences(values []string, appBaseURL string) ([]string, error) {
	normalized := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if strings.HasPrefix(trimmed, "/api/assets/") {
			publicURL, err := assetURLForUpstream(s.store, appBaseURL, strings.TrimPrefix(trimmed, "/api/assets/"))
			if err != nil {
				return nil, err
			}
			normalized = append(normalized, publicURL)
			continue
		}
		if publicURL, err := normalizeReferenceAssetURL(s.store, appBaseURL, trimmed); err != nil {
			return nil, err
		} else if publicURL != trimmed {
			normalized = append(normalized, publicURL)
			continue
		}
		normalized = append(normalized, trimmed)
	}
	return normalized, nil
}

func (s *videoService) markVideoFailedAndRefund(ctx context.Context, task *model.VideoTask, cause error) error {
	return s.repos.Tx.Transaction(ctx, func(tx repository.Tx) error {
		now := time.Now()
		values := map[string]interface{}{"error_message": cause.Error(), "completed_at": now}
		if task.RefundStatus != "refunded" {
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
				RelatedType:   "video_task",
				RelatedID:     &relatedID,
				Remark:        "video generation failed refund",
				OperatorType:  "system",
			}); err != nil {
				return err
			}
			values["refund_status"] = "refunded"
		}
		return s.repos.Videos.UpdateTaskStatus(ctx, tx, task.ID, model.VideoTaskStatusFailed, 100, values)
	})
}

func (s *videoService) Get(ctx context.Context, taskNo string) (*VideoTaskDetail, error) {
	if err := s.markStaleTasksTimeout(ctx); err != nil {
		return nil, err
	}
	task, err := s.repos.Videos.FindTaskByTaskNo(ctx, taskNo)
	if err != nil {
		return nil, err
	}
	assets, err := s.repos.Videos.ListAssetsByTaskID(ctx, task.ID)
	if err != nil {
		return nil, err
	}
	return &VideoTaskDetail{Task: *task, Videos: assets}, nil
}

func (s *videoService) GetForUser(ctx context.Context, userID uint64, taskNo string) (*VideoTaskDetail, error) {
	detail, err := s.Get(ctx, taskNo)
	if err != nil {
		return nil, err
	}
	if detail.Task.UserID != userID {
		return nil, repository.ErrNotFound
	}
	return detail, nil
}

func (s *videoService) ListForUser(ctx context.Context, userID uint64, limit int, offset int) ([]model.VideoTask, error) {
	if err := s.markStaleTasksTimeout(ctx); err != nil {
		return nil, err
	}
	if userID == 0 {
		return nil, ErrInvalidRequest
	}
	limit, offset = normalizePage(limit, offset)
	return s.repos.Videos.ListTasksByUserID(ctx, userID, limit, offset)
}

func (s *videoService) List(ctx context.Context, limit int, offset int) ([]model.VideoTask, error) {
	return s.ListFiltered(ctx, repository.TaskFilter{}, limit, offset)
}

func (s *videoService) ListFiltered(ctx context.Context, filter repository.TaskFilter, limit int, offset int) ([]model.VideoTask, error) {
	if err := s.markStaleTasksTimeout(ctx); err != nil {
		return nil, err
	}
	limit, offset = normalizePage(limit, offset)
	return s.repos.Videos.ListTasksFiltered(ctx, filter, limit, offset)
}

func (s *videoService) markStaleTasksTimeout(ctx context.Context) error {
	return s.repos.Videos.MarkStaleRunningTimeout(ctx, taskTimeoutCutoff(time.Now()), taskTimeoutMessage)
}

func newVideoTaskNo() string {
	return fmt.Sprintf("video_%d", time.Now().UnixNano())
}

func safeJSON(value string) string {
	if json.Valid([]byte(value)) {
		return value
	}
	return "{}"
}

func nonNilStringSlice(values []string) []string {
	if values == nil {
		return []string{}
	}
	return values
}

func videoFailureError(status string, result *provider.VideoStatusResult) error {
	if result == nil {
		return fmt.Errorf("video task failed: %s", status)
	}
	if result.ErrorCode != "" && result.ErrorMessage != "" {
		return fmt.Errorf("video task failed: %s: %s", result.ErrorCode, result.ErrorMessage)
	}
	if result.ErrorMessage != "" {
		return fmt.Errorf("video task failed: %s", result.ErrorMessage)
	}
	if result.ErrorCode != "" {
		return fmt.Errorf("video task failed: %s", result.ErrorCode)
	}
	return fmt.Errorf("video task failed: %s", status)
}

func firstNonEmptyString(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

func extensionForVideoMimeType(mimeType string) string {
	switch strings.ToLower(mimeType) {
	case "video/quicktime":
		return ".mov"
	case "video/webm":
		return ".webm"
	case "video/x-m4v":
		return ".m4v"
	default:
		return ".mp4"
	}
}
