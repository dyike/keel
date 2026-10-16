// SPDX-License-Identifier: Unlicense OR MIT
package text

import (
	"github.com/dyike/keel/gio/font"
	"github.com/dyike/keel/gio/io/system"
	"github.com/dyike/keel/gio/op"
	"github.com/dyike/keel/gio/op/clip"
	"github.com/dyike/keel/typesetting/shaping"
	"golang.org/x/image/math/fixed"
)

type nativeFontShaper interface {
	setQuery(font.Font, []string)
	orderLine(shaping.Line, bool)
	finishShape()
	shape(*shaperImpl, fixed.Int26_6, system.Locale, []rune) []shaping.Output
	glyphFont(GlyphID) (uintptr, fixed.Int26_6, uint16, bool)
	glyphFile(GlyphID) (string, int, int, fixed.Int26_6, uint16, bool)
	outline(*op.Ops, []Glyph) clip.PathSpec
	bitmaps(*op.Ops, []Glyph) op.CallOp
}
