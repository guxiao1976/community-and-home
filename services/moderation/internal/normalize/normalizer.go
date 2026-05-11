package normalize

import (
	"strings"
)

// Func represents a single rune-to-rune normalization transformation.
type Func func(rune) rune

// Normalizer applies a chain of rune transformations to normalize text.
// It maintains position mapping from the normalized string back to the original.
type Normalizer struct {
	funcs []Func
}

// Option configures a Normalizer.
type Option func(*Normalizer)

// New creates a Normalizer with the given options applied in order.
// Each option appends a rune transformation function to the chain.
func New(opts ...Option) *Normalizer {
	n := &Normalizer{}
	for _, opt := range opts {
		opt(n)
	}
	return n
}

// WithWidth appends fullwidth-to-halfwidth conversion to the normalization chain.
func WithWidth() Option {
	return func(n *Normalizer) {
		n.funcs = append(n.funcs, IgnoreWidth)
	}
}

// WithCase appends case folding (lowercase) to the normalization chain.
func WithCase() Option {
	return func(n *Normalizer) {
		n.funcs = append(n.funcs, IgnoreCase)
	}
}

// WithChinese appends traditional-to-simplified Chinese conversion to the normalization chain.
func WithChinese() Option {
	return func(n *Normalizer) {
		n.funcs = append(n.funcs, IgnoreChinese)
	}
}

// WithNumStyle appends digit style normalization to the normalization chain.
func WithNumStyle() Option {
	return func(n *Normalizer) {
		n.funcs = append(n.funcs, IgnoreNumStyle)
	}
}

// WithEnglishStyle appends English letter style normalization to the normalization chain.
func WithEnglishStyle() Option {
	return func(n *Normalizer) {
		n.funcs = append(n.funcs, IgnoreEnglishStyle)
	}
}

// WithRepeat sets up repeat compression (3+ consecutive same chars -> 1 char).
// Unlike other options, this operates on the full string, not per-character.
// It must be handled specially in Normalize.
func WithRepeat() Option {
	return func(n *Normalizer) {
		n.funcs = append(n.funcs, nil) // sentinel; handled in Normalize
	}
}

// hasRepeat returns true if WithRepeat was configured.
func (n *Normalizer) hasRepeat() bool {
	for _, f := range n.funcs {
		if f == nil {
			return true
		}
	}
	return false
}

// charFuncs returns only the non-nil rune transformation functions.
func (n *Normalizer) charFuncs() []Func {
	var funcs []Func
	for _, f := range n.funcs {
		if f != nil {
			funcs = append(funcs, f)
		}
	}
	return funcs
}

// NormalizeChar applies all rune-level transformation functions to a single rune.
func (n *Normalizer) NormalizeChar(c rune) rune {
	for _, f := range n.charFuncs() {
		c = f(c)
	}
	return c
}

// Normalize transforms a full text string and returns the normalized result
// along with a position mapping. posMap[i] gives the index in the original
// string where normalized character i originated.
//
// When WithRepeat is enabled, consecutive runs of 3+ identical characters
// (after other normalizations) are compressed to a single character. The
// posMap entry for the compressed character points to the first original
// character in the run.
func (n *Normalizer) Normalize(text string) (normalized string, posMap []int) {
	if len(n.funcs) == 0 {
		return text, buildIdentityPosMap(text)
	}

	runes := []rune(text)
	result := make([]rune, 0, len(runes))
	positions := make([]int, 0, len(runes))

	for i, r := range runes {
		nr := n.NormalizeChar(r)
		result = append(result, nr)
		positions = append(positions, i)
	}

	// Apply repeat compression if configured
	if n.hasRepeat() {
		result, positions = compressRepeat(result, positions)
	}

	return string(result), positions
}

// buildIdentityPosMap creates a position map where posMap[i] == i for each byte position.
func buildIdentityPosMap(text string) []int {
	pm := make([]int, 0, len(text))
	for i := 0; i < len(text); i++ {
		pm = append(pm, i)
	}
	return pm
}

// compressRepeat compresses runs of 3+ identical runes into a single rune.
// The position map entry for the surviving rune points to the first rune in the run.
func compressRepeat(runes []rune, positions []int) ([]rune, []int) {
	if len(runes) == 0 {
		return runes, positions
	}

	result := make([]rune, 0, len(runes))
	pm := make([]int, 0, len(positions))

	i := 0
	for i < len(runes) {
		j := i + 1
		for j < len(runes) && runes[j] == runes[i] {
			j++
		}
		runLen := j - i

		if runLen >= 3 {
			// Compress: keep only the first character
			result = append(result, runes[i])
			pm = append(pm, positions[i])
		} else {
			// Keep all characters in short runs
			for k := i; k < j; k++ {
				result = append(result, runes[k])
				pm = append(pm, positions[k])
			}
		}
		i = j
	}

	return result, pm
}

// NormalizeFull runs a full normalization with all options enabled.
// This is a convenience function equivalent to New(WithWidth, WithCase, WithChinese, WithNumStyle, WithEnglishStyle, WithRepeat).Normalize(text).
func NormalizeFull(text string) (string, []int) {
	return New(WithWidth(), WithCase(), WithChinese(), WithNumStyle(), WithEnglishStyle(), WithRepeat()).Normalize(text)
}

// MapNormalizedToOriginal maps a position in the normalized string back to
// the corresponding position in the original string using the position map.
// If the position is out of bounds, it returns -1.
func MapNormalizedToOriginal(posMap []int, normalizedPos int) int {
	if normalizedPos < 0 || normalizedPos >= len(posMap) {
		return -1
	}
	return posMap[normalizedPos]
}

// MapOriginalToNormalized maps a position in the original string to the
// corresponding position in the normalized string using the position map.
// If the original position is not found, it returns -1.
func MapOriginalToNormalized(posMap []int, originalPos int) int {
	for i, p := range posMap {
		if p == originalPos {
			return i
		}
	}
	return -1
}

// IsNormalizedEqual checks if two strings are equal after normalization
// with the given options.
func IsNormalizedEqual(a, b string, opts ...Option) bool {
	na, _ := New(opts...).Normalize(a)
	nb, _ := New(opts...).Normalize(b)
	return na == nb
}

// NormalizeAndSearch normalizes the haystack text and returns whether
// the normalized needle can be found in it.
func NormalizeAndSearch(haystack, needle string, opts ...Option) bool {
	nh, _ := New(opts...).Normalize(haystack)
	nn, _ := New(opts...).Normalize(needle)
	return strings.Contains(nh, nn)
}
