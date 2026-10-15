// SPDX-License-Identifier: Unlicense OR MIT

package gpu

import (
	"bytes"
	"hash/maphash"
	"unsafe"

	"github.com/dyike/keel/gio/internal/stroke"
)

const (
	maxStrokeEntries = 128
	maxStrokeBytes   = 1 << 20
)

type strokeKey struct {
	content uint64
	width   float32
}

type strokeEntry struct {
	key   strokeKey
	path  []byte
	quads stroke.StrokeQuads
	stamp uint64
	keep  bool
}

// Cache the expensive path-space outline, not raster-space vertices. Rotation
// changes the latter but not the former. Entries own their bytes and quads:
// recorded Ops and stroking scratch are reused independently on later frames.
type strokeCache struct {
	index   map[strokeKey]int
	entries []strokeEntry
	bytes   int
	clock   uint64
	frameID uint64
	// Admission retains hashes only, never frame-owned path or scratch data.
	// Collisions can delay admission, but get still compares complete paths.
	candidates [maxStrokeEntries]strokeCandidate
}

type strokeCandidate struct {
	key   strokeKey
	frame uint64
	valid bool
}

func (s *strokeCache) admit(path []byte, width float32) bool {
	k := strokeKey{maphash.Bytes(pathSeed, path), width}
	c := &s.candidates[k.content%uint64(len(s.candidates))]
	if c.valid && c.key == k && c.frame+1 == s.frameID {
		return true
	}
	*c = strokeCandidate{key: k, frame: s.frameID, valid: true}
	return false
}

func (s *strokeCache) get(path []byte, width float32) (stroke.StrokeQuads, bool) {
	k := strokeKey{maphash.Bytes(pathSeed, path), width}
	i, ok := s.index[k]
	if !ok {
		return nil, false
	}
	e := &s.entries[i]
	if !bytes.Equal(e.path, path) {
		return nil, false // Hash collisions cannot substitute a different path.
	}
	s.clock++
	e.stamp, e.keep = s.clock, true
	return e.quads, true
}

func (s *strokeCache) put(path []byte, width float32, quads stroke.StrokeQuads) {
	cost := len(path) + len(quads)*int(unsafe.Sizeof(stroke.StrokeQuad{}))
	if cost > maxStrokeBytes {
		return // One exceptional path must not turn into unbounded retention.
	}
	k := strokeKey{maphash.Bytes(pathSeed, path), width}
	if i, ok := s.index[k]; ok {
		s.remove(i)
	}
	i := -1
	for j := range s.entries {
		if s.entries[j].path == nil {
			i = j
			break
		}
	}
	for s.bytes+cost > maxStrokeBytes || i < 0 && len(s.entries) == maxStrokeEntries {
		oldest := -1
		for j := range s.entries {
			if s.entries[j].path != nil && (oldest < 0 || s.entries[j].stamp < s.entries[oldest].stamp) {
				oldest = j
			}
		}
		s.remove(oldest)
		i = oldest
	}
	if i < 0 {
		i = len(s.entries)
		s.entries = append(s.entries, strokeEntry{})
	}
	if s.index == nil {
		s.index = make(map[strokeKey]int)
	}
	s.clock++
	e := strokeEntry{key: k, stamp: s.clock, keep: true}
	e.path = make([]byte, len(path))
	copy(e.path, path)
	e.quads = make(stroke.StrokeQuads, len(quads))
	copy(e.quads, quads)
	s.entries[i] = e
	s.index[k] = i
	s.bytes += e.size()
}

func (e strokeEntry) size() int {
	return len(e.path) + len(e.quads)*int(unsafe.Sizeof(stroke.StrokeQuad{}))
}

func (s *strokeCache) remove(i int) {
	e := s.entries[i]
	delete(s.index, e.key)
	s.bytes -= e.size()
	s.entries[i] = strokeEntry{}
}

func (s *strokeCache) frame() {
	s.frameID++
	for i := range s.entries {
		e := &s.entries[i]
		if e.path == nil {
			continue
		}
		if !e.keep {
			s.remove(i)
		} else {
			e.keep = false
		}
	}
}
