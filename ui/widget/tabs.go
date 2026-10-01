package widget

import (
	"image"

	"gioui.org/font"
	"gioui.org/io/pointer"
	"gioui.org/io/semantic"
	giolayout "gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/widget"
	"gioui.org/widget/material"

	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/layout"
	"github.com/dyike/keel/ui/theme"
)

// TabsView shows one page at a time under a row of tab titles.
type TabsView struct {
	titles   []string
	pages    []core.Widget
	current  int
	disabled bool
	clicks   []widget.Clickable
	onChange func(index int)
}

// Tabs creates an empty tab set; add pages with Add.
func Tabs() *TabsView { return &TabsView{} }

// Add appends a page with its title.
func (t *TabsView) Add(title string, page core.Widget) *TabsView {
	t.titles = append(t.titles, title)
	t.pages = append(t.pages, page)
	t.clicks = append(t.clicks, widget.Clickable{})
	return t
}

func (t *TabsView) OnChange(fn func(index int)) *TabsView { t.onChange = fn; return t }
func (t *TabsView) Current() int                          { return t.current }
func (t *TabsView) SetDisabled(v bool)                    { t.disabled = v }

// SetCurrent shows page i without calling OnChange.
func (t *TabsView) SetCurrent(i int) {
	if i >= 0 && i < len(t.pages) {
		t.current = i
	}
}

func (t *TabsView) Layout(gtx C) D {
	if t.disabled {
		gtx = gtx.Disabled()
	}
	gtx.Constraints.Min.X = gtx.Constraints.Max.X
	for i := range t.clicks {
		for t.clicks[i].Clicked(gtx) {
			if i != t.current {
				t.current = i
				core.Call(gtx, func() {
					if t.onChange != nil {
						t.onChange(i)
					}
				})
			}
		}
	}
	heads := make([]giolayout.FlexChild, len(t.titles))
	for i, title := range t.titles {
		heads[i] = giolayout.Rigid(func(gtx C) D { return t.head(gtx, i, title) })
	}
	return giolayout.Flex{Axis: giolayout.Vertical}.Layout(gtx,
		giolayout.Rigid(func(gtx C) D { return giolayout.Flex{}.Layout(gtx, heads...) }),
		giolayout.Rigid(layout.Divider().Layout),
		giolayout.Rigid(giolayout.Spacer{Height: 12}.Layout),
		giolayout.Rigid(func(gtx C) D {
			if len(t.pages) == 0 {
				return D{}
			}
			return t.pages[t.current].Layout(gtx)
		}),
	)
}

func (t *TabsView) head(gtx C, i int, title string) D {
	active := i == t.current
	return t.clicks[i].Layout(gtx, func(gtx C) D {
		pointer.CursorPointer.Add(gtx.Ops)
		return core.Semantic(gtx, func(gtx C) D {
			d := giolayout.Inset{Top: 8 + theme.CJKNudge, Bottom: 8 - theme.CJKNudge, Left: 14, Right: 14}.Layout(gtx, func(gtx C) D {
				lb := material.Label(theme.Material, theme.BodySize, title)
				lb.Color = theme.Muted
				if active {
					lb.Color, lb.Font.Weight = theme.Primary, font.Bold
				}
				return lb.Layout(gtx)
			})
			if active { // underline
				r := image.Rect(0, d.Size.Y-gtx.Dp(2), d.Size.X, d.Size.Y)
				paint.FillShape(gtx.Ops, theme.Primary, clip.Rect(r).Op())
			}
			return d
		}, semantic.Button, core.Role("tab"), semantic.LabelOp(title), semantic.SelectedOp(active), semantic.EnabledOp(gtx.Enabled()))
	})
}
