// SPDX-License-Identifier: Unlicense OR MIT

package stroke

// Scratch reuses CPU stroking buffers. It is not safe for concurrent use.
// Commands returns borrowed quads valid until the next Commands or Trim call;
// callers retaining results (such as a path cache) must copy them.
type Scratch struct {
	decoded, output, rhs, lhs, curve StrokeQuads
	cubic                            []QuadSegment
	states                           []strokeState
}

func (s *Scratch) Commands(style StrokeStyle, data []byte) StrokeQuads {
	s.decoded, s.cubic = decodeStrokeQuads(data, s.decoded[:0], s.cubic[:0])
	s.output = s.output[:0]
	for start := 0; start < len(s.decoded); {
		end := start + 1
		for end < len(s.decoded) && s.decoded[end].Contour == s.decoded[start].Contour {
			end++
		}
		path := s.decoded[start:end]
		rhs, lhs := path.offsetReuse(.5*style.Width, style, s)
		if lhs == nil {
			s.output = appendStrokeCopy(s.output, rhs)
		} else {
			var area float32
			for i := 1; i < len(path); i++ {
				a, b := path[i].Quad.To, path[i-1].Quad.To
				area += (a.X - b.X) * (a.Y + b.Y)
			}
			if area <= 0 {
				reverseStroke(lhs)
				s.output = appendStrokeCopy(s.output, rhs)
				s.output = appendStrokeCopy(s.output, lhs)
			} else {
				reverseStroke(rhs)
				s.output = appendStrokeCopy(s.output, lhs)
				s.output = appendStrokeCopy(s.output, rhs)
			}
		}
		start = end
	}
	return s.output
}

// Trim drops unusually large temporary arrays after the borrowed output has
// been consumed. Ordinary animation buffers remain available for the next miss.
func (s *Scratch) Trim() {
	const maxQuads = 16384
	for _, q := range []*StrokeQuads{&s.decoded, &s.output, &s.rhs, &s.lhs, &s.curve} {
		if cap(*q) > maxQuads {
			*q = nil
		}
	}
	if cap(s.cubic) > maxQuads {
		s.cubic = nil
	}
	if cap(s.states) > maxQuads {
		s.states = nil
	}
}

func reverseStroke(q StrokeQuads) {
	for i, j := 0, len(q)-1; i <= j; i, j = i+1, j-1 {
		a, b := q[i], q[j]
		a.Quad.From, a.Quad.To = a.Quad.To, a.Quad.From
		b.Quad.From, b.Quad.To = b.Quad.To, b.Quad.From
		q[i], q[j] = b, a
	}
}

// Unlike StrokeQuads.append, this never adopts the source's backing array.
// Each scratch slice must remain independently writable on the next call.
func appendStrokeCopy(dst, src StrokeQuads) StrokeQuads {
	if len(dst) == 0 {
		return append(dst, src...)
	}
	return dst.append(src)
}
