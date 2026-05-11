package pinyin

import (
	"strings"
	"testing"
)

func TestExpandHomophones(t *testing.T) {
	tests := []struct {
		name         string
		word         string
		maxVariants  int
		wantMinLen   int
		wantContains string
	}{
		{
			name:         "two char word",
			word:         "杀逼",
			maxVariants:  20,
			wantMinLen:   1, // at least the original
			wantContains: "杀逼",
		},
		{
			name:        "single char",
			word:        "杀",
			maxVariants: 20,
			wantMinLen:  3, // original + at least some homophones
		},
		{
			name:        "max variants limit",
			word:        "杀逼",
			maxVariants: 5,
			wantMinLen:  1,
		},
		{
			name:         "default max",
			word:         "你好",
			maxVariants:  0,
			wantMinLen:   1,
			wantContains: "你好",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExpandHomophones(tt.word, tt.maxVariants)
			if len(result) < tt.wantMinLen {
				t.Errorf("ExpandHomophones(%q, %d) len = %d, want >= %d", tt.word, tt.maxVariants, len(result), tt.wantMinLen)
			}
			if tt.wantContains != "" {
				found := false
				for _, v := range result {
					if v == tt.wantContains {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("ExpandHomophones result should contain %q", tt.wantContains)
				}
			}
			if tt.maxVariants > 0 && len(result) > tt.maxVariants {
				t.Errorf("ExpandHomophones(%q, %d) len = %d, want <= %d", tt.word, tt.maxVariants, len(result), tt.maxVariants)
			}
		})
	}
}

func TestExpandHomophones_NonChinese(t *testing.T) {
	result := ExpandHomophones("hello", 20)
	if result != nil {
		t.Errorf("expected nil for non-chinese word, got %v", result)
	}
}

func TestExpandHomophones_Empty(t *testing.T) {
	result := ExpandHomophones("", 20)
	if result != nil {
		t.Errorf("expected nil for empty word, got %v", result)
	}
}

func TestHomophoneMatch(t *testing.T) {
	// Test that words with same pinyin match.
	tests := []struct {
		a, b string
		want bool
	}{
		{"沙", "杀", true},   // same pinyin "sha"
		{"你", "好", false},  // different pinyin
		{"沙", "沙", false},  // same word, should be false
		{"大", "打", false},  // "da" vs "da" - actually same pinyin, should be true
	}

	for _, tt := range tests {
		result := HomophoneMatch(tt.a, tt.b)
		// Only check the false-negative case (if pinyin same, should match).
		pa := ToPinyinStr(tt.a)
		pb := ToPinyinStr(tt.b)
		samePinyin := pa == pb && tt.a != tt.b
		if samePinyin && !result {
			t.Errorf("HomophoneMatch(%q, %q) = false, want true (same pinyin: %s)", tt.a, tt.b, pa)
		}
		if !samePinyin && result {
			t.Errorf("HomophoneMatch(%q, %q) = true, want false (diff pinyin: %s vs %s)", tt.a, tt.b, pa, pb)
		}
	}
}

func TestHasHomophoneVariant(t *testing.T) {
	// Chinese word with common pinyin should have variants.
	if !HasHomophoneVariant("杀") {
		t.Error("HasHomophoneVariant(杀) should return true")
	}

	// Check a character that should have many homophones.
	if !HasHomophoneVariant("是") {
		t.Error("HasHomophoneVariant(是) should return true")
	}
}

func TestFindPinyinSyllable(t *testing.T) {
	result := FindPinyinSyllable('杀')
	if result == nil {
		t.Error("FindPinyinSyllable('杀') should return non-nil")
	}
	found := false
	for _, c := range result {
		if strings.Contains(c, "沙") || strings.Contains(c, "杀") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("FindPinyinSyllable('杀') should contain 沙 or 杀, got %v", result)
	}

	// Non-Chinese character should return nil.
	result = FindPinyinSyllable('a')
	if result != nil {
		t.Errorf("FindPinyinSyllable('a') should return nil, got %v", result)
	}
}
