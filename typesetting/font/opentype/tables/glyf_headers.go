// SPDX-License-Identifier: Unlicense OR BSD-3-Clause

package tables

import (
	"encoding/binary"
	"fmt"
)

// ParseGlyfHeaders validates glyf without retaining decoded contours. It returns
// the same headers as ParseGlyf with Data cleared, and the point counts needed
// by gvar (component counts for composite glyphs). Outlines can subsequently be
// decoded individually with ParseGlyph. No input bytes are modified.
func ParseGlyfHeaders(src []byte, offsets []uint32) (Glyf, []int, error) {
	if len(offsets) == 0 {
		return nil, nil, fmt.Errorf("invalid empty loca")
	}
	headers := make(Glyf, len(offsets)-1)
	counts := make([]int, len(headers))
	var composite CompositeGlyph
	for i := range headers {
		start, end := offsets[i], offsets[i+1]
		if start > end || uint64(end) > uint64(len(src)) {
			return nil, nil, fmt.Errorf("invalid glyph %d offsets", i)
		}
		if start == end {
			continue
		}
		data := src[start:end]
		if len(data) < 10 {
			return nil, nil, fmt.Errorf("invalid glyph %d header", i)
		}
		h := Glyph{
			numberOfContours: int16(binary.BigEndian.Uint16(data)),
			XMin:             int16(binary.BigEndian.Uint16(data[2:])),
			YMin:             int16(binary.BigEndian.Uint16(data[4:])),
			XMax:             int16(binary.BigEndian.Uint16(data[6:])),
			YMax:             int16(binary.BigEndian.Uint16(data[8:])),
		}
		data = data[10:]
		if h.numberOfContours >= 0 {
			count, err := simpleGlyphPointCount(data, int(h.numberOfContours))
			if err != nil {
				return nil, nil, err
			}
			counts[i] = count
		} else {
			composite.Glyphs = composite.Glyphs[:0]
			composite.Instructions = nil
			if err := composite.parseGlyphs(data); err != nil {
				return nil, nil, err
			}
			counts[i] = len(composite.Glyphs)
		}
		headers[i] = h
	}
	return headers, counts, nil
}

// Validate the same flag runs and coordinate lengths as SimpleGlyph.parsePoints,
// without allocating a point array merely to discard it at font initialization.
func simpleGlyphPointCount(src []byte, contours int) (int, error) {
	end := contours * 2
	if len(src) < end+2 {
		return 0, fmt.Errorf("invalid simple glyph contours")
	}
	points := 0
	if contours != 0 {
		points = int(binary.BigEndian.Uint16(src[end-2:])) + 1
	}
	instructions := int(binary.BigEndian.Uint16(src[end:]))
	end += 2 + instructions
	if len(src) < end {
		return 0, fmt.Errorf("invalid simple glyph instructions")
	}
	src = src[end:]
	coordinates := 0
	for i := 0; i < points; {
		if len(src) == 0 {
			return 0, fmt.Errorf("invalid simple glyph flags")
		}
		flag := src[0]
		src = src[1:]
		run := 1
		if flag&0x08 != 0 {
			if len(src) == 0 {
				return 0, fmt.Errorf("invalid simple glyph flag repeat")
			}
			run += int(src[0])
			src = src[1:]
			run = min(run, points-i) // Match parsePoints' permissive final run.
		}
		x, y := 0, 0
		if flag&xShortVector != 0 {
			x = 1
		} else if flag&xIsSameOrPositiveXShortVector == 0 {
			x = 2
		}
		if flag&yShortVector != 0 {
			y = 1
		} else if flag&yIsSameOrPositiveYShortVector == 0 {
			y = 2
		}
		coordinates += run * (x + y)
		i += run
	}
	if len(src) < coordinates {
		return 0, fmt.Errorf("invalid simple glyph coordinates")
	}
	return points, nil
}
