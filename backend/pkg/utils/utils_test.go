package utils

import (
	"strings"
	"testing"
	"unicode/utf8"
)

func TestTruncateStringPreservesUTF8(t *testing.T) {
	input := strings.Repeat("中文", 30)
	result := TruncateString(input, 50)

	if !utf8.ValidString(result) {
		t.Fatal("truncated string must remain valid UTF-8")
	}
	if got := utf8.RuneCountInString(result); got != 50 {
		t.Fatalf("expected 50 characters, got %d", got)
	}
}

func TestTruncateStringHandlesNonPositiveLength(t *testing.T) {
	if got := TruncateString("中文", 0); got != "" {
		t.Fatalf("expected empty string, got %q", got)
	}
}
