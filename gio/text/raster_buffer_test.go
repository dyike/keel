// SPDX-License-Identifier: Unlicense OR MIT

package text

import (
	"image/color"
	"testing"

	"github.com/dyike/keel/gio/op"
)

func TestGlyphBufferLeaseAndHookChanges(t *testing.T) {
	var sh Shaper
	a, token := sh.BorrowGlyphBuffer()
	b, nested := sh.BorrowGlyphBuffer()
	if &a[:cap(a)][0] == &b[:cap(b)][0] {
		t.Fatal("overlapping borrowers alias")
	}
	sh.ReleaseGlyphBuffer(b, nested)
	sh.ReleaseGlyphBuffer(a, token)
	if got := testing.AllocsPerRun(20, func() { g, token := sh.BorrowGlyphBuffer(); sh.ReleaseGlyphBuffer(g, token) }); got != 0 {
		t.Fatal("warm vector buffer allocated")
	}
	old, token := sh.BorrowGlyphBuffer()
	sh.SetRasterHook(func(*op.Ops, []Glyph, color.NRGBA, float32) bool { return true })
	sh.ReleaseGlyphBuffer(old, token)
	if sh.glyphBuffer != nil {
		t.Fatal("old lease crossed a hook change")
	}
	x, token := sh.BorrowGlyphBuffer()
	if token != 0 {
		t.Fatal("legacy hook opted into borrowing implicitly")
	}
	sh.ReleaseGlyphBuffer(x, token)
	if sh.glyphBuffer != nil {
		t.Fatal("legacy hook's buffer was reused")
	}
	sh.SetBorrowedRasterHook(func(*op.Ops, []Glyph, color.NRGBA, float32) bool { return true })
	if got := testing.AllocsPerRun(20, func() { g, token := sh.BorrowGlyphBuffer(); sh.ReleaseGlyphBuffer(g, token) }); got != 0 {
		t.Fatal("warm borrowed-hook buffer allocated")
	}
}
