package normalize

// IgnoreWidth converts fullwidth Unicode characters to their halfwidth ASCII equivalents.
//
// Fullwidth characters in the range U+FF01 to U+FF5E are mapped to their halfwidth
// equivalents U+0021 to U+007E. The fullwidth space U+3000 is mapped to the regular
// space U+0020. All other characters pass through unchanged.
//
// The U+FF01..U+FF5E range covers fullwidth ASCII variants: ! through ~,
// including fullwidth digits 0-9 (U+FF10..U+FF19) and fullwidth letters A-Z
// (U+FF21..U+FF3A) and a-z (U+FF41..U+FF5A).
//
// Example:
//
//	'ＡＢＣ１２３' -> 'ABC123'
//	'　' -> ' ' (U+3000 -> U+0020)
func IgnoreWidth(c rune) rune {
	// Fullwidth space (U+3000 IDEOGRAPHIC SPACE) -> ASCII space
	if c == 0x3000 {
		return 0x0020
	}
	// Fullwidth ASCII variants (U+FF01..U+FF5E) -> halfwidth (U+0021..U+007E)
	// This range includes: ! (FF01) through ~ (FF5E), digits 0-9 (FF10..FF19),
	// uppercase A-Z (FF21..FF3A), and lowercase a-z (FF41..FF5A).
	if c >= 0xFF01 && c <= 0xFF5E {
		return c - (0xFF01 - 0x0021)
	}
	return c
}
