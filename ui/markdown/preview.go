package markdown

import (
	"image"

	"github.com/dyike/keel/third_party/gio/op"
	"github.com/dyike/keel/third_party/gio/op/clip"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
)

type documentPreview struct {
	lines   int
	clamped bool
	origin  image.Point
	height  int
}

// MaxLines limits the document to n body-line heights, including block spacing.
// Zero or a negative value removes the limit. Taller headings can consume more
// than one line of this budget. It does not add an ellipsis or expand button.
func (d *Doc) MaxLines(n int) *Doc {
	d.preview.lines = max(0, min(n, 100000))
	if n <= 0 {
		d.preview.clamped = false
	}
	if n > 0 {
		d.ranges.reveal = nil
	}
	return d
}

// IsClamped reports whether the most recently painted frame hid content because
// of MaxLines. It is false before the first frame and after removing the limit.
func (d *Doc) IsClamped() bool { return d.preview.clamped }

func (d *Doc) beginPreview(cx *el.Context, root el.Element, gtx core.C) {
	d.preview.origin, _ = cx.PaintGeometry()
	d.preview.height = gtx.Constraints.Max.Y
	clamped := d.preview.lines > 0 && el.ContainerContentSize(root).Y > d.preview.height
	if clamped != d.preview.clamped {
		d.preview.clamped = clamped
		// Let a caller's expand affordance observe the result on its next Render.
		gtx.Execute(op.InvalidateCmd{})
	}
	if d.preview.lines > 0 {
		d.ranges.reveal = nil
	}
}

func (d *Doc) paintPreview(gtx core.C, draw func()) {
	if d.preview.lines <= 0 {
		draw()
		return
	}
	bottom := d.preview.height
	// Reserve only whole ordinary text lines. Images and a line taller than the
	// entire budget retain their visible portion, rather than blanking the view.
	for _, part := range d.selection.parts {
		for _, p := range part.r.rt.pieces {
			if p.run >= len(part.r.rt.runs) {
				continue
			}
			rn := part.r.rt.runs[p.run]
			if rn.image != nil {
				continue
			}
			rect := p.rect.Add(part.rect.Min)
			if rect.Min.Y < bottom && rect.Max.Y > bottom && rect.Dy() <= d.preview.height {
				bottom = rect.Min.Y
			}
		}
	}
	area := clip.Rect(image.Rect(0, 0, gtx.Constraints.Max.X, max(0, bottom))).Push(gtx.Ops)
	draw()
	area.Pop()
}

// Suppress a table row's leading padding/border when none of its first line can
// fit. Long rows still retain the content that fits. The document clip above
// removes a partial last line; this guard prevents a stranded empty row frame.
func (d *Doc) guardPreview(cx *el.Context, row *el.DivEl) {
	row.Decorate(func(gtx core.C, draw func()) {
		if d.preview.lines <= 0 {
			draw()
			return
		}
		origin, _ := cx.PaintGeometry()
		top := origin.Y - d.preview.origin.Y
		first := -1
		el.VisitWidgets(row, gtx.Metric, func(w core.Widget, rect image.Rectangle) {
			r, ok := w.(*richBlock)
			if !ok || len(r.rt.pieces) == 0 {
				return
			}
			height := rect.Min.Y + r.rt.pieces[0].rect.Max.Y
			if first < 0 || height < first {
				first = height
			}
		})
		if first > 0 && first <= d.preview.height && top+first > d.preview.height {
			return
		}
		draw()
	})
}
