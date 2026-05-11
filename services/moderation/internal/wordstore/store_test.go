package wordstore

import (
	"testing"
)

func TestNewWordStore_DefaultOptions(t *testing.T) {
	// Note: NewWordStore requires a DB connection, so we test the options
	// and default values that don't need DB
	opts := []WordStoreOption{
		WithSyncInterval(60),
		WithPinyinExpand(false, 10),
		WithSplitSeparators("xX.*"),
	}

	// Verify options can be applied
	ws := &WordStore{}
	for _, opt := range opts {
		opt(ws)
	}

	if ws.syncInterval != 60 {
		t.Errorf("expected syncInterval 60, got %d", ws.syncInterval)
	}
	if ws.pinyinExpand != false {
		t.Error("expected pinyinExpand false")
	}
	if ws.pinyinMax != 10 {
		t.Errorf("expected pinyinMax 10, got %d", ws.pinyinMax)
	}
	if ws.splitSeparators != "xX.*" {
		t.Errorf("expected splitSeparators 'xX.*', got '%s'", ws.splitSeparators)
	}
}

func TestGetNormalizer(t *testing.T) {
	ws := &WordStore{
		syncInterval:    300,
		pinyinExpand:    true,
		pinyinMax:       20,
		splitSeparators: "xX.* 、",
	}

	n := ws.GetNormalizer()
	if n == nil {
		t.Fatal("expected non-nil normalizer")
	}

	// Test that the normalizer works
	text, _ := n.Normalize("ＡＢＣ测试")
	if text == "" {
		t.Error("normalizer should produce output")
	}
}

func TestGetSplitDetector(t *testing.T) {
	ws := &WordStore{
		splitSeparators: "xX.* 、",
	}

	sd := ws.GetSplitDetector()
	if sd == nil {
		t.Fatal("expected non-nil split detector")
	}

	// Test basic split detection
	frags := sd.Detect("敏*感", nil)
	// nil AC machine should return nil
	if frags != nil {
		t.Error("expected nil for nil AC machine")
	}
}

func TestGetSplitDetector_CustomSeparators(t *testing.T) {
	ws := &WordStore{
		splitSeparators: "abc",
	}

	sd := ws.GetSplitDetector()
	if sd == nil {
		t.Fatal("expected non-nil split detector")
	}

	// The split detector should use the custom separators
	// We can't easily test this without an AC machine, but we verify it was created
	_ = sd
}
