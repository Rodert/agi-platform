package objectstorage

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/javapub/agi-platform-backend/internal/model"
)

type r2Provider struct {
	bucket   string
	client   *s3.Client
	uploader *manager.Uploader
}

func newR2Provider(storageConfig *model.StorageConfig) (*r2Provider, error) {
	if storageConfig.Endpoint == "" || storageConfig.AccessKey == "" || storageConfig.SecretKey == "" || storageConfig.Bucket == "" {
		return nil, fmt.Errorf("Cloudflare R2 配置不完整")
	}
	region := storageConfig.Region
	if region == "" {
		region = "auto"
	}
	endpoint := strings.TrimRight(storageConfig.Endpoint, "/")
	config, err := awsconfig.LoadDefaultConfig(context.Background(), awsconfig.WithRegion(region), awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(storageConfig.AccessKey, storageConfig.SecretKey, "")), awsconfig.WithBaseEndpoint(endpoint))
	if err != nil {
		return nil, err
	}
	client := s3.NewFromConfig(config, func(options *s3.Options) { options.UsePathStyle = true })
	return &r2Provider{bucket: storageConfig.Bucket, client: client, uploader: manager.NewUploader(client, func(options *manager.Uploader) { options.PartSize = 16 * 1024 * 1024; options.Concurrency = 3 })}, nil
}

func (p *r2Provider) Upload(ctx context.Context, input uploadInput) error {
	request := &s3.PutObjectInput{Bucket: aws.String(p.bucket), Key: aws.String(input.Key), Body: input.Body, ContentLength: aws.Int64(input.Size), ContentType: aws.String(input.ContentType)}
	if input.ContentDisposition != "" {
		request.ContentDisposition = aws.String(input.ContentDisposition)
	}
	if input.CacheControl != "" {
		request.CacheControl = aws.String(input.CacheControl)
	}
	_, err := p.uploader.Upload(ctx, request)
	if err != nil {
		return fmt.Errorf("上传 R2 失败: %w", err)
	}
	return nil
}

func (p *r2Provider) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	result, err := p.client.GetObject(ctx, &s3.GetObjectInput{Bucket: aws.String(p.bucket), Key: aws.String(key)})
	if err != nil {
		return nil, fmt.Errorf("读取 R2 对象失败: %w", err)
	}
	return result.Body, nil
}

func (p *r2Provider) PresignGet(ctx context.Context, key string, ttl time.Duration) (string, error) {
	presigner := s3.NewPresignClient(p.client)
	request, err := presigner.PresignGetObject(ctx, &s3.GetObjectInput{Bucket: aws.String(p.bucket), Key: aws.String(key)}, func(options *s3.PresignOptions) { options.Expires = ttl })
	if err != nil {
		return "", fmt.Errorf("生成 R2 临时读取地址失败: %w", err)
	}
	return request.URL, nil
}
