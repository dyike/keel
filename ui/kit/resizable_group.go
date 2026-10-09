package kit

import (
	"math"
	"strconv"

	"github.com/dyike/keel/third_party/gio/io/key"
	"github.com/dyike/keel/third_party/gio/io/pointer"
	"github.com/dyike/keel/third_party/gio/op"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/locale"
)

// ResizablePanel identifies a pane. Size is its initial dp size (zero shares
// available space); Min defaults to zero and Max zero means unbounded.
type ResizablePanel struct {
	ID             string
	Content        el.View
	Size, Min, Max float32
	Hidden         bool
}

// ResizableGroupView lays out independently constrained panes on one axis.
type ResizableGroupView struct {
	auto                         map[string]bool
	panels                       []ResizablePanel
	sizes                        map[string]float32
	handles                      map[string]*ResizableView
	vertical, disabled, measured bool
	total                        float32
	onChange                     func(map[string]float32)
	appearance                   func(ResizableHandleAppearance) ResizableHandleAppearance
}

func ResizableGroup(panels ...ResizablePanel) *ResizableGroupView {
	v := &ResizableGroupView{}
	v.SetPanels(panels...)
	return v
}

// SetPanels preserves sizes for retained IDs; new IDs use Size. Empty and
// duplicate IDs are ignored. Hidden panes retain their size and content state.
func (v *ResizableGroupView) SetPanels(panels ...ResizablePanel) {
	sizes := make(map[string]float32)
	autos := make(map[string]bool)
	out := make([]ResizablePanel, 0, len(panels))
	for _, p := range panels {
		if p.ID == "" {
			continue
		}
		if _, ok := sizes[p.ID]; ok {
			continue
		}
		if !finiteNumber(float64(p.Min)) || p.Min < 0 {
			p.Min = 0
		}
		if !finiteNumber(float64(p.Max)) || p.Max < 0 {
			p.Max = 0
		}
		if !finiteNumber(float64(p.Size)) || p.Size < 0 {
			p.Size = 0
		}
		size, ok := v.sizes[p.ID]
		if !ok {
			size = p.Size
			autos[p.ID] = p.Size == 0
		}
		if ok {
			autos[p.ID] = v.auto[p.ID]
		}
		sizes[p.ID] = size
		out = append(out, p)
	}
	v.panels, v.sizes, v.auto = out, sizes, autos
	v.handles = make(map[string]*ResizableView)
}
func (v *ResizableGroupView) Vertical() *ResizableGroupView { v.vertical = true; return v }
func (v *ResizableGroupView) SetDisabled(on bool)           { v.disabled = on }
func (v *ResizableGroupView) HandleAppearance(fn func(ResizableHandleAppearance) ResizableHandleAppearance) *ResizableGroupView {
	v.appearance = fn
	return v
}
func (v *ResizableGroupView) OnChange(fn func(map[string]float32)) *ResizableGroupView {
	v.onChange = fn
	return v
}
func (v *ResizableGroupView) Sizes() map[string]float32 {
	out := make(map[string]float32, len(v.sizes))
	for id, size := range v.sizes {
		out[id] = size
	}
	return out
}

// SetSizes applies finite nonnegative sizes without a callback. Layout fits
// them to the available axis length and panel constraints on the next render.
func (v *ResizableGroupView) SetSizes(sizes map[string]float32) {
	for id, size := range sizes {
		if _, ok := v.sizes[id]; ok && size >= 0 && finiteNumber(float64(size)) {
			v.sizes[id] = size
			delete(v.auto, id)
		}
	}
}
func (v *ResizableGroupView) SetVisible(id string, visible bool) {
	for i := range v.panels {
		if v.panels[i].ID == id {
			v.panels[i].Hidden = !visible
		}
	}
	for _, h := range v.handles {
		h.pressed, h.dragging = false, false
	}
}
func panelMax(p ResizablePanel) float32 {
	if p.Max > 0 {
		return max(p.Min, p.Max)
	}
	return math.MaxFloat32
}

func (v *ResizableGroupView) fit() []int {
	indices := []int{}
	for i, p := range v.panels {
		if !p.Hidden {
			indices = append(indices, i)
		}
	}
	if !v.measured || len(indices) == 0 {
		return indices
	}
	available := max(v.total-float32(max(len(indices)-1, 0))*handleSize, 0)
	sumMin := float64(0)
	for _, i := range indices {
		sumMin += float64(v.panels[i].Min)
	}
	if sumMin > float64(available) {
		for _, i := range indices {
			p := v.panels[i]
			v.sizes[p.ID] = float32(float64(available) * float64(p.Min) / sumMin)
			delete(v.auto, p.ID)
		}
		return indices
	}
	fixed, autos := float32(0), 0
	for _, i := range indices {
		p := v.panels[i]
		if v.auto[p.ID] {
			autos++
		} else {
			fixed += min(max(v.sizes[p.ID], p.Min), panelMax(p))
		}
	}
	autoSize := max(available-fixed, 0) / float32(max(autos, 1))
	sum := float32(0)
	for _, i := range indices {
		p := v.panels[i]
		size := v.sizes[p.ID]
		if v.auto[p.ID] {
			size = autoSize
			delete(v.auto, p.ID)
		}
		size = min(max(size, p.Min), panelMax(p))
		v.sizes[p.ID] = size
		sum += size
	}
	// Preserve earlier panes when the window changes; consume spare capacity
	// from the last pane backward. If every maximum is reached, leave trailing space.
	delta := available - sum
	for n := len(indices) - 1; n >= 0; n-- {
		p := v.panels[indices[n]]
		size := v.sizes[p.ID]
		change := min(max(delta, p.Min-size), panelMax(p)-size)
		v.sizes[p.ID] = size + change
		delta -= change
	}
	return indices
}

func (v *ResizableGroupView) resize(left, right int, desired float32) {
	if v.disabled || left >= len(v.panels) || right >= len(v.panels) {
		return
	}
	a, b := v.panels[left], v.panels[right]
	if a.Hidden || b.Hidden {
		return
	}
	sum := v.sizes[a.ID] + v.sizes[b.ID]
	lo := max(a.Min, sum-panelMax(b))
	hi := min(panelMax(a), sum-b.Min)
	hi = max(hi, 0)
	lo = min(max(lo, 0), hi)
	next := min(max(desired, lo), hi)
	if next == v.sizes[a.ID] {
		return
	}
	v.sizes[a.ID], v.sizes[b.ID] = next, sum-next
	if v.onChange != nil {
		v.onChange(v.Sizes())
	}
}

func (v *ResizableGroupView) Render(cx *el.Context) el.Element {
	v.fit()
	id := autoID("resizablegroup", v)
	box := el.Div().ID(id).Role("group").Disabled(v.disabled).Items(el.Stretch).Grow()
	if !v.vertical {
		box.Row()
	}
	previous := -1
	live := make(map[string]bool)
	for i, p := range v.panels {
		if !p.Hidden && previous >= 0 {
			left, right := previous, i
			pair := strconv.Quote(v.panels[left].ID) + "/" + strconv.Quote(p.ID)
			live[pair] = true
			state := v.handles[pair]
			if state == nil {
				state = &ResizableView{}
				v.handles[pair] = state
			}
			hid := id + "/handle/" + pair
			state.vertical, state.disabled, state.handleAppearance = v.vertical, v.disabled, v.appearance
			if v.disabled || !cx.Enabled(hid) {
				state.pressed, state.dragging = false, false
			}
			cursor := pointer.CursorColResize
			if v.vertical {
				cursor = pointer.CursorRowResize
			}
			along := func(e el.DragEvent) float32 {
				if v.vertical {
					return e.Y
				}
				return e.X
			}
			h := el.Div().ID(hid).Role("separator").Name(locale.Current().Resize + " " + v.panels[left].ID + " / " + p.ID).Value(strconv.FormatFloat(float64(v.sizes[v.panels[left].ID]), 'f', 0, 32)).NoShrink().Focusable(true).Cursor(cursor)
			painted := v.sizes[v.panels[left].ID]
			h.OnDrag(func(e el.DragEvent) {
				if e.Kind == el.DragStart {
					state.grab = along(e)
					state.pressed = true
					return
				}
				if e.Kind == el.DragEnd {
					state.pressed, state.dragging = false, false
					if e.Canceled {
						return
					}
				}
				if e.Kind == el.DragMove {
					state.dragging = true
				}
				v.resize(left, right, painted+along(e)-state.grab)
			}).OnKey(func(e el.KeyEvent) bool {
				delta := float32(0)
				switch key.Name(e.Name) {
				case key.NameLeftArrow, key.NameUpArrow:
					delta = -16
				case key.NameRightArrow, key.NameDownArrow:
					delta = 16
				case key.NameHome:
					delta = -math.MaxFloat32
				case key.NameEnd:
					delta = math.MaxFloat32
				default:
					return false
				}
				if e.State == el.KeyPress {
					v.resize(left, right, v.sizes[v.panels[left].ID]+delta)
				}
				return true
			})
			if v.vertical {
				h.H(el.Dp(handleSize))
			} else {
				h.W(el.Dp(handleSize))
			}
			state.decorateHandle(cx, hid, h)
			box.Child(h)
		}
		pane := el.Div().ID(id + "/pane/" + strconv.Quote(p.ID)).NoShrink().Items(el.Stretch).Hidden(p.Hidden)
		size := v.sizes[p.ID]
		if v.vertical {
			pane.H(el.Dp(size))
		} else {
			pane.W(el.Dp(size))
		}
		if !v.measured && v.auto[p.ID] {
			pane.Grow()
			if v.vertical {
				pane.H(el.Dp(0)).MinH(el.Dp(p.Min))
				if p.Max > 0 {
					pane.MaxH(el.Dp(panelMax(p)))
				}
			} else {
				pane.W(el.Dp(0)).MinW(el.Dp(p.Min))
				if p.Max > 0 {
					pane.MaxW(el.Dp(panelMax(p)))
				}
			}
		}
		if p.Content != nil {
			pane.Child(p.Content.Render(cx))
		}
		box.Child(pane)
		if !p.Hidden {
			previous = i
		}
	}
	for pair := range v.handles {
		if !live[pair] {
			delete(v.handles, pair)
		}
	}
	return box.Decorate(func(gtx core.C, draw func()) {
		px := gtx.Metric.PxPerDp
		if px <= 0 {
			px = 1
		}
		total := float32(gtx.Constraints.Max.X) / px
		if v.vertical {
			total = float32(gtx.Constraints.Max.Y) / px
		}
		if !v.measured || v.total != total {
			v.total, v.measured = total, true
			gtx.Execute(op.InvalidateCmd{})
		}
		draw()
	})
}
