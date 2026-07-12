package service

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/javapub/agi-platform-backend/pkg/errors"
)

// StorageService 对象存储服务（临时实现，后续对接 MinIO）
type StorageService struct {
	// TODO: 注入 MinIO Client
}

func NewStorageService() *StorageService {
	return &StorageService{}
}

// UploadBase64Image 上传 Base64 图片
func (s *StorageService) UploadBase64Image(base64Str string) (string, error) {
	// 1. 解析 Base64
	imageData, contentType, err := s.decodeBase64(base64Str)
	if err != nil {
		return "", err
	}

	// 2. 验证文件大小（限制5MB）
	if len(imageData) > 5*1024*1024 {
		return "", errors.ErrFileTooLarge
	}

	// 3. 验证文件类型
	if !strings.HasPrefix(contentType, "image/") {
		return "", errors.ErrInvalidFileType
	}

	// 4. 生成文件名
	ext := s.getExtFromContentType(contentType)
	filename := fmt.Sprintf("references/%s/%s%s",
		time.Now().Format("2006/01/02"),
		uuid.New().String(),
		ext,
	)

	// TODO: 上传到 MinIO
	// _, err = s.minioClient.PutObject(...)

	// 临时返回假URL（实际应该上传到对象存储）
	url := fmt.Sprintf("https://cdn.tide.ai/%s", filename)
	return url, nil
}

// UploadFile 上传文件
func (s *StorageService) UploadFile(file io.Reader, contentType string, size int64) (string, error) {
	// 验证大小
	if size > 5*1024*1024 {
		return "", errors.ErrFileTooLarge
	}

	// 生成文件名
	ext := s.getExtFromContentType(contentType)
	filename := fmt.Sprintf("uploads/%s/%s%s",
		time.Now().Format("2006/01/02"),
		uuid.New().String(),
		ext,
	)

	// TODO: 上传到 MinIO
	// _, err := s.minioClient.PutObject(...)

	url := fmt.Sprintf("https://cdn.tide.ai/%s", filename)
	return url, nil
}

// decodeBase64 解码 Base64
func (s *StorageService) decodeBase64(base64Str string) ([]byte, string, error) {
	var contentType string
	var base64Data string

	// 处理 data:image/png;base64,iVBORw0KGgo... 格式
	if strings.HasPrefix(base64Str, "data:") {
		parts := strings.Split(base64Str, ",")
		if len(parts) != 2 {
			return nil, "", fmt.Errorf("Base64格式错误")
		}

		// 提取 content-type
		header := parts[0]
		contentType = strings.TrimPrefix(
			strings.Split(header, ";")[0],
			"data:",
		)
		base64Data = parts[1]
	} else {
		// 纯Base64字符串，默认为JPEG
		contentType = "image/jpeg"
		base64Data = base64Str
	}

	// 解码
	imageData, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		return nil, "", fmt.Errorf("Base64解码失败: %w", err)
	}

	// 验证是否真的是图片
	actualContentType := http.DetectContentType(imageData)
	if !strings.HasPrefix(actualContentType, "image/") {
		return nil, "", fmt.Errorf("文件不是有效的图片")
	}

	return imageData, contentType, nil
}

// getExtFromContentType 根据 ContentType 获取扩展名
func (s *StorageService) getExtFromContentType(contentType string) string {
	switch contentType {
	case "image/jpeg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/webp":
		return ".webp"
	case "image/gif":
		return ".gif"
	default:
		return ".jpg"
	}
}

// ReadImageData 读取图片数据
func (s *StorageService) ReadImageData(file io.Reader) ([]byte, error) {
	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(file)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
