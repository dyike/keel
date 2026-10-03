package kit

import (
	"github.com/dyike/keel/ui/el"
	"testing"
	"time"
)

func TestCarouselControlConfigurationIsolation(t *testing.T) {
	calls := 0
	source := Button("Forward", func() { calls++ }).ID("original").Size(40).Variant(ButtonSecondary)
	car := Carousel(text("one"), text("two")).Loop(false)
	next := car.NextControl(source)
	content := car.Content()
	h := renderView(el.ViewFunc(func(cx *el.Context) el.Element { return el.Div().Child(content.Render(cx), next.Render(cx)) }), 240, 1)
	click(t, h, "Forward")
	if car.Value() != 1 || source.disabled || source.id != "original" || source.height != 40 || source.variant != ButtonSecondary {
		t.Fatal("source configuration mutated")
	}
	source.activate()
	if calls != 1 {
		t.Fatal("source callback replaced")
	}
	car.SetValue(0)
	source.SetDisabled(true)
	h.Frame()
	click(t, h, "Forward")
	if car.Value() != 0 {
		t.Fatal("custom disabled ignored")
	}
}

func TestCarouselContentAutoplay(t *testing.T) {
	c := &clock{now: time.Unix(100, 0)}
	car := Carousel(text("one"), text("two")).Loop(false).Autoplay(time.Second)
	content := car.Content()
	h := c.harness(func(cx *el.Context) el.Element { return el.Div().W(el.Dp(250)).Child(content.Render(cx)) })
	h.Move(390, 290)
	c.advance(h, time.Second)
	h.Frame()
	if car.Value() != 1 {
		t.Fatal("composed content did not autoplay")
	}
}
