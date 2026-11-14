package kit

import (
	"image/color"
	"math"
	"strconv"
	"time"

	"github.com/dyike/keel/gio/io/key"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/theme"
)

// GooeyDockItem is a launch action. Active is an application-owned indicator.
type GooeyDockItem struct {
	ID, Label        string
	Icon             IconName
	IconView         *IconView
	Action           func()
	Active, Disabled bool
}
type gooeyDockState struct {
	size, lift, background, iconScale, ink el.AnimatedFloat
	center, width, top, target             float32
}

// GooeyDockView magnifies neighbouring actions with a cosine falloff and springs.
type GooeyDockView struct {
	items    []GooeyDockItem
	states   []gooeyDockState
	disabled bool
	pointer  bool
	hover    int
	lastTick time.Time
	onHover  func(id string)
	surface  *color.NRGBA
}

func GooeyDock(items ...GooeyDockItem) *GooeyDockView {
	v := &GooeyDockView{hover: -1}
	v.SetItems(items...)
	return v
}

// SetItems replaces actions and resets their animation state. Empty or duplicate
// IDs are ignored. Keep the view rather than rebuilding it inside Render.
func (v *GooeyDockView) SetItems(items ...GooeyDockItem) {
	seen := map[string]bool{}
	v.items = nil
	for _, it := range items {
		if it.ID == "" || seen[it.ID] {
			continue
		}
		seen[it.ID] = true
		v.items = append(v.items, it)
	}
	v.states = make([]gooeyDockState, len(v.items))
	for i := range v.states {
		s := &v.states[i]
		s.size.Set(40)
		s.iconScale.Set(1)
		s.ink.Set(.45)
		s.lift.Set(0)
		s.background.Set(0)
		s.center = 33 + float32(i)*42
		s.width = 40
		s.top = 53
	}
	v.hover = -1
	v.pointer = false
}
func (v *GooeyDockView) SetDisabled(on bool) { v.disabled = on }

// Surface specifies the opaque surface behind the dock for browser-style sRGB
// compositing. The default is theme.Bg; it does not paint the parent surface.
func (v *GooeyDockView) Surface(c color.NRGBA) *GooeyDockView {
	c.A = 255
	v.surface = &c
	return v
}

// Browser CSS blends these translucent fills in sRGB, while Gio blends in
// linear light. Resolve the reference colors against the known surface first.
func gooeyComposite(ink, background color.NRGBA, alpha float32) color.NRGBA {
	channel := func(a, b uint8) uint8 { return uint8(math.Round(float64(float32(a)*alpha + float32(b)*(1-alpha)))) }
	return color.NRGBA{R: channel(ink.R, background.R), G: channel(ink.G, background.G), B: channel(ink.B, background.B), A: 255}
}

// SetActive changes the indicator without activating the action.
func (v *GooeyDockView) SetActive(id string, on bool) {
	for i := range v.items {
		if v.items[i].ID == id {
			v.items[i].Active = on
		}
	}
}

// OnHover observes pointer transitions between actions, throttled to 80ms.
// An application may attach the reference's soft tick sound here.
func (v *GooeyDockView) OnHover(fn func(id string)) *GooeyDockView { v.onHover = fn; return v }

func gooeyFalloff(distance float32) float32 {
	d := math.Abs(float64(distance))
	if d >= 150 {
		return 0
	}
	return float32((1 + math.Cos(d/150*math.Pi)) / 2)
}

var gooeySpring = el.Animation{Stiffness: 200, Damping: 14, Mass: .1, RestSpeed: .01, RestDelta: .005}

func (v *GooeyDockView) move(now time.Time, x, y float32) {
	if v.disabled {
		return
	}
	hovered := -1
	maxSize, width := float32(40), float32(26)
	for i := range v.states {
		s := &v.states[i]
		maxSize = max(maxSize, s.width)
		width += s.width
		if i > 0 {
			width += 2
		}
		if !v.items[i].Disabled && math.Abs(float64(x-s.center)) <= float64(s.width)/2 && y >= s.top && y <= s.top+s.width {
			hovered = i
		}
	}
	// The reserved label space is not an extension of the dock's hit area.
	v.pointer = hovered >= 0 || (x >= 0 && x <= width && y >= 44 && y <= 64+maxSize)
	if !v.pointer {
		v.hover = -1
		return
	}
	for i := range v.states {
		v.states[i].target = gooeyFalloff(x - v.states[i].center)
	}
	if hovered != v.hover && hovered >= 0 && v.onHover != nil && (v.lastTick.IsZero() || now.Sub(v.lastTick) >= 80*time.Millisecond) {
		v.lastTick = now
		v.onHover(v.items[hovered].ID)
	}
	v.hover = hovered
}
func (v *GooeyDockView) Render(cx *el.Context) el.Element {
	id := autoID("gooeydock", v)
	pointer := v.pointer && cx.Hovered(id) && !v.disabled
	ink := color.NRGBA{A: 255}
	if int(theme.Text.R)+int(theme.Text.G)+int(theme.Text.B) > 384 {
		ink.R, ink.G, ink.B = 255, 255, 255
	}
	surface := theme.Bg
	if v.surface != nil {
		surface = *v.surface
	}
	barColor := gooeyComposite(ink, surface, .02)
	bar := el.Div().Row().Items(el.End).Gap(2).Pt(8).Px(12).Pb(10).Rounded(18).Border(1, gooeyComposite(ink, barColor, .06)).Bg(barColor).NoShrink()
	// Reserve overflow above the bar for the lift and label; the dock itself
	// keeps the reference's padding and dimensions.
	outer := el.Div().ID(id).Role("toolbar").Name("Gooey Dock").Pt(44).Items(el.Start).Disabled(v.disabled).
		OnPointerMove(func(x, y float32) { v.move(cx.Now(), x, y) })
	offset := float32(13)
	maxSize := float32(40)
	for i, it := range v.items {
		state := &v.states[i]
		itemID := id + "/" + it.ID
		disabled := v.disabled || it.Disabled
		hovered := !disabled && cx.Hovered(itemID)
		focused := !disabled && cx.FocusVisible(itemID)
		target := float32(0)
		if pointer && !disabled {
			target = state.target
		}
		if focused {
			target = 1
		}
		sample := func(value *el.AnimatedFloat, target float32, tr el.Animation) float32 {
			return value.Sample(cx, target, tr)
		}
		size := sample(&state.size, 40+18*target, gooeySpring)
		lift := sample(&state.lift, -14*target, gooeySpring)
		bg := sample(&state.background, .08*target, gooeySpring)
		iconScale := sample(&state.iconScale, 1+(size-40)/18*.2, gooeySpring)
		opacity := float32(.45)
		if hovered || focused {
			opacity = .8
		}
		if disabled {
			opacity = .2
		}
		foreground := sample(&state.ink, opacity, el.Tween(150*time.Millisecond, el.CubicBezier(.25, .1, .25, 1)))
		state.center, state.width = offset+size/2, size
		maxSize = max(maxSize, size)
		offset += size + 2
		cellColor := gooeyComposite(ink, barColor, bg)
		item := el.Div().ID(itemID).Role("button").Name(it.Label).Selected(it.Active).Size(el.Dp(size)).NoShrink().Rounded(12).Bg(cellColor).
			Translate(0, lift).Center().CursorPointer().Focusable(true).Disabled(disabled).
			FocusStyle(func(s *el.Style) { s.BorderColor(theme.Primary) }).
			OnClick(func() {
				if !v.disabled && !it.Disabled && it.Action != nil {
					it.Action()
				}
			}).
			OnKey(func(e el.KeyEvent) bool {
				if e.Modifiers != 0 {
					return false
				}
				next := i
				switch key.Name(e.Name) {
				case key.NameLeftArrow:
					for n := i - 1; n >= 0; n-- {
						if !v.items[n].Disabled {
							next = n
							break
						}
					}
				case key.NameRightArrow:
					for n := i + 1; n < len(v.items); n++ {
						if !v.items[n].Disabled {
							next = n
							break
						}
					}
				case key.NameHome:
					for n := range v.items {
						if !v.items[n].Disabled {
							next = n
							break
						}
					}
				case key.NameEnd:
					for n := len(v.items) - 1; n >= 0; n-- {
						if !v.items[n].Disabled {
							next = n
							break
						}
					}
				default:
					return false
				}
				if e.State == el.KeyPress {
					cx.Focus(id + "/" + v.items[next].ID)
				}
				return true
			})
		icon := it.IconView
		if icon == nil {
			icon = Icon(it.Icon)
		}
		item.Child(el.Div().Size(el.Dp(20)).Center().AnimatePose(el.Pose{Scale: iconScale}, el.Instant).Child(icon.Size(20).Color(gooeyComposite(ink, cellColor, foreground)).Render(cx)))
		if it.Active {
			item.Child(el.Div().Absolute().Bottom(2).Left(size/2 - 1.5).Size(el.Dp(3)).Rounded(1.5).Bg(gooeyComposite(ink, cellColor, .35)))
		}
		presence := cx.Presence(itemID+"/label", hovered || focused, el.PresenceOptions{Enter: el.Pose{Y: 4, Fade: 1}, Exit: el.Pose{Y: 4, Fade: 1}, Animation: el.Tween(150*time.Millisecond, el.EaseOut)})
		if presence.Visible {
			label := el.Div().Presence(presence).TextSize(11).Medium().TextColor(gooeyComposite(ink, surface, .6)).Child(el.Text(it.Label).MaxLines(1))
			cx.Overlay(itemID+"/label", el.Anchored(itemID, label).Placement(el.Top, el.Center).Offset(8))
		}
		bar.Child(item)
	}
	for i := range v.states {
		s := &v.states[i]
		s.top = 53 + maxSize - s.width + s.lift.Get()
	}
	if !pointer {
		v.hover = -1
	}
	return outer.Child(bar).Value(strconv.Itoa(len(v.items)))
}
