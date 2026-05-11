package normalize

// CompressRepeat compresses consecutive runs of 3+ identical runes into a single rune.
// This is a standalone function that can be used independently of the Normalizer chain.
//
// Example:
//
//	"heeeeello" -> "helo"
//	"abc" -> "abc"
//	"aabbcc" -> "aabbcc" (only 2 repeats, not compressed)
//	"woooow" -> "wow"
func CompressRepeat(text string) string {
	if len(text) == 0 {
		return text
	}
	runes := []rune(text)
	result := make([]rune, 0, len(runes))

	i := 0
	for i < len(runes) {
		j := i + 1
		for j < len(runes) && runes[j] == runes[i] {
			j++
		}
		if j-i >= 3 {
			// Compress: keep only the first character
			result = append(result, runes[i])
		} else {
			// Keep all characters in short runs (0, 1, or 2 repeats)
			for k := i; k < j; k++ {
				result = append(result, runes[k])
			}
		}
		i = j
	}
	return string(result)
}

// CompressRepeatWithPosMap compresses consecutive runs of 3+ identical runes
// into a single rune, and returns a position map from the compressed result
// back to the original rune positions.
func CompressRepeatWithPosMap(text string) (compressed string, posMap []int) {
	runes := []rune(text)
	if len(runes) == 0 {
		return "", nil
	}

	result := make([]rune, 0, len(runes))
	pm := make([]int, 0, len(runes))

	i := 0
	for i < len(runes) {
		j := i + 1
		for j < len(runes) && runes[j] == runes[i] {
			j++
		}
		if j-i >= 3 {
			result = append(result, runes[i])
			pm = append(pm, i)
		} else {
			for k := i; k < j; k++ {
				result = append(result, runes[k])
				pm = append(pm, k)
			}
		}
		i = j
	}

	return string(result), pm
}
