package kit

import (
	"strconv"

	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/locale"
	"github.com/dyike/keel/ui/theme"
)

// PreviousControl and NextControl create independently placeable navigation
// views. Keep the returned views across frames. Nil uses an axis-aware icon.
// A supplied button provides appearance, content, name, loading and disabled
// settings; its ID and callback are replaced without mutating the source.
func (v *CarouselView) PreviousControl(button *ButtonView) el.View {
	return &carouselControl{carousel: v, direction: -1, button: button}
}

func (v *CarouselView) NextControl(button *ButtonView) el.View {
	return &carouselControl{carousel: v, direction: 1, button: button}
}

// PaginationItem creates one independently placeable, keyboard-operable page
// button. The index is zero-based; invalid indices are disabled, never wrapped.
// Nil uses a numbered button with the selected page emphasized. Custom buttons
// retain their appearance; selected state is always exposed in semantics.
func (v *CarouselView) PaginationItem(index int, button *ButtonView) el.View {
	return &carouselControl{carousel: v, index: index, button: button}
}

type carouselControl struct {
	carousel  *CarouselView
	direction int
	index     int
	button    *ButtonView
}

func (c *carouselControl) Render(cx *el.Context) el.Element {
	v := c.carousel
	b := Button("", nil).Variant(ButtonGhost).Size(28)
	name := strconv.Itoa(c.index + 1)
	allowed := c.index >= 0 && c.index < len(v.slides) && !v.disabled
	action := func() {
		if c.index >= 0 && c.index < len(v.slides) {
			v.goTo(c.index)
		}
	}
	switch c.direction {
	case -1:
		name, allowed, action = locale.Current().PrevSlide, v.CanPrevious(), v.Previous
		icon := IconChevronLeft
		if v.vertical {
			icon = IconChevronUp
		}
		b.Icon(icon)
	case 1:
		name, allowed, action = locale.Current().NextSlide, v.CanNext(), v.Next
		icon := IconChevronRight
		if v.vertical {
			icon = IconChevronDown
		}
		b.Icon(icon)
	default:
		b.SetText(name)
		if c.index == v.current {
			b.Variant(ButtonPrimary)
		}
	}
	if c.button != nil {
		copy := *c.button
		b = &copy
	}
	if b.name == "" && b.text == "" {
		b.Name(name)
	}
	b.ID(autoID("carousel-control", c))
	b.onClick = action
	b.SetDisabled(b.disabled || !allowed)
	box := b.renderWithRadius(cx, theme.RadiusMd)
	if c.direction == 0 {
		box.Role("toggle").Selected(c.index >= 0 && c.index < len(v.slides) && c.index == v.current)
	}
	return box
}
