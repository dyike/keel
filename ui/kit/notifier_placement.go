package kit

import "github.com/dyike/keel/ui/el"

// NoticePlacement chooses one of eight window-edge positions.
type NoticePlacement uint8

const (
	NoticeDefault NoticePlacement = iota
	NoticeTopRight
	NoticeTopLeft
	NoticeTopCenter
	NoticeBottomRight
	NoticeBottomLeft
	NoticeBottomCenter
	NoticeLeftCenter
	NoticeRightCenter
)

// Placement sets this host's default for notices without an override. Existing
// notices follow changes. Default restores top right; invalid values are ignored.
func (v *NotifierView) Placement(position NoticePlacement) *NotifierView {
	if position <= NoticeRightCenter {
		v.placement = position
	}
	return v
}
func (v *NotifierView) position(position NoticePlacement) NoticePlacement {
	if position == NoticeDefault || position > NoticeRightCenter {
		position = v.placement
	}
	if position == NoticeDefault {
		position = NoticeTopRight
	}
	return position
}
func noticeAnchor(position NoticePlacement, w, h float32) (x, y float32, side el.Side, align el.Align) {
	x, y, side, align = max(16, w-16), float32(16), el.Bottom, el.End
	switch position {
	case NoticeTopLeft:
		x, align = 16, el.Start
	case NoticeTopCenter:
		x, align = w/2, el.Center
	case NoticeBottomRight:
		y, side = max(16, h-16), el.Top
	case NoticeBottomLeft:
		x, y, side, align = 16, max(16, h-16), el.Top, el.Start
	case NoticeBottomCenter:
		x, y, side, align = w/2, max(16, h-16), el.Top, el.Center
	case NoticeLeftCenter:
		x, y, side, align = 16, h/2, el.Right, el.Center
	case NoticeRightCenter:
		y, side, align = h/2, el.Left, el.Center
	}
	return
}
