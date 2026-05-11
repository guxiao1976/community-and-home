package whitelist

import (
	"testing"
)

func TestBuild_LongestMatch(t *testing.T) {
	wl := NewWhitelist()
	wl.Build([]string{"白菜", "大白菜", "白"})

	word, length := wl.LongestMatch("今天吃大白菜")
	if word != "大白菜" {
		t.Errorf("expected '大白菜', got '%s'", word)
	}
	if length != len("大白菜") {
		t.Errorf("expected length 3, got %d", length)
	}
}

func TestLongestMatch_NoMatch(t *testing.T) {
	wl := NewWhitelist()
	wl.Build([]string{"白菜"})

	word, length := wl.LongestMatch("今天吃西瓜")
	if word != "" {
		t.Errorf("expected empty, got '%s'", word)
	}
	if length != 0 {
		t.Errorf("expected 0, got %d", length)
	}
}

func TestLongestMatch_EmptyWhitelist(t *testing.T) {
	wl := NewWhitelist()
	wl.Build([]string{})

	word, length := wl.LongestMatch("test")
	if word != "" || length != 0 {
		t.Errorf("expected empty/0 for empty whitelist, got '%s'/%d", word, length)
	}
}

func TestContains(t *testing.T) {
	wl := NewWhitelist()
	wl.Build([]string{"白菜", "萝卜"})

	if !wl.Contains("我喜欢吃白菜") {
		t.Error("expected Contains to return true")
	}
	if wl.Contains("我喜欢吃西瓜") {
		t.Error("expected Contains to return false")
	}
}

func TestContainsAny(t *testing.T) {
	wl := NewWhitelist()
	wl.Build([]string{"白菜", "大白菜"})

	if !wl.ContainsAny("今天 吃 白 菜") {
		t.Error("expected ContainsAny to find match after stripping spaces")
	}
	if !wl.ContainsAny("白,菜") {
		t.Error("expected ContainsAny to find match after stripping commas")
	}
}

func TestWhitelistLongerThanBlacklist(t *testing.T) {
	// Simulates the scenario: blacklist has "白", whitelist has "白菜"
	// The whitelist match (length 2) > blacklist match (length 1), so should skip
	wl := NewWhitelist()
	wl.Build([]string{"白菜"})

	_, wlLen := wl.LongestMatch("白菜汤真好喝")
	if wlLen != len("白菜") {
		t.Errorf("expected whitelist length %d, got %d", len("白菜"), wlLen)
	}

	// If blacklist matched "白" at length len("白"), whitelist "白菜" should be longer
	// This is how text_engine.go uses it
	if wlLen <= len("白") {
		t.Error("whitelist should be longer than blacklist word '白'")
	}
}
