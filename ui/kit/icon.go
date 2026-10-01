package kit

import (
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
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
	IconInfo
	IconWarning
	IconError
	IconUser
	IconInbox
)

// IconView wraps a Gio vector icon. Labels belong to its containing control.
type IconView struct {
	icon  *giowidget.Icon
	size  float32
	color *color.NRGBA
}

func Icon(name IconName) *IconView {
	data := icons.ActionCheckCircle
	switch name {
	case IconInfo:
		data = icons.ActionInfo
	case IconWarning:
		data = icons.AlertWarning
	case IconError:
		data = icons.AlertError
	case IconUser:
		data = icons.SocialPerson
	case IconInbox:
		data = icons.ContentInbox
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
func (i *IconView) Size(dp float32) *IconView {
	if dp > 0 {
		i.size = dp
	}
	return i
}
func (i *IconView) Color(c color.NRGBA) *IconView { i.color = &c; return i }
func (i *IconView) Render(*el.Context) el.Element {
	c := theme.Text
	if i.color != nil {
		c = *i.color
	}
	return el.Widget(core.Func(func(gtx core.C) core.D {
		if i.icon == nil {
			return core.D{}
		}
		size := gtx.Constraints.Constrain(image.Pt(gtx.Dp(unit.Dp(i.size)), gtx.Dp(unit.Dp(i.size))))
		gtx.Constraints.Min, gtx.Constraints.Max = size, size
		return i.icon.Layout(gtx, c)
	})).Size(el.Dp(i.size))
}
