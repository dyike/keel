// SPDX-License-Identifier: Unlicense OR MIT

// Package surface holds images that live in the platform's GPU memory, which
// renderers sample in place. Keel patch.
package surface

import (
	"image"
	"sync"
)

// Surface is a picture its owner replaces from frame to frame: on Apple
// platforms an IOSurface, such as a hardware video decoder's output. The
// owner keeps each surface it sets valid until it has set two later ones,
// by which time no frame still samples it.
type Surface struct {
	mu   sync.Mutex
	ref  uintptr
	size image.Point
	gen  uint64
}

// Set makes ref, of the given size in pixels, the current picture.
func (s *Surface) Set(ref uintptr, size image.Point) {
	s.mu.Lock()
	s.ref, s.size = ref, size
	s.gen++
	s.mu.Unlock()
}

// Current returns the current picture and a number that changes with it.
func (s *Surface) Current() (ref uintptr, size image.Point, gen uint64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.ref, s.size, s.gen
}
