package service

import (
	"encoding/json"
	"testing"

	"github.com/javapub/agi-platform-backend/internal/worker/adapter"
)

func TestGeminiImageParamsConfig(t *testing.T) {
	var pro, flash map[string]modelParamConfig
	if err := json.Unmarshal(geminiImageParamsConfig("gemini-3-pro-image-preview"), &pro); err != nil {
		t.Fatal(err)
	}
	if len(pro["ratio"].Options) != 10 || len(pro["resolution"].Options) != 3 {
		t.Fatalf("unexpected Pro Image config: %#v", pro)
	}
	if err := json.Unmarshal(geminiImageParamsConfig("gemini-2.5-flash-image"), &flash); err != nil {
		t.Fatal(err)
	}
	if len(flash["ratio"].Options) != 9 || len(flash["resolution"].Options) != 0 {
		t.Fatalf("unexpected Flash Image config: %#v", flash)
	}
}

func TestGrokVideoParamsConfig(t *testing.T) {
	var standard map[string]struct {
		Options []struct {
			Value string `json:"value"`
		} `json:"options"`
	}
	if err := json.Unmarshal(grokVideoParamsConfig("grok-image-video"), &standard); err != nil {
		t.Fatal(err)
	}
	if len(standard["ratio"].Options) != 7 || standard["resolution"].Options[0].Value != "720p" || len(standard["duration"].Options) != 3 {
		t.Fatalf("unexpected standard config: %#v", standard)
	}

	var video15 map[string]struct {
		Options []struct {
			Value string `json:"value"`
		} `json:"options"`
	}
	if err := json.Unmarshal(grokVideoParamsConfig("grok-video-1.5"), &video15); err != nil {
		t.Fatal(err)
	}
	if len(video15["ratio"].Options) != 2 || len(video15["duration"].Options) != 6 {
		t.Fatalf("unexpected grok-video-1.5 config: %#v", video15)
	}

	var fast map[string]struct {
		Options []struct {
			Value string `json:"value"`
		} `json:"options"`
	}
	if err := json.Unmarshal(grokVideoParamsConfig("grok-video-1.5fast"), &fast); err != nil {
		t.Fatal(err)
	}
	if len(fast["ratio"].Options) != 2 || len(fast["duration"].Options) != 6 {
		t.Fatalf("unexpected grok-video-1.5fast config: %#v", fast)
	}
}

func TestIsLegacyGrokImageVideo(t *testing.T) {
	if !isLegacyGrokImageVideo("image", adapter.DiscoveredModel{Name: "grok-image-video", Type: "video"}, "grok") {
		t.Fatal("expected legacy Grok image-to-video record to be repairable")
	}
	if isLegacyGrokImageVideo("image", adapter.DiscoveredModel{Name: "other-image-video", Type: "video"}, "grok") {
		t.Fatal("unexpected repair for an unrelated model")
	}
}
