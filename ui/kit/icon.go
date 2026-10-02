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

// IconName names a built-in icon. The zero value IconNone draws nothing and
// takes no space, so an optional icon field can be left unset.
type IconName uint8

const (
	IconNone IconName = iota
	IconCheck
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
	IconDone // a plain check mark, e.g. inside a checkbox
	IconChevronLeft
	IconMinus
	IconStar
	IconStarOutline
	IconCalendar
	IconClock
	IconSettings
	IconBell
	IconLock
	IconFolder
	IconFile
	IconArchive
	IconReceipt
	IconHome
	IconTrash
	IconEdit
)

// IconView wraps a Gio vector icon. Labels belong to its containing control.
type IconView struct {
	icon  *giowidget.Icon
	size  float32
	color *color.NRGBA
}

var iconData = [...][]byte{
	IconCheck:        icons.ActionCheckCircle,
	IconClose:        icons.NavigationClose,
	IconPlus:         icons.ContentAdd,
	IconSearch:       icons.ActionSearch,
	IconCopy:         icons.ContentContentCopy,
	IconChevronDown:  icons.NavigationExpandMore,
	IconChevronRight: icons.NavigationChevronRight,
	IconInfo:         icons.ActionInfo,
	IconWarning:      icons.AlertWarning,
	IconError:        icons.AlertError,
	IconUser:         icons.SocialPerson,
	IconInbox:        icons.ContentInbox,
	IconDone:         icons.ActionDone,
	IconChevronLeft:  icons.NavigationChevronLeft,
	IconMinus:        icons.ContentRemove,
	IconStar:         icons.ToggleStar,
	IconStarOutline:  icons.ToggleStarBorder,
	IconCalendar:     icons.ActionDateRange,
	IconClock:        icons.ActionSchedule,
	IconSettings:     icons.ActionSettings,
	IconBell:         icons.SocialNotifications,
	IconLock:         icons.ActionLock,
	IconFolder:       icons.FileFolder,
	IconFile:         icons.ActionDescription,
	IconArchive:      icons.ContentArchive,
	IconReceipt:      icons.ActionReceipt,
	IconHome:         icons.ActionHome,
	IconTrash:        icons.ActionDelete,
	IconEdit:         icons.EditorModeEdit,
}

// decoded caches parsed icons: Icon is called on every Render. Only touched
// under the frame lock.
var decoded [len(iconData)]*giowidget.Icon

func Icon(name IconName) *IconView {
	if name == IconNone || int(name) >= len(iconData) {
		return &IconView{size: 18}
	}
	if decoded[name] == nil {
		ic, err := giowidget.NewIcon(iconData[name])
		if err != nil {
			panic(err) // All bundled icon data is known at build time.
		}
		decoded[name] = ic
	}
	return &IconView{icon: decoded[name], size: 18}
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
