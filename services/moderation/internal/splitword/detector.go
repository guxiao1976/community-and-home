package splitword

import (
	"strings"
	"unicode/utf8"

	"github.com/guxiao/community-and-home/services/moderation/internal/ac"
)

// SplitDetector detects sensitive words that have been obfuscated by
// inserting separator characters between characters of the word
// (e.g., "敏 感 词" or "敏*感*词").
type SplitDetector struct {
	separatorRunes map[rune]bool
}

// NewSplitDetector creates a new SplitDetector that treats the given
// separator characters as obfuscation markers.
func NewSplitDetector(separators string) *SplitDetector {
	sd := &SplitDetector{
		separatorRunes: make(map[rune]bool, len(separators)),
	}
	for _, r := range separators {
		sd.separatorRunes[r] = true
	}
	return sd
}

// Detect removes separator characters from text, collects the remaining
// fragments, and tries to reassemble consecutive short fragments into
// sensitive words by matching against the AC machine.
//
// Strategy:
//  1. Split text by separator characters into fragments with their positions.
//  2. Try matching each individual fragment against the AC machine.
//  3. For consecutive short fragments (1-2 chars each), concatenate them
//     pairwise and try matching the concatenation.
//
// Returns matches mapped to original text positions via ac.MatchResult.
func (d *SplitDetector) Detect(text string, acMachine *ac.ACMachine) []ac.MatchResult {
	if acMachine == nil || len(text) == 0 {
		return nil
	}

	fragments := d.split(text)
	if len(fragments) == 0 {
		return nil
	}

	var results []ac.MatchResult
	seen := make(map[string]bool)

	// Try matching individual fragments.
	for _, frag := range fragments {
		if len(frag.text) == 0 {
			continue
		}
		matches := acMachine.Match(frag.text)
		for _, m := range matches {
			key := m.Word
			if seen[key] {
				continue
			}
			seen[key] = true
			results = append(results, ac.MatchResult{
				Word:     m.Word,
				Start:    frag.start,
				End:      frag.start + len(frag.text),
				Category: m.Category,
				Severity: m.Severity,
			})
		}
	}

	// Try concatenating consecutive short fragments (1-2 chars each).
	for i := 0; i < len(fragments)-1; i++ {
		curr := fragments[i]
		next := fragments[i+1]
		if len(curr.text) == 0 || len(next.text) == 0 {
			continue
		}
		// Only combine short fragments to avoid excessive false positives.
		if len([]rune(curr.text)) > 2 || len([]rune(next.text)) > 2 {
			continue
		}

		combined := curr.text + next.text
		matches := acMachine.Match(combined)
		for _, m := range matches {
			key := m.Word
			if seen[key] {
				continue
			}
			seen[key] = true
			results = append(results, ac.MatchResult{
				Word:     m.Word,
				Start:    curr.start,
				End:      next.start + len(next.text),
				Category: m.Category,
				Severity: m.Severity,
			})
		}
	}

	return results
}

// fragment represents a piece of text between separators.
type fragment struct {
	text  string
	start int // byte offset in the original text
}

// split breaks text into fragments, stripping out separator characters.
// Each fragment records its start position in the original text (byte offset).
func (d *SplitDetector) split(text string) []fragment {
	var fragments []fragment
	var buf strings.Builder
	fragStart := -1

	// Iterate byte-by-byte to correctly track byte offsets.
	for bytePos := 0; bytePos < len(text); {
		r, size := utf8.DecodeRuneInString(text[bytePos:])
		bytePos += size

		if d.separatorRunes[r] {
			// Flush the current fragment when we hit a separator.
			if buf.Len() > 0 {
				fragments = append(fragments, fragment{
					text:  buf.String(),
					start: fragStart,
				})
				buf.Reset()
				fragStart = -1
			}
			continue
		}

		if buf.Len() == 0 {
			// Record the byte offset of this fragment's start.
			fragStart = bytePos - size
		}
		buf.WriteRune(r)
	}

	// Flush the last fragment.
	if buf.Len() > 0 {
		fragments = append(fragments, fragment{
			text:  buf.String(),
			start: fragStart,
		})
	}

	return fragments
}
