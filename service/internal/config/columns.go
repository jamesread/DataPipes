package config

import "sort"

// ExtractColumnMap maps a source column name to a zero-based CSV column index.
type ExtractColumnMap map[string]int

func SortedExtractColumnNames(m ExtractColumnMap) []string {
	if len(m) == 0 {
		return nil
	}
	type pair struct {
		name string
		idx  int
	}
	pairs := make([]pair, 0, len(m))
	for name, idx := range m {
		pairs = append(pairs, pair{name: name, idx: idx})
	}
	sort.Slice(pairs, func(i, j int) bool {
		if pairs[i].idx == pairs[j].idx {
			return pairs[i].name < pairs[j].name
		}
		return pairs[i].idx < pairs[j].idx
	})
	out := make([]string, len(pairs))
	for i, p := range pairs {
		out[i] = p.name
	}
	return out
}
