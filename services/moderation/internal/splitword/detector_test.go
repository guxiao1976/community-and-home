package splitword

import (
	"testing"

	"github.com/guxiao/community-and-home/services/moderation/internal/ac"
)

func newTestACMachine(words []ac.WordEntry) *ac.ACMachine {
	m := ac.NewACMachine()
	_ = m.Build(words)
	return m
}

func TestDetect_NoSeparators(t *testing.T) {
	detector := NewSplitDetector(" *xX")
	machine := newTestACMachine([]ac.WordEntry{
		{Word: "敏感词", Category: "test", Severity: 1},
	})

	// Normal text without separators should not match through split detection
	// (it would match through direct AC, but split detection re-matches fragments)
	results := detector.Detect("这是正常文本", machine)
	if len(results) > 0 {
		t.Errorf("expected no matches for normal text, got %d", len(results))
	}
}

func TestDetect_SingleCharSeparator(t *testing.T) {
	detector := NewSplitDetector(" *xX")
	machine := newTestACMachine([]ac.WordEntry{
		{Word: "敏感", Category: "test", Severity: 1},
	})

	// "敏 感" — space inserted between characters
	// split produces fragments ["敏", "感"], both 1 char → combined to "敏感"
	results := detector.Detect("敏 感", machine)
	if len(results) == 0 {
		t.Fatal("expected to detect split word '敏感'")
	}
	found := false
	for _, r := range results {
		if r.Word == "敏感" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected '敏感' in results, got: %+v", results)
	}
}

func TestDetect_StarSeparator(t *testing.T) {
	detector := NewSplitDetector(" *xX")
	machine := newTestACMachine([]ac.WordEntry{
		{Word: "傻逼", Category: "abuse", Severity: 1},
	})

	results := detector.Detect("傻*逼", machine)
	if len(results) == 0 {
		t.Fatal("expected to detect '傻*逼'")
	}
	if results[0].Word != "傻逼" {
		t.Errorf("expected '傻逼', got '%s'", results[0].Word)
	}
}

func TestDetect_MultipleSeparators(t *testing.T) {
	detector := NewSplitDetector(" *xX")
	machine := newTestACMachine([]ac.WordEntry{
		{Word: "敏感", Category: "test", Severity: 1},
	})

	// Mixed separators: "敏*x*感" → fragments ["敏", "感"] → combined "敏感"
	results := detector.Detect("敏*x*感", machine)
	if len(results) == 0 {
		t.Fatal("expected to detect split word with mixed separators")
	}
}

func TestDetect_LongFragmentNotCombined(t *testing.T) {
	detector := NewSplitDetector(" *xX")
	machine := newTestACMachine([]ac.WordEntry{
		{Word: "政治敏感", Category: "political", Severity: 1},
	})

	// "政治*敏感" — "政治" is 2 chars (will be combined), "敏感" is 2 chars
	// Both are short enough (<=2) to combine
	results := detector.Detect("政治*敏感", machine)
	if len(results) == 0 {
		t.Fatal("expected to detect '政治敏感' from 2-char fragments")
	}
}

func TestDetect_NoMatch(t *testing.T) {
	detector := NewSplitDetector(" *xX")
	machine := newTestACMachine([]ac.WordEntry{
		{Word: "敏感词", Category: "test", Severity: 1},
	})

	results := detector.Detect("你好*世界", machine)
	if len(results) > 0 {
		t.Errorf("expected no matches, got %d", len(results))
	}
}

func TestDetect_NilACMachine(t *testing.T) {
	detector := NewSplitDetector(" *xX")
	results := detector.Detect("敏 感 词", nil)
	if results != nil {
		t.Errorf("expected nil for nil AC machine, got %d results", len(results))
	}
}

func TestDetect_EmptyText(t *testing.T) {
	detector := NewSplitDetector(" *xX")
	machine := newTestACMachine([]ac.WordEntry{{Word: "test"}})
	results := detector.Detect("", machine)
	if results != nil {
		t.Errorf("expected nil for empty text, got %d results", len(results))
	}
}

func TestDetect_NormalTextUnaffected(t *testing.T) {
	detector := NewSplitDetector(" *xX")
	machine := newTestACMachine([]ac.WordEntry{
		{Word: "敏感词", Category: "test", Severity: 1},
	})

	// Text with no separators and no sensitive content
	results := detector.Detect("今天天气真好", machine)
	if len(results) > 0 {
		t.Errorf("expected no matches for clean text, got %d", len(results))
	}
}

func TestSplit_Simple(t *testing.T) {
	detector := NewSplitDetector(" ")

	frags := detector.split("a b c")
	if len(frags) != 3 {
		t.Fatalf("expected 3 fragments, got %d", len(frags))
	}
	if frags[0].text != "a" || frags[1].text != "b" || frags[2].text != "c" {
		t.Errorf("fragments = %+v", frags)
	}
}

func TestSplit_NoSeparators(t *testing.T) {
	detector := NewSplitDetector(" ")

	frags := detector.split("abc")
	if len(frags) != 1 || frags[0].text != "abc" {
		t.Errorf("expected 1 fragment 'abc', got %+v", frags)
	}
}
