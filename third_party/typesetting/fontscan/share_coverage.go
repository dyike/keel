package fontscan

// System font metadata is read-only after publication. Fonts in a family
// commonly have identical coverage across weights and styles; retain one
// copy of each distinct coverage rather than one per face.
func (index systemFontsIndex) shareRuneSets() {
	seen := make(map[uint64][]RuneSet)
	for i := range index {
		for j := range index[i].footprints {
			fp := &index[i].footprints[j]
			if len(fp.Runes) == 0 {
				continue
			}
			hash := uint64(14695981039346656037)
			for _, page := range fp.Runes {
				hash = (hash ^ uint64(page.ref)) * 1099511628211
				for _, word := range page.set {
					hash = (hash ^ uint64(word)) * 1099511628211
				}
			}
			shared := false
			for _, prior := range seen[hash] {
				if len(prior) != len(fp.Runes) {
					continue
				}
				equal := true
				for k, p := range prior {
					if p != fp.Runes[k] {
						equal = false
						break
					}
				}
				if equal {
					fp.Runes = prior
					shared = true
					break
				}
			}
			if !shared {
				seen[hash] = append(seen[hash], fp.Runes)
			}
		}
	}
}
