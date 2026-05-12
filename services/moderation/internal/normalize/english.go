package normalize

// englishStyleMap maps special Unicode English letter representations to standard ASCII letters.
var englishStyleMap = map[rune]rune{
	// Circled uppercase letters U+24B6..U+24CF (Ⓐ..Ⓩ)
	'Ⓐ': 'A', 'Ⓑ': 'B', 'Ⓒ': 'C', 'Ⓓ': 'D', 'Ⓔ': 'E',
	'Ⓕ': 'F', 'Ⓖ': 'G', 'Ⓗ': 'H', 'Ⓘ': 'I', 'Ⓙ': 'J',
	'Ⓚ': 'K', 'Ⓛ': 'L', 'Ⓜ': 'M', 'Ⓝ': 'N', 'Ⓞ': 'O',
	'Ⓟ': 'P', 'Ⓠ': 'Q', 'Ⓡ': 'R', 'Ⓢ': 'S', 'Ⓣ': 'T',
	'Ⓤ': 'U', 'Ⓥ': 'V', 'Ⓦ': 'W', 'Ⓧ': 'X', 'Ⓨ': 'Y', 'Ⓩ': 'Z',

	// Circled lowercase letters U+24D0..U+24E9 (ⓐ..ⓩ)
	'ⓐ': 'a', 'ⓑ': 'b', 'ⓒ': 'c', 'ⓓ': 'd', 'ⓔ': 'e',
	'ⓕ': 'f', 'ⓖ': 'g', 'ⓗ': 'h', 'ⓘ': 'i', 'ⓙ': 'j',
	'ⓚ': 'k', 'ⓛ': 'l', 'ⓜ': 'm', 'ⓝ': 'n', 'ⓞ': 'o',
	'ⓟ': 'p', 'ⓠ': 'q', 'ⓡ': 'r', 'ⓢ': 's', 'ⓣ': 't',
	'ⓤ': 'u', 'ⓥ': 'v', 'ⓦ': 'w', 'ⓧ': 'x', 'ⓨ': 'y', 'ⓩ': 'z',

	// Small circled uppercase letters U+249C..U+24B5 (⒜..⒵)
	'⒜': 'A', '⒝': 'B', '⒞': 'C', '⒟': 'D', '⒠': 'E',
	'⒡': 'F', '⒢': 'G', '⒣': 'H', '⒤': 'I', '⒥': 'J',
	'⒦': 'K', '⒧': 'L', '⒨': 'M', '⒩': 'N', '⒪': 'O',
	'⒫': 'P', '⒬': 'Q', '⒭': 'R', '⒮': 'S', '⒯': 'T',
	'⒰': 'U', '⒱': 'V', '⒲': 'W', '⒳': 'X', '⒴': 'Y', '⒵': 'Z',

	// Negative circled Latin uppercase U+1F150..U+1F169 (🅐..🅩)
	'🅐': 'A', '🅑': 'B', '🅒': 'C', '🅓': 'D', '🅔': 'E',
	'🅕': 'F', '🅖': 'G', '🅗': 'H', '🅘': 'I', '🅙': 'J',
	'🅚': 'K', '🅛': 'L', '🅜': 'M', '🅝': 'N', '🅞': 'O',
	'🅟': 'P', '🅠': 'Q', '🅡': 'R', '🅢': 'S', '🅣': 'T',
	'🅤': 'U', '🅥': 'V', '🅦': 'W', '🅧': 'X', '🅨': 'Y', '🅩': 'Z',

	// Negative circled Latin lowercase U+1F170..U+1F189 (🅰..🆉)
	'🅰': 'A', '🅱': 'B', '🅲': 'C', '🅳': 'D', '🅴': 'E',
	'🅵': 'F', '🅶': 'G', '🅷': 'H', '🅸': 'I', '🅹': 'J',
	'🅺': 'K', '🅻': 'L', '🅼': 'M', '🅽': 'N', '🅾': 'O',
	'🅿': 'P', '🆀': 'Q', '🆁': 'R', '🆂': 'S', '🆃': 'T',
	'🆄': 'U', '🆅': 'V', '🆆': 'W', '🆇': 'X', '🆈': 'Y', '🆉': 'Z',

	// Squared Latin uppercase U+1F130..U+1F149 (🄰..🅉)
	'🄰': 'A', '🄱': 'B', '🄲': 'C', '🄳': 'D', '🄴': 'E',
	'🄵': 'F', '🄶': 'G', '🄷': 'H', '🄸': 'I', '🄹': 'J',
	'🄺': 'K', '🄻': 'L', '🄼': 'M', '🄽': 'N', '🄾': 'O',
	'🄿': 'P', '🅀': 'Q', '🅁': 'R', '🅂': 'S', '🅃': 'T',
	'🅄': 'U', '🅅': 'V', '🅆': 'W', '🅇': 'X', '🅈': 'Y', '🅉': 'Z',

	// Negative squared Latin uppercase U+1F170..U+1F189 (partial overlap with above)
	// Already covered above, but adding lowercase squared variants:
	'🆎': 'A', // AB squared
	'🆏': 'A', // CL squared
	'🆐': 'A', // CR squared
	'🆑': 'C', // C with circle
	'🆒': 'H', // H with square
	'🆓': 'F', // FREE (map to F for consistency)
	'🆔': 'I', // ID
	'🆕': 'N', // NEW
	'🆖': 'N', // NG
	'🆗': 'O', // OK
	'🆘': 'S', // SOS
	'🆙': 'U', // UP
	'🆚': 'V', // VS

	// Parenthesized Latin uppercase U+1F110..U+1F129
	'🄐': 'A', '🄑': 'B', '🄒': 'C', '🄓': 'D', '🄔': 'E',
	'🄕': 'F', '🄖': 'G', '🄗': 'H', '🄘': 'I', '🄙': 'J',
	'🄚': 'K', '🄛': 'L', '🄜': 'M', '🄝': 'N', '🄞': 'O',
	'🄟': 'P', '🄠': 'Q', '🄡': 'R', '🄢': 'S', '🄣': 'T',
	'🄤': 'U', '🄥': 'V', '🄦': 'W', '🄧': 'X', '🄨': 'Y', '🄩': 'Z',

	// Fullwidth Latin uppercase A-Z U+FF21..U+FF3A
	'Ａ': 'A', 'Ｂ': 'B', 'Ｃ': 'C', 'Ｄ': 'D', 'Ｅ': 'E',
	'Ｆ': 'F', 'Ｇ': 'G', 'Ｈ': 'H', 'Ｉ': 'I', 'Ｊ': 'J',
	'Ｋ': 'K', 'Ｌ': 'L', 'Ｍ': 'M', 'Ｎ': 'N', 'Ｏ': 'O',
	'Ｐ': 'P', 'Ｑ': 'Q', 'Ｒ': 'R', 'Ｓ': 'S', 'Ｔ': 'T',
	'Ｕ': 'U', 'Ｖ': 'V', 'Ｗ': 'W', 'Ｘ': 'X', 'Ｙ': 'Y', 'Ｚ': 'Z',

	// Fullwidth Latin lowercase a-z U+FF41..U+FF5A
	'ａ': 'a', 'ｂ': 'b', 'ｃ': 'c', 'ｄ': 'd', 'ｅ': 'e',
	'ｆ': 'f', 'ｇ': 'g', 'ｈ': 'h', 'ｉ': 'i', 'ｊ': 'j',
	'ｋ': 'k', 'ｌ': 'l', 'ｍ': 'm', 'ｎ': 'n', 'ｏ': 'o',
	'ｐ': 'p', 'ｑ': 'q', 'ｒ': 'r', 'ｓ': 's', 'ｔ': 't',
	'ｕ': 'u', 'ｖ': 'v', 'ｗ': 'w', 'ｘ': 'x', 'ｙ': 'y', 'ｚ': 'z',
}

// IgnoreEnglishStyle normalizes special Unicode English letter representations
// (circled, parenthesized, squared, fullwidth, etc.) to standard ASCII letters.
// Characters not in the mapping are returned unchanged.
//
// Example:
//
//	'ⒶⒷⒸ' -> 'ABC'
//	'ⓐⓑⓒ' -> 'abc'
//	'⒜⒝⒞' -> 'ABC'
func IgnoreEnglishStyle(c rune) rune {
	if s, ok := englishStyleMap[c]; ok {
		return s
	}
	return c
}
