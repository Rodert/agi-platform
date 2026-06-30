package storage

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/tencentyun/cos-go-sdk-v5"
)

type COSConfig struct {
	SecretID      string
	SecretKey     string
	Bucket        string
	Region        string
	PublicBaseURL string
	UploadPrefix  string
}

type COSStore struct {
	client        *cos.Client
	publicBaseURL string
	uploadPrefix  string
}

func NewCOSStore(cfg COSConfig) (*COSStore, error) {
	bucketURL, err := url.Parse(fmt.Sprintf("https://%s.cos.%s.myqcloud.com", cfg.Bucket, cfg.Region))
	if err != nil {
		return nil, err
	}
	return &COSStore{
		client: cos.NewClient(&cos.BaseURL{BucketURL: bucketURL}, &http.Client{
			Transport: &cos.AuthorizationTransport{SecretID: cfg.SecretID, SecretKey: cfg.SecretKey},
		}),
		publicBaseURL: strings.TrimRight(cfg.PublicBaseURL, "/"),
		uploadPrefix:  strings.Trim(cfg.UploadPrefix, "/"),
	}, nil
}

func (s *COSStore) Put(ctx context.Context, key string, body io.Reader, size int64, mimeType string) (Object, error) {
	objectKey := s.objectKey(key)
	_, err := s.client.Object.Put(ctx, objectKey, body, &cos.ObjectPutOptions{
		ObjectPutHeaderOptions: &cos.ObjectPutHeaderOptions{ContentType: mimeType},
	})
	if err != nil {
		return Object{}, ErrUploadFailed
	}
	return Object{
		Key:       objectKey,
		AppURL:    "/api/assets/" + objectKey,
		PublicURL: s.publicBaseURL + "/" + objectKey,
		Provider:  "cos",
		MimeType:  mimeType,
		Size:      size,
	}, nil
}

func (s *COSStore) PublicURL(key string) (string, bool) {
	clean, ok := CleanAssetKey(key)
	if !ok {
		return "", false
	}
	return s.publicBaseURL + "/" + clean, true
}

func (s *COSStore) objectKey(key string) string {
	if s.uploadPrefix == "" || strings.HasPrefix(key, s.uploadPrefix+"/") {
		return key
	}
	return filepath.ToSlash(filepath.Join(s.uploadPrefix, key))
}
