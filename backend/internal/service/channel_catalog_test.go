package service

import (
	"encoding/json"
	"testing"
)

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
}
