package ac

// ExpandEntriesWithPinyin expands high-severity entries with pinyin homophone variants.
// Only entries with Severity == 1 (high risk) are expanded.
// pinyinExpand is a function that generates pinyin homophone variants for a word.
// maxVariants limits the number of expansions per word (default 20 if <= 0).
func ExpandEntriesWithPinyin(entries []WordEntry, pinyinExpand func(word string, max int) []string, maxVariants int) []WordEntry {
	if pinyinExpand == nil {
		return entries
	}
	if maxVariants <= 0 {
		maxVariants = 20
	}

	expanded := make([]WordEntry, 0, len(entries)+(len(entries)/2)*maxVariants)

	for _, entry := range entries {
		// Always include the original entry.
		expanded = append(expanded, entry)

		// Only expand high-severity entries.
		if entry.Severity != 1 {
			continue
		}

		variants := pinyinExpand(entry.Word, maxVariants)
		for _, variant := range variants {
			if variant == "" || variant == entry.Word {
				continue
			}
			expanded = append(expanded, WordEntry{
				Word:     variant,
				Category: entry.Category,
				Severity: entry.Severity,
			})
		}
	}

	return expanded
}
