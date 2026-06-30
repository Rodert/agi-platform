package provider

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

var ErrProviderNotFound = errors.New("provider adapter not found")

type ResponseError struct {
	Message     string
	RawResponse string
}

func (e *ResponseError) Error() string {
	return e.Message
}

func NewResponseError(message string, rawResponse string) error {
	return &ResponseError{Message: message, RawResponse: rawResponse}
}

func ResponseErrorRaw(err error) string {
	var responseErr *ResponseError
	if errors.As(err, &responseErr) {
		return responseErr.RawResponse
	}
	return ""
}

func NewHTTPResponseError(label string, statusCode int, rawResponse string) error {
	return NewResponseError(fmt.Sprintf("%s: status=%d body=%s", label, statusCode, truncateForError(rawResponse, 1000)), rawResponse)
}

func truncateForError(value string, limit int) string {
	if len(value) <= limit {
		return value
	}
	return value[:limit] + "..."
}

type ImageProvider interface {
	Type() string
	Generate(ctx context.Context, req ImageRequest) (*ImageResult, error)
}

type ImageCapabilityProvider interface {
	SupportsImageGeneration() bool
}

type VideoProvider interface {
	Type() string
	CreateVideo(ctx context.Context, req VideoRequest) (*VideoCreateResult, error)
	GetVideo(ctx context.Context, req VideoStatusRequest) (*VideoStatusResult, error)
	DownloadVideo(ctx context.Context, req VideoContentRequest) ([]byte, string, error)
}

type ImageRequest struct {
	Model           string
	Prompt          string
	NegativePrompt  string
	Size            string
	N               int
	UserID          uint64
	BaseURL         string
	APIKey          string
	TimeoutSeconds  int
	ReferenceImages []ReferenceImage
	Extra           map[string]interface{}
}

type ReferenceImage struct {
	URL      string
	DataURL  string
	Filename string
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

type VideoRequest struct {
	Model          string
	Prompt         string
	Seconds        int
	AspectRatio    string
	Images         []string
	Videos         []string
	Audios         []string
	UserID         uint64
	BaseURL        string
	APIKey         string
	TimeoutSeconds int
	Extra          map[string]interface{}
}

type VideoStatusRequest struct {
	TaskID         string
	BaseURL        string
	APIKey         string
	TimeoutSeconds int
}

type VideoContentRequest struct {
	TaskID         string
	BaseURL        string
	APIKey         string
	TimeoutSeconds int
}

type VideoCreateResult struct {
	TaskID      string
	Status      string
	URL         string
	RawResponse string
}

type VideoStatusResult struct {
	TaskID       string
	Status       string
	Progress     int
	URL          string
	ErrorCode    string
	ErrorMessage string
	RawResponse  string
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

func SupportsImageGeneration(provider ImageProvider) bool {
	if capable, ok := provider.(ImageCapabilityProvider); ok {
		return capable.SupportsImageGeneration()
	}
	return true
}

func (r *Registry) GetVideo(providerType string) (VideoProvider, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	provider, ok := r.providers[providerType]
	if !ok {
		return nil, ErrProviderNotFound
	}
	videoProvider, ok := provider.(VideoProvider)
	if !ok {
		return nil, ErrProviderNotFound
	}
	return videoProvider, nil
}
