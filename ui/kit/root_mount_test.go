package kit

import (
	"testing"

	"github.com/dyike/keel/third_party/gio/io/key"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/internal/uitest"
)

func TestShowAndWindowNotifierNeedNoTree(t *testing.T) {
	dlg := Dialog("Rename").Body(text("New name"))
	sheet := Sheet(el.Right, "Details").Body(text("Order 42"))
	var notifiers []*NotifierView
	h := uitest.New(el.Root(viewFunc(func(cx *el.Context) el.Element {
		// The page holds only buttons; nothing renders the overlays.
		return el.Div().Row().Gap(8).Child(
			Button("Open dialog", func() { dlg.Show(cx) }).Render(cx),
			Button("Open sheet", func() { sheet.Show(cx) }).Render(cx),
			Button("Notify", func() {
				for range 2 {
					n := WindowNotifier(cx)
					notifiers = append(notifiers, n)
					n.Notify(Notice{Title: "Saved"})
				}
			}).Render(cx),
		)
	})))
	click(t, h, "Open dialog")
	h.Frame()
	if !shown(h, "New name") {
		t.Fatal("shown dialog")
	}
	h.Key(key.NameEscape, 0)
	h.Frame()
	h.Frame()
	if dlg.Value() || shown(h, "New name") {
		t.Fatal("dialog closes with Esc")
	}
	click(t, h, "Open sheet")
	h.Frame()
	if !shown(h, "Order 42") {
		t.Fatal("shown sheet")
	}
	sheet.SetValue(false)
	h.Frame()
	h.Frame()
	click(t, h, "Notify")
	h.Frame()
	if len(notifiers) != 2 || notifiers[0] != notifiers[1] || notifiers[0].Len() != 2 || !shown(h, "Saved") {
		t.Fatal("one notifier per window", len(notifiers), notifiers[0].Len())
	}
	// The dialog can open again after it left the window.
	click(t, h, "Open dialog")
	h.Frame()
	if !shown(h, "New name") {
		t.Fatal("dialog reopens")
	}
}

func TestMountOrderReplaceAndLeafRoot(t *testing.T) {
	var cx *el.Context
	h := uitest.New(el.Root(viewFunc(func(c *el.Context) el.Element { cx = c; return el.Text("leaf root") })))
	cx.Mount("a", text("first"))
	h.Frame()
	h.Frame()
	if !shown(h, "leaf root") || !shown(h, "first") {
		t.Fatal("a text root still hosts mounts")
	}
	cx.Mount("a", text("second"))
	h.Frame()
	h.Frame()
	if shown(h, "first") || !shown(h, "second") || !cx.Mounted("a") {
		t.Fatal("mounting a key again replaces it")
	}
	cx.Unmount("a")
	h.Frame()
	h.Frame()
	if shown(h, "second") || cx.Mounted("a") || cx.MountedView("a") != nil {
		t.Fatal("unmount")
	}
}
