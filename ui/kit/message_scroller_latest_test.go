package kit

import (
	"testing"
	"time"

	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/internal/uitest"
	"github.com/dyike/keel/ui/theme"
)

func TestMessageScrollerLatestCustomization(t *testing.T) {
	called := 0
	custom := Button("查看新消息", func() { called++ }).ID("application-button").Size(40).Variant(ButtonPrimary)
	sc := MessageScroller(variableKeys(100), 60, func(_ *el.Context, i int) el.Element { return el.Div().H(el.Dp(40)) }).LatestTransition(0)
	sc.LatestRenderer(func(*ButtonView) *ButtonView { return custom })
	var cx *el.Context
	h := uitest.New(el.Root(el.ViewFunc(func(ctx *el.Context) el.Element { cx = ctx; return sc.Render(ctx) })))
	settle(h)
	sc.ScrollToMessage("30")
	settle(h)
	if !shown(h, "查看新消息") {
		t.Fatal("custom button missing")
	}
	click(t, h, "查看新消息")
	settle(h)
	if sc.IsScrolledUp(cx) || called != 0 || custom.id != "application-button" {
		t.Fatal("action or identity ownership")
	}
	custom.activate()
	if called != 1 {
		t.Fatal("source button mutated")
	}
	sc.LatestRenderer(nil).LatestLabel("最新内容")
	sc.ScrollToMessage("30")
	settle(h)
	if !shown(h, "最新内容") {
		t.Fatal("label override")
	}
	sc.LatestButton(false)
	settle(h)
	if shown(h, "最新内容") || !sc.IsScrolledUp(cx) {
		t.Fatal("hide changed scrolling")
	}
	sc.LatestButton(true).SetDisabled(true)
	settle(h)
	click(t, h, "最新内容")
	settle(h)
	if !sc.IsScrolledUp(cx) {
		t.Fatal("disabled control activated")
	}
}

func TestMessageScrollerLatestTransition(t *testing.T) {
	old := el.ReducedMotion()
	theme.SetReducedMotion(false)
	defer theme.SetReducedMotion(old)
	now := time.Unix(100, 0)
	visible := false
	sc := MessageScroller(nil, 60, nil).LatestTransition(time.Second)
	root := el.Root(el.ViewFunc(func(cx *el.Context) el.Element {
		box := el.Div().W(el.Dp(400)).H(el.Dp(300))
		sc.renderLatest(cx, box, visible)
		return box
	}))
	h := uitest.NewFunc(func(gtx core.C) { gtx.Now = now; root.Layout(gtx) })
	visible = true
	h.Frame()
	now = now.Add(500 * time.Millisecond)
	h.Frame()
	if sc.latestMotion.value != 0.5 {
		t.Fatal("enter", sc.latestMotion.value)
	}
	visible = false
	h.Frame()
	now = now.Add(500 * time.Millisecond)
	h.Frame()
	if sc.latestMotion.value != 0.25 {
		t.Fatal("reverse", sc.latestMotion.value)
	}
	now = now.Add(500 * time.Millisecond)
	h.Frame()
	if shown(h, "回到最新") {
		t.Fatal("finished leave still mounted")
	}
	visible = true
	sc.LatestTransition(0)
	h.Frame()
	if sc.latestMotion.value != 1 {
		t.Fatal("zero duration")
	}
	sc.LatestTransition(time.Second)
	theme.SetReducedMotion(true)
	visible = false
	h.Frame()
	if sc.latestMotion.value != 0 || shown(h, "回到最新") {
		t.Fatal("reduced motion")
	}
}
