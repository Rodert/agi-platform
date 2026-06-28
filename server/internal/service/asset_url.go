package service

import (
	"fmt"
	"net"
	"net/url"
	"strings"

	"agi-platform/server/internal/storage"
)

const appAssetPrefix = "/api/assets/"

func assetURLForUpstream(store storage.Store, appBaseURL string, key string) (string, error) {
	clean, ok := storage.CleanAssetKey(key)
	if !ok {
		return "", fmt.Errorf("%w: invalid asset path", ErrInvalidRequest)
	}
	if baseURL := publicAppBaseURL(appBaseURL); baseURL != "" {
		return baseURL + appAssetPrefix + clean, nil
	}
	publicURL, ok := store.PublicURL(clean)
	if !ok || !isAbsoluteHTTPURL(publicURL) {
		return "", fmt.Errorf("%w: asset is not publicly accessible", ErrInvalidRequest)
	}
	return publicURL, nil
}

func normalizeReferenceAssetURL(store storage.Store, appBaseURL string, value string) (string, error) {
	trimmed := strings.TrimSpace(value)
	if strings.HasPrefix(trimmed, appAssetPrefix) {
		return assetURLForUpstream(store, appBaseURL, strings.TrimPrefix(trimmed, appAssetPrefix))
	}

	parsed, err := url.Parse(trimmed)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return trimmed, nil
	}
	if !strings.EqualFold(parsed.Scheme, "http") && !strings.EqualFold(parsed.Scheme, "https") {
		return trimmed, nil
	}
	if !strings.HasPrefix(parsed.EscapedPath(), appAssetPrefix) && !strings.HasPrefix(parsed.Path, appAssetPrefix) {
		return trimmed, nil
	}

	key := strings.TrimPrefix(parsed.Path, appAssetPrefix)
	if isLocalHost(parsed.Hostname()) {
		return assetURLForUpstream(store, "", key)
	}
	if baseURL := publicAppBaseURL(appBaseURL); baseURL != "" {
		return baseURL + appAssetPrefix + key, nil
	}
	return trimmed, nil
}

func publicAppBaseURL(value string) string {
	trimmed := strings.TrimRight(strings.TrimSpace(value), "/")
	if trimmed == "" {
		return ""
	}
	parsed, err := url.Parse(trimmed)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" || isLocalHost(parsed.Hostname()) {
		return ""
	}
	if !strings.EqualFold(parsed.Scheme, "http") && !strings.EqualFold(parsed.Scheme, "https") {
		return ""
	}
	return parsed.Scheme + "://" + parsed.Host
}

func isAbsoluteHTTPURL(value string) bool {
	parsed, err := url.Parse(strings.TrimSpace(value))
	return err == nil && parsed.Host != "" && (strings.EqualFold(parsed.Scheme, "http") || strings.EqualFold(parsed.Scheme, "https"))
}

func isLocalHost(host string) bool {
	normalized := strings.Trim(strings.ToLower(host), "[]")
	if normalized == "localhost" || normalized == "0.0.0.0" || normalized == "::1" {
		return true
	}
	if strings.HasSuffix(normalized, ".localhost") {
		return true
	}
	ip := net.ParseIP(normalized)
	return ip != nil && ip.IsLoopback()
}
