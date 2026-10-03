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
// dot per slide. Arrow keys follow orientation; Home/End select the endpoints.
// Autoplay advances on a timer,
// paused while the pointer is over it and off with reduced motion.
type CarouselView struct {
	slides   []el.View
	current  int
	autoplay time.Duration
	disabled bool
	height   float32
	vertical bool
	onChange func(int)
}

func Carousel(slides ...el.View) *CarouselView { return &CarouselView{slides: slides, height: 200} }

// Height sets the slide area height in dp, 200 by default.
func (v *CarouselView) Height(dp float32) *CarouselView {
	if dp > 0 && finiteNumber(float64(dp)) {
		v.height = dp
	}
	return v
}

// Vertical arranges navigation above/below its indicators beside the stage.
// Up/Down replace Left/Right; changing orientation preserves selection and focus.
func (v *CarouselView) Vertical(on bool) *CarouselView { v.vertical = on; return v }

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
	stage := el.Div().ID(id + "/stage").H(el.Dp(v.height)).Rounded(theme.RadiusLg).Bg(theme.Subtle).Items(el.Stretch).Justify(el.Center)
	if v.current < len(v.slides) && v.slides[v.current] != nil {
		stage.Child(v.slides[v.current].Render(cx))
	}
	dots := el.Div().ID(id + "/dots").Row().Gap(theme.SpaceSm).Justify(el.Center)
	for i := range v.slides {
		i := i
		c := theme.Border
		if i == v.current {
			c = theme.Primary
		}
		dots.Child(el.Div().Name(strconv.Itoa(i + 1)).Size(el.Dp(8)).Rounded(theme.RadiusSm).Bg(c).CursorPointer().Focusable(false).
			OnClick(func() { v.goTo(i) }))
	}
	prevIcon, nextIcon := IconChevronLeft, IconChevronRight
	indicators := el.Div().ID(id + "/indicators").Grow().Items(el.Center).Child(dots)
	nav := el.Div().ID(id + "/nav").Row().Items(el.Center).Gap(theme.SpaceMd)
	root := el.Div().ID(id).WFull().Disabled(v.disabled).Role("group").Name(strconv.Itoa(v.current+1) + "/" + strconv.Itoa(len(v.slides))).
		Gap(theme.SpaceMd).Items(el.Stretch).Rounded(theme.RadiusLg).Focusable(true).
		FocusStyle(func(s *el.Style) { s.BorderColor(theme.Primary) })
	if v.vertical {
		prevIcon, nextIcon = IconChevronUp, IconChevronDown
		root.Row()
		stage.Grow().W(el.Dp(0))
		dots.Col()
		indicators.H(el.Dp(0)).ScrollY().Justify(el.Center)
		nav.Col().H(el.Dp(max(v.height, 72))).NoShrink()
	}
	nav.Child(
		Button("", func() { v.goTo(v.current - 1) }).ID(id+"/previous").Name(text.PrevSlide).Icon(prevIcon).Variant(ButtonGhost).Size(28).Render(cx),
		indicators,
		Button("", func() { v.goTo(v.current + 1) }).ID(id+"/next").Name(text.NextSlide).Icon(nextIcon).Variant(ButtonGhost).Size(28).Render(cx),
	)
	return root.
		OnKey(func(e el.KeyEvent) bool {
			next, previous := key.NameRightArrow, key.NameLeftArrow
			if v.vertical {
				next, previous = key.NameDownArrow, key.NameUpArrow
			}
			target := v.current
			switch key.Name(e.Name) {
			case next:
				target++
			case previous:
				target--
			case key.NameHome:
				target = 0
			case key.NameEnd:
				target = len(v.slides) - 1
			default:
				return false
			}
			if e.State == el.KeyPress {
				v.goTo(target)
			}
			return true
		}).
		Child(stage, nav)
}

type carouselKey struct {
	id  string
	cur int
}
