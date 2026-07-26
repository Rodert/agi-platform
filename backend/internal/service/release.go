package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"

	apperrors "github.com/javapub/agi-platform-backend/pkg/errors"
	"github.com/redis/go-redis/v9"
)

const (
	releaseCacheKey    = "agi:system:github_releases"
	releaseCacheTTL    = time.Hour
	githubReleasesURL = "https://api.github.com/repos/Rodert/agi-platform/releases?per_page=20"
)

var releaseTagPattern = regexp.MustCompile(`^v?\d+\.\d+\.\d+$`)

// ReleaseInfo is the public release metadata shown in the administration console.
type ReleaseInfo struct {
	Version     string `json:"version"`
	PublishedAt string `json:"published_at"`
}

type githubRelease struct {
	TagName     string    `json:"tag_name"`
	Draft       bool      `json:"draft"`
	Prerelease  bool      `json:"prerelease"`
	PublishedAt time.Time `json:"published_at"`
}

// ReleaseService reads GitHub Releases server-side and shares the result through Redis.
// Regular checks use the cache for one hour; callers can explicitly force a fresh check.
type ReleaseService struct {
	redis  *redis.Client
	client *http.Client
}

func NewReleaseService(redisClient *redis.Client) *ReleaseService {
	return &ReleaseService{
		redis: redisClient,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (s *ReleaseService) List(ctx context.Context, force bool) ([]ReleaseInfo, time.Time, error) {
	if !force {
		if releases, checkedAt, ok := s.cached(ctx); ok {
			return releases, checkedAt, nil
		}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, githubReleasesURL, nil)
	if err != nil {
		return nil, time.Time{}, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "agi-platform-release-checker")
	if token := strings.TrimSpace(os.Getenv("GITHUB_TOKEN")); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	response, err := s.client.Do(req)
	if err != nil {
		s.cache(ctx, nil, time.Now())
		return nil, time.Time{}, apperrors.NewWithDetails(apperrors.ErrCodeInternalServer, "版本检查失败，请稍后重试", err.Error())
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		s.cache(ctx, nil, time.Now())
		return nil, time.Time{}, apperrors.NewWithDetails(apperrors.ErrCodeInternalServer, "版本检查失败，请稍后重试", fmt.Sprintf("GitHub returned HTTP %d", response.StatusCode))
	}

	var source []githubRelease
	if err := json.NewDecoder(response.Body).Decode(&source); err != nil {
		s.cache(ctx, nil, time.Now())
		return nil, time.Time{}, apperrors.NewWithDetails(apperrors.ErrCodeInternalServer, "版本检查失败，请稍后重试", err.Error())
	}

	releases := make([]ReleaseInfo, 0, len(source))
	for _, release := range source {
		if release.Draft || release.Prerelease || !releaseTagPattern.MatchString(release.TagName) {
			continue
		}
		releases = append(releases, ReleaseInfo{
			Version: strings.TrimPrefix(release.TagName, "v"),
			PublishedAt: release.PublishedAt.Format("2006-01-02"),
		})
	}
	sort.Slice(releases, func(i, j int) bool { return compareReleaseVersions(releases[i].Version, releases[j].Version) > 0 })

	checkedAt := time.Now()
	s.cache(ctx, releases, checkedAt)
	return releases, checkedAt, nil
}

func (s *ReleaseService) cache(ctx context.Context, releases []ReleaseInfo, checkedAt time.Time) {
	payload, err := json.Marshal(struct {
		Releases  []ReleaseInfo `json:"releases"`
		CheckedAt time.Time     `json:"checked_at"`
	}{Releases: releases, CheckedAt: checkedAt})
	if err == nil {
		_ = s.redis.Set(ctx, releaseCacheKey, payload, releaseCacheTTL).Err()
	}
}

func (s *ReleaseService) cached(ctx context.Context) ([]ReleaseInfo, time.Time, bool) {
	payload, err := s.redis.Get(ctx, releaseCacheKey).Bytes()
	if err != nil {
		return nil, time.Time{}, false
	}
	var cached struct {
		Releases  []ReleaseInfo `json:"releases"`
		CheckedAt time.Time     `json:"checked_at"`
	}
	if json.Unmarshal(payload, &cached) != nil {
		return nil, time.Time{}, false
	}
	return cached.Releases, cached.CheckedAt, true
}

func compareReleaseVersions(left, right string) int {
	var leftMajor, leftMinor, leftPatch, rightMajor, rightMinor, rightPatch int
	fmt.Sscanf(left, "%d.%d.%d", &leftMajor, &leftMinor, &leftPatch)
	fmt.Sscanf(right, "%d.%d.%d", &rightMajor, &rightMinor, &rightPatch)
	for _, pair := range [][2]int{{leftMajor, rightMajor}, {leftMinor, rightMinor}, {leftPatch, rightPatch}} {
		if pair[0] > pair[1] {
			return 1
		}
		if pair[0] < pair[1] {
			return -1
		}
	}
	return 0
}
