package kit

import "image"

// Labels keep their association with the node's center. Prefer the hovered
// node, then larger flows; omit a label instead of moving it onto another node.
// Actual laid-out rectangles account for font scaling and custom label heights.
type sankeyLabelPlacement struct {
	bounds   image.Rectangle
	gap      int
	occupied []image.Rectangle
}

func (p *sankeyLabelPlacement) reset(bounds image.Rectangle, gap int) {
	p.bounds = bounds
	p.gap = gap
	p.occupied = p.occupied[:0]
}
func (p *sankeyLabelPlacement) take(rect image.Rectangle) bool {
	if rect.Empty() || !rect.In(p.bounds) {
		return false
	}
	for _, prior := range p.occupied {
		if rect.Overlaps(prior.Inset(-p.gap)) {
			return false
		}
	}
	p.occupied = append(p.occupied, rect)
	return true
}
