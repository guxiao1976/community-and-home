package normalize

import "unicode"

// IgnoreCase converts a rune to lowercase using Unicode case folding.
// This handles ASCII as well as non-ASCII letters (e.g., accented characters, Greek, Cyrillic).
//
// Example:
//
//	'Hello World 你好' -> 'hello world 你好'
//	'CAFÉ' -> 'café'
func IgnoreCase(c rune) rune {
	return unicode.ToLower(c)
}
