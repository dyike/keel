package opentype

import "bytes"

// Shared is a [Resource] over bytes that stay unchanged while the fonts
// parsed from it are in use: a memory-mapped file, or a font compiled into
// the program. Tables of a Shared resource are slices of its bytes instead of
// copies, so a mapped font's tables stay clean file-backed pages the system
// can drop and read back, and parsing allocates far less.
//
// Keel patch: upstream v0.3.5 copies every table it reads.
type Shared struct {
	*bytes.Reader
	data []byte
}

// NewShared returns a Shared resource over data, which must not change
// afterwards.
func NewShared(data []byte) Shared { return Shared{bytes.NewReader(data), data} }

// Bytes returns the whole resource.
func (s Shared) Bytes() []byte { return s.data }
