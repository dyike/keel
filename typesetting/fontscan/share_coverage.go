package fontscan

import (
	"encoding/binary"
	"errors"
	"hash/maphash"
)

// A decoder owns this table for one index read. Never retain record bytes:
// deserializeIndex reuses their storage for the next file. Only decoded,
// immutable coverage is shared; face/instance metadata remains independent.
type runeSetDecoder struct {
	seed       maphash.Seed
	sets       map[uint64]RuneSet
	collisions map[uint64][]RuneSet
}

func (d *runeSetDecoder) decode(dst *RuneSet, data []byte) (int, error) {
	if d == nil {
		return dst.deserializeFrom(data)
	}
	if len(data) < 2 {
		return 0, errors.New("invalid rune set (EOF)")
	}
	pages := int(binary.BigEndian.Uint16(data))
	n := 2 + runePageSize*pages
	if len(data) < n {
		return 0, errors.New("invalid rune set size (EOF)")
	}
	if pages == 0 {
		*dst = RuneSet{}
		return n, nil
	}
	key := maphash.Bytes(d.seed, data[:n])
	prior, found := d.sets[key]
	if found && matchesEncodedRunes(prior, data[:n]) {
		*dst = prior
		return n, nil
	}
	for _, prior := range d.collisions[key] {
		if matchesEncodedRunes(prior, data[:n]) {
			*dst = prior
			return n, nil
		}
	}
	read, err := dst.deserializeFrom(data[:n])
	if err != nil {
		return 0, err
	}
	if !found {
		d.sets[key] = *dst
	} else {
		if d.collisions == nil {
			d.collisions = make(map[uint64][]RuneSet)
		}
		d.collisions[key] = append(d.collisions[key], *dst)
	}
	return read, nil
}

// Hashes only select candidates. Compare every word before sharing, including
// the page references, so a collision can never substitute another coverage.
func matchesEncodedRunes(prior RuneSet, data []byte) bool {
	if len(prior) != int(binary.BigEndian.Uint16(data)) {
		return false
	}
	for i, page := range prior {
		encoded := data[2+runePageSize*i:][:runePageSize]
		// Compare bitmap pairs directly, without a bounds-checked decoder
		// loop or an owned copy of the encoded record on every cache miss.
		if page.ref != binary.BigEndian.Uint16(encoded) ||
			uint64(page.set[0])<<32|uint64(page.set[1]) != binary.BigEndian.Uint64(encoded[2:10]) ||
			uint64(page.set[2])<<32|uint64(page.set[3]) != binary.BigEndian.Uint64(encoded[10:18]) ||
			uint64(page.set[4])<<32|uint64(page.set[5]) != binary.BigEndian.Uint64(encoded[18:26]) ||
			uint64(page.set[6])<<32|uint64(page.set[7]) != binary.BigEndian.Uint64(encoded[26:34]) {
			return false
		}
	}
	return true
}

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
