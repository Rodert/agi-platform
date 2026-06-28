package storage

import (
	"context"
	"io"
	"os"
	"path/filepath"
)

type LocalStore struct {
	Root string
}

func NewLocalStore(root string) *LocalStore {
	if root == "" {
		root = "uploads"
	}
	return &LocalStore{Root: root}
}

func (s *LocalStore) Put(ctx context.Context, key string, body io.Reader, size int64, mimeType string) (Object, error) {
	_ = ctx
	path := filepath.Join(s.Root, filepath.FromSlash(key))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return Object{}, err
	}
	out, err := os.Create(path)
	if err != nil {
		return Object{}, err
	}
	defer out.Close()
	written, err := io.Copy(out, body)
	if err != nil {
		return Object{}, err
	}
	if size <= 0 {
		size = written
	}
	return Object{Key: key, AppURL: "/uploads/" + key, PublicURL: "/uploads/" + key, Provider: "local", MimeType: mimeType, Size: size}, nil
}

func (s *LocalStore) PublicURL(key string) (string, bool) {
	if clean, ok := CleanAssetKey(key); ok {
		return "/uploads/" + clean, true
	}
	return "", false
}
