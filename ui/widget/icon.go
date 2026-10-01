package widget

import (
	"image"
	"image/color"

	"gioui.org/unit"
	giowidget "gioui.org/widget"
	"github.com/dyike/keel/ui/theme"
	"golang.org/x/exp/shiny/materialdesign/icons"
)

type IconName uint8

const (
	IconCheck IconName = iota
	IconClose
	IconPlus
	IconSearch
	IconCopy
	IconChevronDown
	IconChevronRight
)

// IconView wraps a Gio vector icon. Labels belong to its containing control.
type IconView struct {
	icon  *giowidget.Icon
	size  unit.Dp
	color *color.NRGBA
}

func Icon(name IconName) *IconView {
	data := icons.ActionCheckCircle
	switch name {
	case IconClose:
		data = icons.NavigationClose
	case IconPlus:
		data = icons.ContentAdd
	case IconSearch:
		data = icons.ActionSearch
	case IconCopy:
		data = icons.ContentContentCopy
	case IconChevronDown:
		data = icons.NavigationExpandMore
	case IconChevronRight:
		data = icons.NavigationChevronRight
	}
	ic, err := giowidget.NewIcon(data)
	if err != nil {
		panic(err)
	} // All bundled icon data is known at build time.
	return &IconView{icon: ic, size: 18}
}

// VectorIcon supports custom Gio icons, including icons decoded by widget.NewIcon.
func VectorIcon(icon *giowidget.Icon) *IconView { return &IconView{icon: icon, size: 18} }
func (i *IconView) Size(dp unit.Dp) *IconView {
	if dp > 0 {
		i.size = dp
	}
	return i
}
func (i *IconView) Color(c color.NRGBA) *IconView { i.color = &c; return i }
func (i *IconView) Layout(gtx C) D {
	if i.icon == nil {
		return D{}
	}
	size := gtx.Constraints.Constrain(image.Pt(gtx.Dp(i.size), gtx.Dp(i.size)))
	gtx.Constraints.Min, gtx.Constraints.Max = size, size
	c := theme.Text
	if i.color != nil {
		c = *i.color
	}
	return i.icon.Layout(gtx, c)
}
