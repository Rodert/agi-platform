package objectstorage

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/javapub/agi-platform-backend/internal/model"
	"github.com/javapub/agi-platform-backend/internal/repository"
)

type StoredObject struct {
	StorageConfigID int64
	ResourceType    string
	ObjectKey       string
	PublicURL       string
	ContentType     string
	SizeBytes       int64
	ExpiresAt       *time.Time
}

type uploadInput struct {
	Key                string
	Body               io.Reader
	Size               int64
	ContentType        string
	ContentDisposition string
	CacheControl       string
}

type provider interface {
	Upload(context.Context, uploadInput) error
	Download(context.Context, string) (io.ReadCloser, error)
	PresignGet(context.Context, string, time.Duration) (string, error)
}

// TemporaryReadURL grants an upstream provider a short-lived GET URL for a
// private object. Public resource policies keep using their stable public URL.
func (m *Manager) TemporaryReadURL(ctx context.Context, stored *StoredObject, ttl time.Duration) (string, error) {
	if stored == nil || stored.StorageConfigID == 0 || stored.ObjectKey == "" {
		return "", fmt.Errorf("参考资源信息不完整")
	}
	policy, err := m.policyRepo.FindByType(stored.ResourceType)
	if err != nil {
		return "", err
	}
	if policy.IsPublic {
		if stored.PublicURL == "" {
			return "", fmt.Errorf("公开参考资源缺少访问地址")
		}
		return stored.PublicURL, nil
	}
	if ttl <= 0 {
		ttl = 15 * time.Minute
	}
	config, err := m.configRepo.FindByID(stored.StorageConfigID)
	if err != nil {
		return "", err
	}
	provider, err := newProvider(config)
	if err != nil {
		return "", err
	}
	return provider.PresignGet(ctx, stored.ObjectKey, ttl)
}

type Manager struct {
	configRepo *repository.StorageConfigRepository
	policyRepo *repository.ResourcePolicyRepository
	httpClient *http.Client
}

// Download opens an existing object through the storage configuration that
// originally stored it. This keeps task downloads independent from the
// currently active storage configuration.
func (m *Manager) Download(ctx context.Context, asset *model.MediaAsset) (io.ReadCloser, error) {
	if asset == nil || asset.StorageConfigID == 0 || asset.ObjectKey == "" {
		return nil, fmt.Errorf("下载资源信息不完整")
	}
	config, err := m.configRepo.FindByID(asset.StorageConfigID)
	if err != nil {
		return nil, err
	}
	provider, err := newProvider(config)
	if err != nil {
		return nil, err
	}
	return provider.Download(ctx, asset.ObjectKey)
}

func NewManager(configRepo *repository.StorageConfigRepository, policyRepo *repository.ResourcePolicyRepository) *Manager {
	return &Manager{configRepo: configRepo, policyRepo: policyRepo, httpClient: &http.Client{Timeout: 10 * time.Minute}}
}

func (m *Manager) UploadBase64(ctx context.Context, resourceType, encoded string) (*StoredObject, error) {
	contentType, data, err := decodeBase64(encoded)
	if err != nil {
		return nil, err
	}
	if resourceType == "image" || resourceType == "thumbnail" || resourceType == "reference" {
		if !strings.HasPrefix(contentType, "image/") {
			return nil, fmt.Errorf("资源不是有效图片")
		}
	}
	return m.upload(ctx, resourceType, bytes.NewReader(data), int64(len(data)), contentType)
}

// UploadFromURL streams an upstream result through a bounded local temporary
// file, allowing size validation before an object-store upload begins.
func (m *Manager) UploadFromURL(ctx context.Context, resourceType, sourceURL string) (*StoredObject, error) {
	policy, err := m.policyRepo.FindByType(resourceType)
	if err != nil {
		return nil, err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, sourceURL, nil)
	if err != nil {
		return nil, err
	}
	response, err := m.httpClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("下载上游资源失败: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, fmt.Errorf("下载上游资源失败: %s", response.Status)
	}
	maxSize := int64(policy.MaxSizeMB) * 1024 * 1024
	if response.ContentLength > maxSize {
		return nil, fmt.Errorf("资源超过 %dMB 限制", policy.MaxSizeMB)
	}
	file, err := os.CreateTemp("", "agi-storage-*")
	if err != nil {
		return nil, err
	}
	defer os.Remove(file.Name())
	defer file.Close()
	size, err := io.Copy(file, io.LimitReader(response.Body, maxSize+1))
	if err != nil {
		return nil, err
	}
	if size > maxSize {
		return nil, fmt.Errorf("资源超过 %dMB 限制", policy.MaxSizeMB)
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return nil, err
	}
	contentType := strings.Split(response.Header.Get("Content-Type"), ";")[0]
	if contentType == "" || contentType == "application/octet-stream" {
		contentType = fallbackContentType(resourceType, sourceURL)
	}
	if resourceType == "video" && !strings.HasPrefix(contentType, "video/") {
		return nil, fmt.Errorf("上游结果不是视频文件")
	}
	if (resourceType == "image" || resourceType == "thumbnail") && !strings.HasPrefix(contentType, "image/") {
		return nil, fmt.Errorf("上游结果不是图片文件")
	}
	return m.upload(ctx, resourceType, file, size, contentType)
}

// Promote copies a generated asset into a resource class with its own lifecycle.
// The source can be in an old storage configuration; the copy always lands in
// the currently enabled configuration so published works have an independent URL.
func (m *Manager) Promote(ctx context.Context, source *model.MediaAsset, resourceType string) (*StoredObject, error) {
	if source == nil || source.ObjectKey == "" || source.ContentType == "" || source.SizeBytes <= 0 {
		return nil, fmt.Errorf("原始资源记录不完整")
	}
	if source.ExpiresAt != nil && !source.ExpiresAt.After(time.Now()) {
		return nil, fmt.Errorf("原始资源已过期")
	}
	policy, err := m.policyRepo.FindByType(resourceType)
	if err != nil {
		return nil, err
	}
	if !policy.IsPublic {
		return nil, fmt.Errorf("发布资源策略必须开启公开访问")
	}
	maxSize := int64(policy.MaxSizeMB) * 1024 * 1024
	if source.SizeBytes > maxSize {
		return nil, fmt.Errorf("资源超过发布策略的 %dMB 限制", policy.MaxSizeMB)
	}
	sourceConfig, err := m.configRepo.FindByID(source.StorageConfigID)
	if err != nil {
		return nil, fmt.Errorf("原始资源的存储配置不存在: %w", err)
	}
	sourceProvider, err := newProvider(sourceConfig)
	if err != nil {
		return nil, err
	}
	body, err := sourceProvider.Download(ctx, source.ObjectKey)
	if err != nil {
		return nil, err
	}
	defer body.Close()

	config, err := m.configRepo.GetEnabledConfig()
	if err != nil {
		return nil, fmt.Errorf("未找到启用的对象存储配置: %w", err)
	}
	objectKey, err := buildKey(policy.KeyPrefix, source.ContentType, resourceType)
	if err != nil {
		return nil, err
	}
	publicURL, err := buildPublicURL(config, policy, objectKey)
	if err != nil {
		return nil, err
	}
	targetProvider, err := newProvider(config)
	if err != nil {
		return nil, err
	}
	cacheControl := ""
	if policy.CacheMaxAge > 0 {
		cacheControl = fmt.Sprintf("public, max-age=%d", policy.CacheMaxAge)
	}
	if err := targetProvider.Upload(ctx, uploadInput{Key: objectKey, Body: body, Size: source.SizeBytes, ContentType: source.ContentType, ContentDisposition: "inline", CacheControl: cacheControl}); err != nil {
		return nil, err
	}
	return &StoredObject{StorageConfigID: config.ID, ResourceType: resourceType, ObjectKey: objectKey, PublicURL: publicURL, ContentType: source.ContentType, SizeBytes: source.SizeBytes}, nil
}

func (m *Manager) upload(ctx context.Context, resourceType string, body io.Reader, size int64, contentType string) (*StoredObject, error) {
	policy, err := m.policyRepo.FindByType(resourceType)
	if err != nil {
		return nil, err
	}
	maxSize := int64(policy.MaxSizeMB) * 1024 * 1024
	if size <= 0 || size > maxSize {
		return nil, fmt.Errorf("资源大小必须在 1B 到 %dMB 之间", policy.MaxSizeMB)
	}
	config, err := m.configRepo.GetEnabledConfig()
	if err != nil {
		return nil, fmt.Errorf("未找到启用的对象存储配置: %w", err)
	}
	objectKey, err := buildKey(policy.KeyPrefix, contentType, resourceType)
	if err != nil {
		return nil, err
	}
	publicURL, err := buildPublicURL(config, policy, objectKey)
	if err != nil {
		return nil, err
	}
	provider, err := newProvider(config)
	if err != nil {
		return nil, err
	}
	cacheControl := ""
	if policy.IsPublic && policy.CacheMaxAge > 0 {
		cacheControl = fmt.Sprintf("public, max-age=%d", policy.CacheMaxAge)
	}
	if err := provider.Upload(ctx, uploadInput{Key: objectKey, Body: body, Size: size, ContentType: contentType, ContentDisposition: "inline", CacheControl: cacheControl}); err != nil {
		return nil, err
	}
	var expiresAt *time.Time
	if policy.RetentionDays > 0 {
		value := time.Now().AddDate(0, 0, policy.RetentionDays)
		expiresAt = &value
	}
	return &StoredObject{StorageConfigID: config.ID, ResourceType: resourceType, ObjectKey: objectKey, PublicURL: publicURL, ContentType: contentType, SizeBytes: size, ExpiresAt: expiresAt}, nil
}

func buildKey(prefix, contentType, resourceType string) (string, error) {
	prefix = strings.Trim(prefix, "/")
	if prefix == "" || strings.Contains(prefix, "..") {
		return "", fmt.Errorf("资源路径前缀无效")
	}
	extensions, _ := mime.ExtensionsByType(contentType)
	extension := ""
	if len(extensions) > 0 {
		extension = extensions[0]
	}
	if extension == "" {
		extension = map[string]string{"video": ".mp4", "image": ".png", "thumbnail": ".jpg", "reference": ".png"}[resourceType]
	}
	return path.Join(prefix, time.Now().Format("2006/01/02"), uuid.NewString()+extension), nil
}

func buildPublicURL(config *model.StorageConfig, policy *model.ResourcePolicy, objectKey string) (string, error) {
	if !policy.IsPublic {
		return "", nil
	}
	base := strings.TrimRight(config.Domain, "/")
	if base == "" && config.Type == "local" {
		base = "http://localhost:8080/uploads"
	}
	if base == "" {
		return "", fmt.Errorf("公开资源策略要求配置访问域名")
	}
	return base + "/" + objectKey, nil
}

func decodeBase64(encoded string) (string, []byte, error) {
	contentType := "image/png"
	data := encoded
	if strings.HasPrefix(encoded, "data:") {
		parts := strings.SplitN(encoded, ",", 2)
		if len(parts) != 2 {
			return "", nil, fmt.Errorf("Base64 格式错误")
		}
		contentType = strings.TrimPrefix(strings.Split(parts[0], ";")[0], "data:")
		data = parts[1]
	}
	decoded, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return "", nil, fmt.Errorf("Base64 解码失败: %w", err)
	}
	actualType := http.DetectContentType(decoded)
	if strings.HasPrefix(actualType, "image/") {
		contentType = actualType
	}
	return contentType, decoded, nil
}

func fallbackContentType(resourceType, sourceURL string) string {
	switch resourceType {
	case "video":
		return "video/mp4"
	case "thumbnail", "image", "reference":
		if strings.EqualFold(filepath.Ext(sourceURL), ".webp") {
			return "image/webp"
		}
		if strings.EqualFold(filepath.Ext(sourceURL), ".jpg") || strings.EqualFold(filepath.Ext(sourceURL), ".jpeg") {
			return "image/jpeg"
		}
		return "image/png"
	}
	return "application/octet-stream"
}
