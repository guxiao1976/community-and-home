package normalize

// numStyleMap maps various Unicode digit representations to standard ASCII digits '0'-'9'.
var numStyleMap = map[rune]rune{
	// Fullwidth digits U+FF10..U+FF19
	'０': '0', '１': '1', '２': '2', '３': '3', '４': '4',
	'５': '5', '６': '6', '７': '7', '８': '8', '９': '9',

	// Circled digit zero U+24EA
	'⓪': '0',

	// Circled digits ①..⑨ U+2460..U+2468
	'①': '1', '②': '2', '③': '3', '④': '4', '⑤': '5',
	'⑥': '6', '⑦': '7', '⑧': '8', '⑨': '9',

	// Circled number ten U+2469
	'⑩': '0', // multi-char mappings not supported per-rune; map to 0 as fallback

	// Parenthesized digits ⑴..⑼ U+2474..U+247C
	'⑴': '1', '⑵': '2', '⑶': '3', '⑷': '4', '⑸': '5',
	'⑹': '6', '⑺': '7', '⑻': '8', '⑼': '9',

	// Dotted digits ⒈..⒐ U+2488..U+2490
	'⒈': '1', '⒉': '2', '⒊': '3', '⒋': '4', '⒌': '5',
	'⒍': '6', '⒎': '7', '⒏': '8', '⒐': '9',

	// Negative circled digits ❶..❿ U+2776..U+277F
	'❶': '1', '❷': '2', '❸': '3', '❹': '4', '❺': '5',
	'❻': '6', '❼': '7', '❽': '8', '❾': '9', '❿': '0',

	// Dingbat negative circled digits ➀..➉ U+2780..U+2789
	'➀': '1', '➁': '2', '➂': '3', '➃': '4', '➄': '5',
	'➅': '6', '➆': '7', '➇': '8', '➈': '9', '➉': '0',

	// Typographic circled digits ➊..➓ U+278A..U+2793
	'➊': '1', '➋': '2', '➌': '3', '➍': '4', '➎': '5',
	'➏': '6', '➐': '7', '➑': '8', '➒': '9', '➓': '0',

	// CJK enclosed digits ㊀..㊉ U+3280..U+3289
	'㊀': '1', '㊁': '2', '㊂': '3', '㊃': '4', '㊄': '5',
	'㊅': '6', '㊆': '7', '㊇': '8', '㊈': '9', '㊉': '0',

	// CJK enclosed circled ideograph twenty..ninety ㊊..㊐ U+328A..U+3290
	'㊊': '0', '㊋': '0', '㊌': '0', '㊍': '0', '㊎': '0',
	'㊏': '0', '㊐': '0',

	// CJK enclosed numbers thirty-one..forty-nine ㉑..㊿ U+3251..U+325F
	'㉑': '1', '㉒': '2', '㉓': '3', '㉔': '4', '㉕': '5',
	'㉖': '6', '㉗': '7', '㉘': '8', '㉙': '9',

	// Superscript digits
	// ⁰ U+2070
	'⁰': '0',
	// ¹²³ U+00B9, U+00B2, U+00B3
	'¹': '1', '²': '2', '³': '3',
	// ⁴⁵⁶⁷⁸⁹ U+2074..U+2079
	'⁴': '4', '⁵': '5', '⁶': '6', '⁷': '7', '⁸': '8', '⁹': '9',

	// Subscript digits ₀₁₂₃₄₅₆₇₈₉ U+2080..U+2089
	'₀': '0', '₁': '1', '₂': '2', '₃': '3', '₄': '4',
	'₅': '5', '₆': '6', '₇': '7', '₈': '8', '₉': '9',

	// Roman numeral uppercase Ⅰ..Ⅻ U+2160..U+216B
	'Ⅰ': '1', 'Ⅱ': '2', 'Ⅲ': '3', 'Ⅳ': '4', 'Ⅴ': '5',
	'Ⅵ': '6', 'Ⅶ': '7', 'Ⅷ': '8', 'Ⅸ': '9', 'Ⅹ': '0',
	'Ⅺ': '0', 'Ⅻ': '0',

	// Roman numeral lowercase ⅰ..ⅻ U+2170..U+217B
	'ⅰ': '1', 'ⅱ': '2', 'ⅲ': '3', 'ⅳ': '4', 'ⅴ': '5',
	'ⅵ': '6', 'ⅶ': '7', 'ⅷ': '8', 'ⅸ': '9', 'ⅹ': '0',
	'ⅺ': '0', 'ⅻ': '0',

	// Roman numeral combining ⅬⅭⅮⅯ U+216C..U+216F (50, 100, 500, 1000)
	'Ⅼ': '0', 'Ⅽ': '0', 'Ⅾ': '0', 'Ⅿ': '0',

	// Roman numeral lowercase combining ⅼⅽⅾⅿ U+217C..U+217F
	'ⅼ': '0', 'ⅽ': '0', 'ⅾ': '0', 'ⅿ': '0',

	// Chinese number digits (lowercase) 一..九
	'〇': '0', '零': '0',
	'一': '1', '二': '2', '三': '3', '四': '4', '五': '5',
	'六': '6', '七': '7', '八': '8', '九': '9',

	// Formal Chinese numerals (banker's anti-fraud) 壹..玖
	'壹': '1', '贰': '2', '叁': '3', '肆': '4', '伍': '5',
	'陆': '6', '柒': '7', '捌': '8', '玖': '9',

	// Hangzhou numerals 〡..〩 U+3021..U+3029
	'〡': '1', '〢': '2', '〣': '3', '〤': '4', '〥': '5',
	'〦': '6', '〧': '7', '〨': '8', '〩': '9',

	// Counting rod numerals 𝍠..𝍩 U+1D360..U+1D369
	'𝍠': '1', '𝍡': '2', '𝍢': '3', '𝍣': '4', '𝍤': '5',
	'𝍥': '6', '𝍦': '7', '𝍧': '8', '𝍨': '9',
}

// IgnoreNumStyle normalizes various Unicode digit representations to standard ASCII digits.
// Characters not in the mapping are returned unchanged.
//
// For multi-character number representations (e.g., "⑩" = 10, "Ⅹ" = 10), the last
// digit is used since per-rune conversion cannot produce multi-character output.
//
// Example:
//
//	'①２³三肆' -> '12334'
//	'ⅧⅨ' -> '89'
//	'❶❷❸' -> '123'
func IgnoreNumStyle(c rune) rune {
	if d, ok := numStyleMap[c]; ok {
		return d
	}
	return c
}
