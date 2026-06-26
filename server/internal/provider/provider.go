package provider

import (
	"context"
	"errors"
	"sync"
)

var ErrProviderNotFound = errors.New("provider adapter not found")

type ImageProvider interface {
	Type() string
	Generate(ctx context.Context, req ImageRequest) (*ImageResult, error)
}

type ImageRequest struct {
	Model          string
	Prompt         string
	NegativePrompt string
	Size           string
	N              int
	UserID         uint64
	APIKey         string
	Extra          map[string]interface{}
}

type ImageResult struct {
	ProviderTaskID string
	Status         string
	Images         []ImageItem
	RawResponse    string
}

type ImageItem struct {
	URL    string
	Width  int
	Height int
}

type Registry struct {
	mu        sync.RWMutex
	providers map[string]ImageProvider
}

func NewRegistry() *Registry {
	return &Registry{
		providers: make(map[string]ImageProvider),
	}
}

func (r *Registry) Register(provider ImageProvider) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.providers[provider.Type()] = provider
}

func (r *Registry) Get(providerType string) (ImageProvider, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	provider, ok := r.providers[providerType]
	if !ok {
		return nil, ErrProviderNotFound
	}
	return provider, nil
}
