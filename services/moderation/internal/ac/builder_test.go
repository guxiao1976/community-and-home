package ac

import "testing"

func TestExpandEntriesWithPinyin(t *testing.T) {
	entries := []WordEntry{
		{Word: "kill", Category: "violence", Severity: 1},
		{Word: "buy", Category: "ads", Severity: 2},
		{Word: "hello", Category: "greeting", Severity: 0},
	}

	// Mock pinyin expand function that generates variants.
	pinyinExpand := func(word string, max int) []string {
		switch word {
		case "kill":
			return []string{"k1ll", "ki11"}
		default:
			return nil
		}
	}

	expanded := ExpandEntriesWithPinyin(entries, pinyinExpand, 20)

	// Should have: original 3 + 2 expansions for "kill" = 5.
	if len(expanded) != 5 {
		t.Errorf("expected 5 entries, got %d: %+v", len(expanded), expanded)
	}

	// First entry should be the original "kill".
	if expanded[0].Word != "kill" {
		t.Errorf("expanded[0].Word = %q, want %q", expanded[0].Word, "kill")
	}

	// "buy" and "hello" should not be expanded (not severity 1).
	foundBuy := false
	foundHello := false
	for _, e := range expanded {
		if e.Word == "buy" && e.Category == "ads" {
			foundBuy = true
		}
		if e.Word == "hello" && e.Category == "greeting" {
			foundHello = true
		}
	}
	if !foundBuy {
		t.Error("original 'buy' entry not found")
	}
	if !foundHello {
		t.Error("original 'hello' entry not found")
	}

	// Expansion entries (k1ll, ki11) should inherit category and severity.
	expandedWords := []string{"k1ll", "ki11"}
	for _, ew := range expandedWords {
		found := false
		for _, e := range expanded {
			if e.Word == ew {
				found = true
				if e.Category != "violence" {
					t.Errorf("expanded entry %q category = %q, want %q", ew, e.Category, "violence")
				}
				if e.Severity != 1 {
					t.Errorf("expanded entry %q severity = %d, want 1", ew, e.Severity)
				}
			}
		}
		if !found {
			t.Errorf("expanded entry %q not found", ew)
		}
	}
}

func TestExpandEntriesWithPinyin_NilFunc(t *testing.T) {
	entries := []WordEntry{
		{Word: "test", Category: "test", Severity: 1},
	}

	result := ExpandEntriesWithPinyin(entries, nil, 20)
	if len(result) != 1 {
		t.Errorf("with nil expand func, expected 1 entry, got %d", len(result))
	}
}

func TestExpandEntriesWithPinyin_DefaultMaxVariants(t *testing.T) {
	entries := []WordEntry{
		{Word: "test", Category: "test", Severity: 1},
	}

	callCount := 0
	pinyinExpand := func(word string, max int) []string {
		callCount++
		if max != 20 {
			t.Errorf("expected default maxVariants=20, got %d", max)
		}
		return nil
	}

	ExpandEntriesWithPinyin(entries, pinyinExpand, 0)
	if callCount != 1 {
		t.Errorf("expected expand to be called once, got %d", callCount)
	}
}

func TestExpandEntriesWithPinyin_SkipsEmptyVariants(t *testing.T) {
	entries := []WordEntry{
		{Word: "test", Category: "test", Severity: 1},
	}

	pinyinExpand := func(word string, max int) []string {
		return []string{"", "variant", "test"} // empty and same-as-original should be skipped.
	}

	result := ExpandEntriesWithPinyin(entries, pinyinExpand, 20)
	// Original + 1 (variant) = 2. Empty and "test" are skipped.
	if len(result) != 2 {
		t.Errorf("expected 2 entries, got %d: %+v", len(result), result)
	}
}
