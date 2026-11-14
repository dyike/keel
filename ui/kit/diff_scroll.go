package kit

import (
	"github.com/dyike/keel/gio/op"
	"github.com/dyike/keel/ui/core"
	"math"
)

// diffMappedLine maps unchanged lines exactly; a replacement's interior is
// interpolated. A pure deletion maps to the boundary in the other document.
func diffMappedLine(line float32, changes []diffRange, reverse bool) float32 {
	offset := 0
	for _, c := range changes {
		a, z, b, e := c.oldStart, c.oldEnd, c.newStart, c.newEnd
		if reverse {
			a, z, b, e = b, e, a, z
		}
		if line < float32(a) {
			break
		}
		if line < float32(z) {
			return float32(b) + (line-float32(a))*float32(e-b)/float32(z-a)
		}
		offset = e - z
	}
	return max(0, line+float32(offset))
}

func (v *DiffView) syncSplitScroll(gtx core.C, draw func()) {
	draw()
	now := [2]float32{v.old.scrollY, v.new.scrollY}
	if !v.scrollKnown {
		v.lastScroll = now
		v.scrollKnown = true
		return
	}
	oldMoved, newMoved := now[0] != v.lastScroll[0], now[1] != v.lastScroll[1]
	if !oldMoved && !newMoved {
		return
	}
	source, target := v.old, v.new
	reverse := newMoved && (!oldMoved || v.new.focused)
	if reverse {
		source, target = v.new, v.old
	}
	if source.metrics.lh > 0 && target.metrics.lh > 0 {
		before := target.scrollY
		target.scrollY = diffScrollPosition(source, target, v.changes, reverse)
		target.clampScroll()
		if target.scrollY != before {
			gtx.Execute(op.InvalidateCmd{})
		}
	}
	// Include the clamped mirrored value in the checkpoint, avoiding feedback
	// redraws when one side reaches its end before the other.
	v.lastScroll = [2]float32{v.old.scrollY, v.new.scrollY}
}

// Control rows match by region identity, including pure insertion/deletion
// boundaries where ordinary source-line interpolation maps past the region.
func diffScrollPosition(source, target *CodeEditorView, changes []diffRange, reverse bool) float32 {
	source.ensureBlocks()
	target.ensureBlocks()
	for i, b := range source.blocks.rows {
		top := float32(b.row*source.metrics.lh + i*source.blocks.height)
		if top > source.scrollY {
			break
		}
		if source.scrollY >= top+float32(source.blocks.height) {
			continue
		}
		for j, other := range target.blocks.rows {
			if other.index == b.index {
				return float32(other.row*target.metrics.lh+j*target.blocks.height) + (source.scrollY-top)*float32(target.blocks.height)/float32(max(1, source.blocks.height))
			}
		}
		break
	}
	whole := source.rowAt(source.scrollY)
	offset := source.scrollY - float32(source.rowTop(whole))
	fraction := max(float32(0), offset/float32(source.metrics.lh))
	line := float32(source.vrow(whole).line) + fraction
	mapped := diffMappedLine(line, changes, reverse)
	targetLine := int(math.Floor(float64(mapped)))
	result := float32(target.rowTop(target.vrowOf(codePos{targetLine, 0}))) + (mapped-float32(targetLine))*float32(target.metrics.lh)
	if offset < 0 {
		result += offset
	}
	return result
}
