// SPDX-License-Identifier: Unlicense OR MIT

package paint

import (
	"encoding/binary"
	"image"
	"image/color"

	"github.com/dyike/keel/gio/internal/ops"
	"github.com/dyike/keel/gio/op"
	"github.com/dyike/keel/gio/op/clip"
)

// ImageQuad places an image at Offset and paints only its Rect, given in the
// image's own pixel coordinates; a glyph cut from an atlas page, say.
type ImageQuad struct {
	Offset image.Point
	Rect   image.Rectangle
}

// PaintQuads paints i once per quad. It draws the same pixels as
//
//	for _, q := range quads {
//		t := op.Offset(q.Offset).Push(o)
//		c := clip.Rect(q.Rect).Push(o)
//		i.Add(o)
//		PaintOp{}.Add(o)
//		c.Pop()
//		t.Pop()
//	}
//
// and leaves i as the current brush, but costs one operation instead of
// seven per quad.
//
// Keel patch: text drawn from a glyph atlas spent most of its frame encoding
// and decoding those operations, twice (for input and for drawing).
func (i ImageOp) PaintQuads(o *op.Ops, quads []ImageQuad) {
	i.paintQuads(o, color.NRGBA{R: 255, G: 255, B: 255, A: 255}, quads)
}

func (i ImageOp) paintQuads(o *op.Ops, tint color.NRGBA, quads []ImageQuad) {
	if len(quads) == 0 {
		return
	}
	if size := i.Size(); i.uniform || size.X <= 0 || size.Y <= 0 {
		for _, q := range quads {
			t := op.Offset(q.Offset).Push(o)
			c := clip.Rect(q.Rect).Push(o)
			i.addTinted(o, tint)
			PaintOp{}.Add(o)
			c.Pop()
			t.Pop()
		}
		return
	}
	// The brush, then the quads in an Aux operation; like a path's, it must
	// be the only content of a macro.
	i.write(o, ops.TypeImageQuads, tint)
	m := op.Record(o)
	ops.BeginMulti(&o.Internal)
	aux := ops.WriteMulti(&o.Internal, ops.TypeAuxLen+len(quads)*ops.ImageQuadLen)
	aux[0] = byte(ops.TypeAux)
	bo := binary.LittleEndian
	b := aux[ops.TypeAuxLen:]
	for _, q := range quads {
		bo.PutUint32(b[0:], uint32(int32(q.Offset.X)))
		bo.PutUint32(b[4:], uint32(int32(q.Offset.Y)))
		bo.PutUint32(b[8:], uint32(int32(q.Rect.Min.X)))
		bo.PutUint32(b[12:], uint32(int32(q.Rect.Min.Y)))
		bo.PutUint32(b[16:], uint32(int32(q.Rect.Max.X)))
		bo.PutUint32(b[20:], uint32(int32(q.Rect.Max.Y)))
		b = b[ops.ImageQuadLen:]
	}
	ops.EndMulti(&o.Internal)
	m.Stop().Add(o)
}
