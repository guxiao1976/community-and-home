package ac

import (
	"testing"
)

func TestNewACMachine(t *testing.T) {
	m := NewACMachine()
	if m == nil {
		t.Fatal("NewACMachine returned nil")
	}
	if m.lookup == nil {
		t.Fatal("lookup map should be initialized")
	}
}

func TestBuildAndMatch(t *testing.T) {
	m := NewACMachine()

	entries := []WordEntry{
		{Word: "badword", Category: "abuse", Severity: 1},
		{Word: "spam", Category: "spam", Severity: 2},
		{Word: "test", Category: "test", Severity: 0},
	}

	if err := m.Build(entries); err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	tests := []struct {
		name      string
		text      string
		wantCount int
		wantFirst string
	}{
		{
			name:      "single match",
			text:      "this is a badword here",
			wantCount: 1,
			wantFirst: "badword",
		},
		{
			name:      "multiple matches",
			text:      "spam and badword in text",
			wantCount: 2,
			wantFirst: "spam",
		},
		{
			name:      "no match",
			text:      "hello world",
			wantCount: 0,
		},
		{
			name:      "empty text",
			text:      "",
			wantCount: 0,
		},
		{
			name:      "exact match",
			text:      "spam",
			wantCount: 1,
			wantFirst: "spam",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := m.Match(tt.text)
			if len(results) != tt.wantCount {
				t.Errorf("Match(%q) got %d results, want %d: %+v", tt.text, len(results), tt.wantCount, results)
				return
			}
			if tt.wantCount > 0 && results[0].Word != tt.wantFirst {
				t.Errorf("Match(%q) first word = %q, want %q", tt.text, results[0].Word, tt.wantFirst)
			}
		})
	}
}

func TestMatchMetadata(t *testing.T) {
	m := NewACMachine()

	entries := []WordEntry{
		{Word: "kill", Category: "violence", Severity: 1},
		{Word: "buy", Category: "ads", Severity: 2},
	}

	_ = m.Build(entries)

	results := m.Match("kill them and buy now")
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}

	// First result should be "kill" with violence/1.
	if results[0].Word != "kill" {
		t.Errorf("results[0].Word = %q, want %q", results[0].Word, "kill")
	}
	if results[0].Category != "violence" {
		t.Errorf("results[0].Category = %q, want %q", results[0].Category, "violence")
	}
	if results[0].Severity != 1 {
		t.Errorf("results[0].Severity = %d, want 1", results[0].Severity)
	}

	// Second result should be "buy" with ads/2.
	if results[1].Word != "buy" {
		t.Errorf("results[1].Word = %q, want %q", results[1].Word, "buy")
	}
	if results[1].Category != "ads" {
		t.Errorf("results[1].Category = %q, want %q", results[1].Category, "ads")
	}
}

func TestMatchPositions(t *testing.T) {
	m := NewACMachine()

	entries := []WordEntry{
		{Word: "bad", Category: "test", Severity: 1},
	}

	_ = m.Build(entries)

	results := m.Match("a bad example")
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Start != 2 {
		t.Errorf("Start = %d, want 2", results[0].Start)
	}
	if results[0].End != 5 {
		t.Errorf("End = %d, want 5", results[0].End)
	}
}

func TestMatchMultipleOccurrences(t *testing.T) {
	m := NewACMachine()

	entries := []WordEntry{
		{Word: "bad", Category: "test", Severity: 1},
	}

	_ = m.Build(entries)

	results := m.Match("bad bad bad")
	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}
	if results[0].Start != 0 {
		t.Errorf("results[0].Start = %d, want 0", results[0].Start)
	}
	if results[1].Start != 4 {
		t.Errorf("results[1].Start = %d, want 4", results[1].Start)
	}
	if results[2].Start != 8 {
		t.Errorf("results[2].Start = %d, want 8", results[2].Start)
	}
}

func TestRebuild(t *testing.T) {
	m := NewACMachine()

	entries1 := []WordEntry{
		{Word: "old", Category: "old", Severity: 1},
	}
	_ = m.Build(entries1)

	results := m.Match("old word")
	if len(results) != 1 {
		t.Fatalf("initial match: expected 1, got %d", len(results))
	}

	entries2 := []WordEntry{
		{Word: "new", Category: "new", Severity: 1},
	}
	_ = m.Rebuild(entries2)

	results = m.Match("old word")
	if len(results) != 0 {
		t.Errorf("after rebuild, old word should not match, got %d results", len(results))
	}

	results = m.Match("new word")
	if len(results) != 1 {
		t.Errorf("after rebuild, new word should match, got %d results", len(results))
	}
}

func TestBuildSkipsEmptyWords(t *testing.T) {
	m := NewACMachine()

	entries := []WordEntry{
		{Word: "", Category: "empty", Severity: 0},
		{Word: "valid", Category: "test", Severity: 1},
		{Word: "", Category: "empty2", Severity: 0},
	}

	if err := m.Build(entries); err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	results := m.Match("valid")
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Word != "valid" {
		t.Errorf("Word = %q, want %q", results[0].Word, "valid")
	}
}

func TestContains(t *testing.T) {
	m := NewACMachine()

	entries := []WordEntry{
		{Word: "bad", Category: "test", Severity: 1},
	}

	_ = m.Build(entries)

	if !m.Contains("this is bad") {
		t.Error("Contains should return true for text with bad word")
	}
	if m.Contains("this is good") {
		t.Error("Contains should return false for clean text")
	}
	if m.Contains("") {
		t.Error("Contains should return false for empty text")
	}
}

func TestMatchOnNilMatcher(t *testing.T) {
	m := NewACMachine()
	results := m.Match("some text")
	if results != nil {
		t.Errorf("expected nil results for nil matcher, got %+v", results)
	}
}
