package pinyin

import (
	"testing"
)

func TestToPinyin(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantLen  int
		check    func(t *testing.T, result [][]string)
	}{
		{
			name:    "single chinese char",
			input:   "你",
			wantLen: 1,
			check: func(t *testing.T, result [][]string) {
				if len(result[0]) == 0 {
					t.Error("expected pinyin for 你")
				}
			},
		},
		{
			name:    "multiple chinese chars",
			input:   "你好",
			wantLen: 2,
			check: func(t *testing.T, result [][]string) {
				if len(result[0]) == 0 || len(result[1]) == 0 {
					t.Error("expected pinyin for both characters")
				}
			},
		},
		{
			name:    "empty string",
			input:   "",
			wantLen: 0,
		},
		{
			name:    "non-chinese",
			input:   "abc",
			wantLen: 0, // go-pinyin returns empty for non-Chinese
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToPinyin(tt.input)
			if len(result) != tt.wantLen {
				t.Errorf("ToPinyin(%q) len = %d, want %d", tt.input, len(result), tt.wantLen)
			}
			if tt.check != nil {
				tt.check(t, result)
			}
		})
	}
}

func TestToPinyinStr(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "chinese word",
			input: "你好",
			want:  "nihao",
		},
		{
			name:  "single char",
			input: "好",
			want:  "hao",
		},
		{
			name:  "empty",
			input: "",
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToPinyinStr(tt.input)
			if result != tt.want {
				t.Errorf("ToPinyinStr(%q) = %q, want %q", tt.input, result, tt.want)
			}
		})
	}
}
