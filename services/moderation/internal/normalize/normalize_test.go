package normalize

import (
	"strings"
	"testing"
)

func TestIgnoreWidth(t *testing.T) {
	tests := []struct {
		input    rune
		expected rune
	}{
		{'Ａ', 'A'}, {'Ｂ', 'B'}, {'Ｃ', 'C'}, {'ａ', 'a'}, {'ｂ', 'b'},
		{'１', '1'}, {'２', '2'}, {'０', '0'}, {'９', '9'},
		{'！', '!'}, {'＠', '@'}, {'＃', '#'}, {'　', ' '}, // U+3000 fullwidth space
		{'A', 'A'}, // ASCII passthrough
		{'中', '中'}, // CJK passthrough
	}
	for _, tc := range tests {
		got := IgnoreWidth(tc.input)
		if got != tc.expected {
			t.Errorf("IgnoreWidth(%q U+%04X) = %q (U+%04X), want %q (U+%04X)",
				tc.input, tc.input, got, got, tc.expected, tc.expected)
		}
	}
}

func TestIgnoreCase(t *testing.T) {
	tests := []struct {
		input    rune
		expected rune
	}{
		{'A', 'a'}, {'Z', 'z'}, {'a', 'a'}, {'É', 'é'}, {'Ω', 'ω'},
		{'中', '中'}, // CJK passthrough
	}
	for _, tc := range tests {
		got := IgnoreCase(tc.input)
		if got != tc.expected {
			t.Errorf("IgnoreCase(%q) = %q, want %q", tc.input, got, tc.expected)
		}
	}
}

func TestIgnoreChinese(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"龜醫開關長門鐵電國黨廣廠壓歲塵僕價優償險隸雙雜離難雲霧靈靜韓韻響頁頂領",
			"龟医开关长门铁电国党广厂压岁尘仆价优偿险隶双杂离难云雾灵静韩韵响页顶领"},
		{"繁體中文", "繁体中文"},
		{"台灣", "台湾"},
		{"壓歲", "压岁"},
		{"關門", "关门"},
		{"頂部", "顶部"},
		{"靜電", "静电"},
		{"身體", "身体"},
		{"歡迎", "欢迎"},
		{"國家", "国家"},
		{"電影", "电影"},
		{"學習", "学习"},
		{"Traditional", "Traditional"}, // ASCII passthrough
		{"简体", "简体"}, // Already simplified, passthrough
	}
	for _, tc := range tests {
		var got strings.Builder
		for _, r := range tc.input {
			got.WriteRune(IgnoreChinese(r))
		}
		if got.String() != tc.expected {
			t.Errorf("IgnoreChinese(%q) = %q, want %q", tc.input, got.String(), tc.expected)
		}
	}
}

func TestIgnoreChineseMapSize(t *testing.T) {
	if len(traditionalToSimplified) < 200 {
		t.Errorf("traditionalToSimplified has %d entries, want at least 200", len(traditionalToSimplified))
	}
	t.Logf("traditionalToSimplified has %d entries", len(traditionalToSimplified))
}

func TestIgnoreNumStyle(t *testing.T) {
	tests := []struct {
		input    rune
		expected rune
	}{
		// Fullwidth
		{'１', '1'}, {'０', '0'},
		// Circled
		{'①', '1'}, {'②', '2'}, {'⑨', '9'}, {'⓪', '0'},
		// Chinese digits
		{'一', '1'}, {'二', '2'}, {'三', '3'}, {'九', '9'},
		// Formal Chinese
		{'壹', '1'}, {'贰', '2'}, {'玖', '9'},
		// Superscript
		{'¹', '1'}, {'²', '2'}, {'⁹', '9'},
		// Subscript
		{'₀', '0'}, {'₉', '9'},
		// Roman
		{'Ⅰ', '1'}, {'Ⅸ', '9'},
	}
	for _, tc := range tests {
		got := IgnoreNumStyle(tc.input)
		if got != tc.expected {
			t.Errorf("IgnoreNumStyle(%q U+%04X) = %q, want %q", tc.input, tc.input, got, tc.expected)
		}
	}
}

func TestIgnoreEnglishStyle(t *testing.T) {
	tests := []struct {
		input    rune
		expected rune
	}{
		// Circled uppercase
		{'Ⓐ', 'A'}, {'Ⓑ', 'B'}, {'Ⓩ', 'Z'},
		// Circled lowercase
		{'ⓐ', 'a'}, {'ⓑ', 'b'}, {'ⓩ', 'z'},
		// Small circled
		{'⒜', 'A'}, {'⒵', 'Z'},
		// Fullwidth (should also be handled)
		{'Ａ', 'A'}, {'ａ', 'a'},
		// Passthrough
		{'A', 'A'}, {'a', 'a'},
	}
	for _, tc := range tests {
		got := IgnoreEnglishStyle(tc.input)
		if got != tc.expected {
			t.Errorf("IgnoreEnglishStyle(%q) = %q, want %q", tc.input, got, tc.expected)
		}
	}
}

func TestCompressRepeat(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"heeeeello", "hello"},
		{"abc", "abc"},
		{"aabbcc", "aabbcc"},
		{"woooow", "wow"},
		{"", ""},
		{"x", "x"},
		{"aa", "aa"},
		{"aaa", "a"},
		{"啊啊啊啊啊", "啊"},
		{"哈哈嘻嘻嘻嘻嘻", "哈哈嘻"},
	}
	for _, tc := range tests {
		got := CompressRepeat(tc.input)
		if got != tc.expected {
			t.Errorf("CompressRepeat(%q) = %q, want %q", tc.input, got, tc.expected)
		}
	}
}

func TestNormalizer_NormalizeChar(t *testing.T) {
	n := New(WithWidth(), WithCase())
	tests := []struct {
		input    rune
		expected rune
	}{
		{'Ａ', 'a'}, // fullwidth + uppercase
		{'ａ', 'a'}, // fullwidth
		{'A', 'a'}, // uppercase
		{'a', 'a'}, // already normalized
	}
	for _, tc := range tests {
		got := n.NormalizeChar(tc.input)
		if got != tc.expected {
			t.Errorf("NormalizeChar(%q) = %q, want %q", tc.input, got, tc.expected)
		}
	}
}

func TestNormalizer_Normalize(t *testing.T) {
	n := New(WithWidth(), WithCase())
	text := "Ｈｅｌｌｏ Ｗｏｒｌｄ"
	normalized, posMap := n.Normalize(text)
	if normalized != "hello world" {
		t.Errorf("Normalize(%q) = %q, want %q", text, normalized, "hello world")
	}
	if len(posMap) != len("hello world") {
		t.Errorf("posMap length = %d, want %d", len(posMap), len("hello world"))
	}
}

func TestNormalizer_WithRepeat(t *testing.T) {
	n := New(WithCase(), WithRepeat())
	text := "HEEEEELLO"
	normalized, _ := n.Normalize(text)
	if normalized != "hello" {
		t.Errorf("Normalize(%q) = %q, want %q", text, normalized, "helo")
	}
}

func TestNormalizer_PositionMapping(t *testing.T) {
	n := New(WithWidth())
	text := "ＡＢＣ"
	normalized, posMap := n.Normalize(text)
	if normalized != "ABC" {
		t.Fatalf("Normalize(%q) = %q, want %q", text, normalized, "ABC")
	}
	// Each normalized char should map to its original rune index
	for i := 0; i < len(posMap); i++ {
		if posMap[i] != i {
			t.Errorf("posMap[%d] = %d, want %d", i, posMap[i], i)
		}
	}
}

func TestNormalizeFull(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"ＨＥＬＬＯ", "hello"},           // fullwidth + case (no repeat needed)
		{"ＡＢＣ", "abc"},                 // fullwidth + case
		{"龜", "龟"},                     // traditional chinese
		{"①②③", "123"},                  // num style
		{"ⒶⒷⒸ", "abc"},                 // english style (circled) + case
		{"ＡａＢｂ", "aabb"},             // fullwidth + case (2 repeats each, not compressed)
		{"ＡａＡａＡａ", "a"},             // fullwidth + case + repeat (3+ compressed)
	}
	for _, tc := range tests {
		got, _ := NormalizeFull(tc.input)
		if got != tc.expected {
			t.Errorf("NormalizeFull(%q) = %q, want %q", tc.input, got, tc.expected)
		}
	}
}

func TestIsNormalizedEqual(t *testing.T) {
	tests := []struct {
		a, b     string
		opts     []Option
		expected bool
	}{
		{"ＡＢＣ", "abc", []Option{WithWidth(), WithCase()}, true},
		{"ＡＢＣ", "ABC", []Option{WithWidth()}, true},
		{"abc", "ABC", []Option{WithCase()}, true},
		{"國家", "国家", []Option{WithChinese()}, true},
		{"hello", "helo", []Option{WithRepeat()}, false},
		{"heeeeello", "hello", []Option{WithRepeat()}, true},
	}
	for _, tc := range tests {
		got := IsNormalizedEqual(tc.a, tc.b, tc.opts...)
		if got != tc.expected {
			t.Errorf("IsNormalizedEqual(%q, %q) = %v, want %v", tc.a, tc.b, got, tc.expected)
		}
	}
}

func TestNormalizeAndSearch(t *testing.T) {
	tests := []struct {
		haystack string
		needle   string
		opts     []Option
		expected bool
	}{
		{"Ｈｅｌｌｏ", "Hello", []Option{WithWidth()}, true},
		{"Ｈｅｌｌｏ", "hello", []Option{WithWidth()}, false},
		{"Ｈｅｌｌｏ", "hello", []Option{WithWidth(), WithCase()}, true},
	}
	for _, tc := range tests {
		got := NormalizeAndSearch(tc.haystack, tc.needle, tc.opts...)
		if got != tc.expected {
			t.Errorf("NormalizeAndSearch(%q, %q) = %v, want %v", tc.haystack, tc.needle, got, tc.expected)
		}
	}
}

func TestMapNormalizedToOriginal(t *testing.T) {
	_, posMap := New(WithWidth()).Normalize("ＡＢＣ")
	tests := []struct {
		pos      int
		expected int
	}{
		{0, 0}, {1, 1}, {2, 2},
		{-1, -1}, {3, -1},
	}
	for _, tc := range tests {
		got := MapNormalizedToOriginal(posMap, tc.pos)
		if got != tc.expected {
			t.Errorf("MapNormalizedToOriginal(posMap, %d) = %d, want %d", tc.pos, got, tc.expected)
		}
	}
}

func TestNormalizer_Empty(t *testing.T) {
	n := New(WithWidth(), WithCase(), WithChinese())
	normalized, posMap := n.Normalize("")
	if normalized != "" {
		t.Errorf("Normalize(%q) = %q, want empty", "", normalized)
	}
	if len(posMap) != 0 {
		t.Errorf("posMap length = %d, want 0", len(posMap))
	}
}
