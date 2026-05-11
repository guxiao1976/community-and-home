package ac

import (
	"sort"
	"strings"
	"sync"

	"github.com/cloudflare/ahocorasick"
)

// WordEntry represents a sensitive word with its category and severity.
type WordEntry struct {
	Word     string
	Category string
	Severity int
}

// MatchResult represents a single match found by the AC automaton.
type MatchResult struct {
	Word     string
	Start    int
	End      int
	Category string
	Severity int
}

// ACMachine wraps the Aho-Corasick automaton with metadata lookup.
type ACMachine struct {
	matcher    *ahocorasick.Matcher
	dictWords  []string           // dictionary words passed to the matcher (no empties)
	dictLookup map[int]*WordEntry // dict index -> WordEntry
	lookup     map[string]*WordEntry
	mu         sync.RWMutex
}

// NewACMachine creates a new empty AC machine.
func NewACMachine() *ACMachine {
	return &ACMachine{
		lookup:     make(map[string]*WordEntry),
		dictLookup: make(map[int]*WordEntry),
	}
}

// Build constructs the Aho-Corasick automaton from the given entries.
// It is thread-safe and acquires a write lock.
func (m *ACMachine) Build(entries []WordEntry) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.buildLocked(entries)
}

// Rebuild atomically replaces the existing automaton with a new one built from entries.
// It is thread-safe and acquires a write lock.
func (m *ACMachine) Rebuild(entries []WordEntry) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.buildLocked(entries)
}

// buildLocked builds the automaton without locking. Caller must hold m.mu.
func (m *ACMachine) buildLocked(entries []WordEntry) error {
	dict := make([]string, 0, len(entries))
	dictLookup := make(map[int]*WordEntry, len(entries))
	lookup := make(map[string]*WordEntry, len(entries))

	dictIdx := 0
	for i := range entries {
		entry := &entries[i]
		if entry.Word == "" {
			continue
		}
		dict = append(dict, entry.Word)
		dictLookup[dictIdx] = entry
		dictIdx++
		// Keep the first entry seen for each word (preserves the original entry).
		if _, exists := lookup[entry.Word]; !exists {
			lookup[entry.Word] = entry
		}
	}

	matcher := ahocorasick.NewStringMatcher(dict)

	m.matcher = matcher
	m.dictWords = dict
	m.dictLookup = dictLookup
	m.lookup = lookup

	return nil
}

// Match scans the text and returns all matches found.
// It performs a single-pass O(n) scan and is read-locked for concurrency.
// Results are sorted by Start position.
func (m *ACMachine) Match(text string) []MatchResult {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.matcher == nil || len(text) == 0 {
		return nil
	}

	indices := m.matcher.MatchThreadSafe([]byte(text))

	results := make([]MatchResult, 0, len(indices))

	for _, idx := range indices {
		entry := m.dictLookup[idx]
		if entry == nil {
			continue
		}
		word := entry.Word

		// Find all occurrences of this word in the text.
		searchStart := 0
		for {
			start := strings.Index(text[searchStart:], word)
			if start == -1 {
				break
			}
			absStart := searchStart + start
			results = append(results, MatchResult{
				Word:     word,
				Start:    absStart,
				End:      absStart + len(word),
				Category: entry.Category,
				Severity: entry.Severity,
			})
			searchStart = absStart + 1
		}
	}

	// Sort by Start position.
	sort.Slice(results, func(i, j int) bool {
		return results[i].Start < results[j].Start
	})

	return results
}

// Contains returns true if any dictionary word is found in text.
func (m *ACMachine) Contains(text string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.matcher == nil || len(text) == 0 {
		return false
	}
	return m.matcher.Contains([]byte(text))
}
