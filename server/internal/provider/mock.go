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
