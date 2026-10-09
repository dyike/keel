package kit

import (
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/theme"
	"image"
)

// StaticTable creates a non-virtualized table container. Build its elements
// during Render; retain stateful child views in the application.
func StaticTable() *el.DivEl {
	return el.Div().Role("table").Items(el.Stretch).Border(1, theme.Border).Rounded(theme.RadiusMd).Bg(theme.Surface)
}

// TableHeader groups heading rows. Column widths are configured on each cell.
func TableHeader() *el.DivEl { return el.Div().Role("rowgroup").Items(el.Stretch).Bg(theme.Subtle) }

// TableBody groups any number of rows, including custom row compositions.
func TableBody() *el.DivEl { return el.Div().Role("rowgroup").Items(el.Stretch) }

// TableFooter groups summary rows and draws a top rule.
func TableFooter() *el.DivEl {
	return tableRule(el.Div().Role("rowgroup").Items(el.Stretch).Bg(theme.Subtle), true)
}

// TableRow lays cells out horizontally and draws a bottom rule. Interactive
// children own their events; rows do not select or activate automatically.
func TableRow() *el.DivEl { return tableRule(el.Div().Role("row").Row().Items(el.Stretch), false) }

// TableHead creates a flexible heading cell. Use Flex(0).W(Dp(width))
// for a fixed width, matching the corresponding data cells.
func TableHead() *el.DivEl { return TableDataCell().Role("columnheader").TextColor(theme.Muted) }

// TableDataCell creates a flexible data cell. Use Flex(0).W(Dp(width)).NoShrink()
// with a fixed width, or retain W(0)/Grow for equal flexible widths.
func TableDataCell() *el.DivEl {
	return el.Div().Role("cell").Grow().W(el.Dp(0)).MinW(el.Dp(0)).P(theme.SpaceLg).Items(el.Start)
}

// TableCaption creates supporting content below the rows. It remains a child
// of the table; pass any text or richer elements through Child.
func TableCaption() *el.DivEl {
	return el.Div().Role("caption").P(theme.SpaceLg).Items(el.Center).TextSize(theme.TextSm).TextColor(theme.Muted)
}

func tableRule(box *el.DivEl, top bool) *el.DivEl {
	return box.Decorate(func(gtx core.C, draw func()) {
		draw()
		width, height := gtx.Constraints.Max.X, gtx.Constraints.Max.Y
		thickness := max(1, gtx.Dp(1))
		y := max(0, height-thickness)
		if top {
			y = 0
		}
		paint.FillShape(gtx.Ops, theme.Border, clip.Rect(image.Rect(0, y, width, min(height, y+thickness))).Op())
	})
}
