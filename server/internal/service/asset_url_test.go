package service

import (
	"context"
	"io"
	"testing"

	"agi-platform/server/internal/storage"
)

type fakeStore struct{}

func (fakeStore) Put(ctx context.Context, key string, body io.Reader, size int64, mimeType string) (storage.Object, error) {
	return storage.Object{}, nil
}

func (fakeStore) PublicURL(key string) (string, bool) {
	return "https://bucket.example.com/" + key, true
}

func TestAssetURLForUpstreamUsesCOSWhenAppBaseIsLocal(t *testing.T) {
	got, err := assetURLForUpstream(fakeStore{}, "http://127.0.0.1:8080", "20260628/references/1/image/a.jpg")
	if err != nil {
		t.Fatal(err)
	}
	want := "https://bucket.example.com/20260628/references/1/image/a.jpg"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestAssetURLForUpstreamUsesPublicAppDomain(t *testing.T) {
	got, err := assetURLForUpstream(fakeStore{}, "https://agi.example.com", "20260628/references/1/image/a.jpg")
	if err != nil {
		t.Fatal(err)
	}
	want := "https://agi.example.com/api/assets/20260628/references/1/image/a.jpg"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestNormalizeReferenceAssetURLRewritesLocalAbsoluteAssetURL(t *testing.T) {
	got, err := normalizeReferenceAssetURL(fakeStore{}, "http://localhost:8080", "http://127.0.0.1:8080/api/assets/20260628/references/1/image/a.jpg")
	if err != nil {
		t.Fatal(err)
	}
	want := "https://bucket.example.com/20260628/references/1/image/a.jpg"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestNormalizeReferenceAssetURLKeepsExternalURL(t *testing.T) {
	const external = "https://cdn.example.com/input.jpg"
	got, err := normalizeReferenceAssetURL(fakeStore{}, "https://agi.example.com", external)
	if err != nil {
		t.Fatal(err)
	}
	if got != external {
		t.Fatalf("got %q, want %q", got, external)
	}
}
