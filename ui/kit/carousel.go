package kit

import (
	"strconv"
	"time"

	"gioui.org/io/key"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/locale"
	"github.com/dyike/keel/ui/theme"
)

// CarouselView shows one slide at a time with previous / next buttons and a
// dot per slide. Focused, ← → change slides. Autoplay advances on a timer,
// paused while the pointer is over it and off with reduced motion.
type CarouselView struct {
	slides   []el.View
	current  int
	autoplay time.Duration
	disabled bool
	height   float32
	onChange func(int)
}

func Carousel(slides ...el.View) *CarouselView { return &CarouselView{slides: slides, height: 200} }

// Height sets the slide area height in dp, 200 by default.
func (v *CarouselView) Height(dp float32) *CarouselView {
	if dp > 0 {
		v.height = dp
	}
	return v
}

// Autoplay advances every d; 0 turns it off.
func (v *CarouselView) Autoplay(d time.Duration) *CarouselView { v.autoplay = d; return v }
func (v *CarouselView) OnChange(fn func(int)) *CarouselView    { v.onChange = fn; return v }

// SetDisabled blocks navigation and autoplay; programmatic SetValue still works.
func (v *CarouselView) SetDisabled(on bool) { v.disabled = on }
func (v *CarouselView) Value() int          { return v.current }

// SetValue shows slide i without calling OnChange.
func (v *CarouselView) SetValue(i int) { v.current = min(max(i, 0), max(len(v.slides)-1, 0)) }

func (v *CarouselView) goTo(i int) {
	if v.disabled || len(v.slides) == 0 {
		return
	}
	i = (i%len(v.slides) + len(v.slides)) % len(v.slides)
	if i == v.current {
		return
	}
	v.current = i
	if v.onChange != nil {
		v.onChange(i)
	}
}

func (v *CarouselView) Render(cx *el.Context) el.Element {
	id := autoID("carousel", v)
	text := locale.Current()
	if !v.disabled && v.autoplay > 0 && len(v.slides) > 1 && !cx.Hovered(id) && !el.ReducedMotion() {
		cur := v.current
		cx.AfterEnabled(id, carouselKey{id, cur}, v.autoplay, func() { v.goTo(cur + 1) })
	}
	stage := el.Div().H(el.Dp(v.height)).Rounded(8).Bg(theme.Subtle).Items(el.Stretch).Justify(el.Center)
	if v.current < len(v.slides) && v.slides[v.current] != nil {
		stage.Child(v.slides[v.current].Render(cx))
	}
	dots := el.Div().Row().Gap(6).Justify(el.Center)
	for i := range v.slides {
		i := i
		c := theme.Border
		if i == v.current {
			c = theme.Primary
		}
		dots.Child(el.Div().Name(strconv.Itoa(i + 1)).Size(el.Dp(8)).Rounded(4).Bg(c).CursorPointer().Focusable(false).
			OnClick(func() { v.goTo(i) }))
	}
	nav := el.Div().Row().Items(el.Center).Gap(8).Child(
		Button("", func() { v.goTo(v.current - 1) }).Name(text.PrevSlide).Icon(IconChevronLeft).Variant(ButtonGhost).Size(28).Render(cx),
		el.Div().Grow().Items(el.Center).Child(dots),
		Button("", func() { v.goTo(v.current + 1) }).Name(text.NextSlide).Icon(IconChevronRight).Variant(ButtonGhost).Size(28).Render(cx),
	)
	return el.Div().ID(id).Disabled(v.disabled).Role("group").Name(strconv.Itoa(v.current+1)+"/"+strconv.Itoa(len(v.slides))).
		Gap(8).Items(el.Stretch).Rounded(8).Focusable(true).
		FocusStyle(func(s *el.Style) { s.BorderColor(theme.Primary) }).
		OnKey(func(e el.KeyEvent) bool {
			d := 0
			switch key.Name(e.Name) {
			case key.NameRightArrow:
				d = 1
			case key.NameLeftArrow:
				d = -1
			default:
				return false
			}
			if e.State == el.KeyPress {
				v.goTo(v.current + d)
			}
			return true
		}).
		Child(stage, nav)
}

type carouselKey struct {
	id  string
	cur int
}
