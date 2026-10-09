package kit

import (
	"github.com/dyike/keel/third_party/gio/io/key"
	"github.com/dyike/keel/third_party/gio/unit"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/internal/uitest"
	"image"
	"testing"
	"time"
)

func TestNotifierContentAndActions(t *testing.T) {
	for _, scale := range []int{1, 2} {
		n := Notifier()
		clicks := 0
		disabled := false
		id := n.Notify(Notice{Title: "Rich notice", Body: "Fallback body", Content: Label("Rich body"), Action: Button("Retry", func() { clicks++ }), Timeout: -1})
		h := renderNotifierContent(viewFunc(func(cx *el.Context) el.Element { return el.Div().Disabled(disabled).Child(n.Render(cx)) }), 300, scale)
		h.Frame()
		if !shown(h, "Rich body") || shown(h, "Fallback body") {
			t.Fatal("content replacement")
		}
		card, action := bounds(h, "Rich notice"), bounds(h, "Retry")
		if card.Min.X < 0 || card.Max.X > 300*scale || !action.In(card) || action.Min.Y < bounds(h, "Rich body").Max.Y {
			t.Fatal("content layout", card, action)
		}
		click(t, h, "Retry")
		h.Frame()
		if clicks != 1 || n.Len() != 1 {
			t.Fatal("action should not automatically dismiss")
		}
		disabled = true
		h.Frame()
		h.Click(center(action))
		h.Frame()
		if clicks != 1 {
			t.Fatal("disabled action fired")
		}
		disabled = false
		n.Update(id, Notice{Title: "Rich notice", Body: "Fallback body", Action: Button("Done", func() { n.Dismiss(id) }), Timeout: -1})
		h.Frame()
		if shown(h, "Rich body") || !shown(h, "Fallback body") || shown(h, "Retry") {
			t.Fatal("content reset")
		}
		click(t, h, "Done")
		h.Frame()
		if n.Len() != 0 {
			t.Fatal("action cannot dismiss")
		}
	}
}

func TestNotifierRichContentFocusPausesTimeout(t *testing.T) {
	c := &clock{now: time.Unix(100, 0)}
	n := Notifier()
	calls := 0
	var cx *el.Context
	n.Notify(Notice{Title: "Interactive", Content: Button("Body action", func() { calls++ }), Timeout: 5 * time.Second})
	h := c.harness(func(ctx *el.Context) el.Element { cx = ctx; return el.Div().Child(n.Render(ctx)) })
	c.advance(h, 2*time.Second)
	h.Router.MoveFocus(key.FocusForward)
	h.Frame()
	h.Frame()
	h.Key(key.NameReturn, 0)
	h.Frame()
	if calls != 1 {
		t.Fatal("rich content keyboard action")
	}
	c.advance(h, 20*time.Second)
	if n.Len() != 1 {
		t.Fatal("focused rich content timed out")
	}
	cx.Focus("")
	h.Frame()
	h.Frame()
	c.advance(h, 3*time.Second)
	h.Frame()
	if n.Len() != 0 {
		t.Fatal("remaining timeout lost")
	}
}

func TestNotifierQueuedContentRendersOnlyWhenVisible(t *testing.T) {
	n := Notifier()
	for i := 0; i < MaxNotifications; i++ {
		n.Notify(Notice{Title: "Earlier", Timeout: -1})
	}
	renders := 0
	n.Notify(Notice{Title: "Queued", Content: viewFunc(func(*el.Context) el.Element { renders++; return el.Text("Lazy content") }), Timeout: -1})
	h := renderNotifierContent(n, 400, 1)
	if renders != 0 {
		t.Fatal("queued content rendered")
	}
	n.Dismiss(1)
	h.Frame()
	h.Frame()
	if renders == 0 || !shown(h, "Lazy content") {
		t.Fatal("released content missing")
	}
}

func renderNotifierContent(v el.View, width, scale int) *uitest.Harness {
	root := el.Root(v)
	return uitest.NewFunc(func(gtx core.C) {
		gtx.Metric = unit.Metric{PxPerDp: float32(scale), PxPerSp: float32(scale)}
		gtx.Constraints.Max = image.Pt(width*scale, 1000*scale)
		root.Layout(gtx)
	})
}

func TestNotifierContentSurvivesActionChanges(t *testing.T) {
	n := Notifier()
	input := Input("Notice input")
	input.SetValue("draft")
	item := Notice{Title: "Editable", Content: input, Timeout: -1}
	id := n.Notify(item)
	var cx *el.Context
	h := renderNotifierContent(viewFunc(func(ctx *el.Context) el.Element { cx = ctx; return el.Div().Child(n.Render(ctx)) }), 600, 1)
	cx.Focus(input.FocusID())
	h.Frame()
	h.Frame()
	for _, action := range []el.View{Button("Save", nil), nil, Button("Replace", nil)} {
		item.Action = action
		n.Update(id, item)
		h.Frame()
		h.Frame()
		if !cx.Focused(input.FocusID()) || input.Value() != "draft" {
			t.Fatal("content state lost when action changed")
		}
	}
}
