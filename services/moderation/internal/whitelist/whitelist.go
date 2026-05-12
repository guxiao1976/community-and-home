package whitelist

import (
	"strings"

	"github.com/guxiao/community-and-home/services/moderation/internal/ac"
)

// Whitelist uses an Aho-Corasick machine to efficiently check whether
// text contains any whitelisted (allowed) words. When a whitelist match
// is found, the corresponding content should be treated as compliant.
type Whitelist struct {
	acMachine *ac.ACMachine
}

// NewWhitelist creates a new empty Whitelist.
func NewWhitelist() *Whitelist {
	return &Whitelist{
		acMachine: ac.NewACMachine(),
	}
}

// Build constructs the AC automaton from the given whitelist words.
// This must be called before LongestMatch.
func (w *Whitelist) Build(words []string) {
	entries := make([]ac.WordEntry, len(words))
	for i, word := range words {
		entries[i] = ac.WordEntry{Word: word}
	}
	_ = w.acMachine.Build(entries)
}

// LongestMatch returns the longest matching whitelist word found in text
// and its byte length. If no whitelist word is found, it returns ("", 0).
func (w *Whitelist) LongestMatch(text string) (string, int) {
	results := w.acMachine.Match(text)
	if len(results) == 0 {
		return "", 0
	}

	longest := results[0]
	for _, r := range results[1:] {
		if (r.End - r.Start) > (longest.End - longest.Start) {
			longest = r
		}
	}

	word := longest.Word
	return word, len(word)
}

// Contains returns true if any whitelist word is found in text.
func (w *Whitelist) Contains(text string) bool {
	return w.acMachine.Contains(text)
}

// ContainsAny returns true if text, after stripping whitespace and common
// punctuation, contains any whitelist word.
func (w *Whitelist) ContainsAny(text string) bool {
	if w.Contains(text) {
		return true
	}
	normalized := strings.NewReplacer(
		" ", "", "\t", "", "\n", "", "\r", "",
		",", "", ".", "", "!", "", "?", "",
		";", "", ":", "", "-", "", "_", "",
	).Replace(text)
	return w.Contains(normalized)
}
