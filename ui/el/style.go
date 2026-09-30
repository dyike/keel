package el

import (
	"image/color"

	"gioui.org/io/pointer"
	"gioui.org/unit"
)

// Length is a size along one axis: automatic, a fixed number of dp, or a
// fraction of the parent's content box.
type Length struct {
	kind  lengthKind
	value float32
}

type lengthKind uint8

const (
	autoLen lengthKind = iota
	dpLen
	fracLen
)

// Auto sizes an element to its content, or stretches it where the parent aligns
// items with Stretch.
var Auto = Length{}

// Dp is a fixed length.
func Dp(v float32) Length { return Length{dpLen, v} }

// Frac is a fraction of the parent's content box: Frac(1) is the full width.
func Frac(f float32) Length { return Length{fracLen, f} }

// Full is Frac(1).
var Full = Frac(1)

// px resolves l against a parent size in pixels; -1 means automatic.
func (l Length) px(m unit.Metric, parent int) int {
	switch l.kind {
	case dpLen:
		return m.Dp(unit.Dp(l.value))
	case fracLen:
		if parent < 0 || parent >= inf {
			return -1
		}
		return int(float32(parent) * l.value)
	}
	return -1
}

// Edges are per-side lengths in dp.
type Edges struct{ Top, Right, Bottom, Left float32 }

// Align positions children along an axis.
type Align uint8

const (
	Start Align = iota
	Center
	End
	Stretch      // cross axis only: fill the container
	SpaceBetween // main axis only
	SpaceAround  // main axis only
)

// Style is everything an element can look like. Build it with the methods
// shared by all elements (see Styled); Hover and Active take a func that
// changes a Style, whose methods mirror the element ones.
type Style struct {
	row                 bool // lay children out left to right; default top to bottom
	w, h, minW, minH    Length
	maxW, maxH          Length
	pad, margin         Edges
	gap                 float32
	grow, shrink        float32
	justify, align      Align
	alignSet            bool
	scrollY             bool
	stickBottom         bool
	endVersion          int // ScrollToEndOn; jumps to the end when it changes
	absolute            bool
	top, right, bottom  *float32
	left                *float32
	bg                  *color.NRGBA
	borderWidth, radius float32
	borderColor         color.NRGBA
	cursor              pointer.Cursor
	hidden              bool
	text                textStyle
}

// textStyle is inherited by descendants unless they set their own.
type textStyle struct {
	color *color.NRGBA
	size  unit.Sp
	bold  *bool
	lines int
}

func (t textStyle) inherit(parent textStyle) textStyle {
	if t.color == nil {
		t.color = parent.color
	}
	if t.size == 0 {
		t.size = parent.size
	}
	if t.bold == nil {
		t.bold = parent.bold
	}
	if t.lines == 0 {
		t.lines = parent.lines
	}
	return t
}

// Visual changes, usable in Hover and Active. They return the Style so calls chain.
func (s *Style) Bg(c color.NRGBA) *Style          { s.bg = &c; return s }
func (s *Style) BorderColor(c color.NRGBA) *Style { s.borderColor = c; return s }
func (s *Style) TextColor(c color.NRGBA) *Style   { s.text.color = &c; return s }

func rgb(c uint32) color.NRGBA {
	return color.NRGBA{R: uint8(c >> 16), G: uint8(c >> 8), B: uint8(c), A: 0xff}
}
