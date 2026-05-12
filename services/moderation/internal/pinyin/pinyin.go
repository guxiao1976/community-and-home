package pinyin

import (
	"strings"

	"github.com/mozillazg/go-pinyin"
)

// Default pinyin args: no tone, normal style.
var defaultArgs = pinyin.NewArgs()

func init() {
	defaultArgs.Style = pinyin.Normal // no tone numbers, plain pinyin
	defaultArgs.Separator = "-"
}

// ToPinyin converts Chinese characters to pinyin, returning a slice of pinyin
// slices (one per character). Non-Chinese characters are returned as-is in a
// single-element sub-slice.
func ToPinyin(hans string) [][]string {
	a := pinyin.Pinyin(hans, defaultArgs)
	return a
}

// ToPinyinStr converts a word to its pinyin string with no spaces and no tones.
// For example, "杀逼" becomes "shabi".
// Non-Chinese characters are passed through as-is.
func ToPinyinStr(word string) string {
	py := ToPinyin(word)
	var b strings.Builder
	for _, group := range py {
		if len(group) == 0 {
			continue
		}
		b.WriteString(group[0])
	}
	return b.String()
}
