package provider

import (
	"context"
	"fmt"
	"time"
)

type MockProvider struct{}

func NewMockProvider() *MockProvider {
	return &MockProvider{}
}

func (p *MockProvider) Type() string {
	return "mock"
}

func (p *MockProvider) Generate(_ context.Context, req ImageRequest) (*ImageResult, error) {
	now := time.Now().UnixNano()
	return &ImageResult{
		ProviderTaskID: fmt.Sprintf("mock_%d", now),
		Status:         "succeeded",
		Images: []ImageItem{
			{
				URL:    fmt.Sprintf("https://example.com/mock/%d.png", now),
				Width:  1024,
				Height: 1024,
			},
		},
		RawResponse: fmt.Sprintf(`{"provider":"mock","model":%q,"prompt":%q}`, req.Model, req.Prompt),
	}, nil
}

func (p *MockProvider) CreateVideo(_ context.Context, req VideoRequest) (*VideoCreateResult, error) {
	now := time.Now().UnixNano()
	return &VideoCreateResult{
		TaskID:      fmt.Sprintf("mock_video_%d", now),
		Status:      "running",
		RawResponse: fmt.Sprintf(`{"provider":"mock","model":%q,"prompt":%q}`, req.Model, req.Prompt),
	}, nil
}

func (p *MockProvider) GetVideo(_ context.Context, req VideoStatusRequest) (*VideoStatusResult, error) {
	return &VideoStatusResult{
		TaskID:      req.TaskID,
		Status:      "succeeded",
		Progress:    100,
		RawResponse: fmt.Sprintf(`{"provider":"mock","task_id":%q,"status":"succeeded"}`, req.TaskID),
	}, nil
}

func (p *MockProvider) DownloadVideo(_ context.Context, _ VideoContentRequest) ([]byte, string, error) {
	return []byte("mock video content"), "video/mp4", nil
}
