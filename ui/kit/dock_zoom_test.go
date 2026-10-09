package kit

import (
	"encoding/json"
	"image"
	"testing"

	"github.com/dyike/keel/third_party/gio/io/key"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/internal/uitest"
)

func TestDockZoomMenuEscAndLayout(t *testing.T) {
	var saved DockLayout
	d := Dock(text("编辑器")).
		Panel(DockPanel{ID: "files", Title: "文件", View: text("文件树")}, DockLeft).
		Panel(DockPanel{ID: "term", Title: "终端", View: text("$ go test")}, DockBottom).
		OnLayoutChange(func(l DockLayout) { saved = l })
	root := el.Root(d)
	h := uitest.NewFunc(func(gtx core.C) { gtx.Constraints.Max = image.Pt(900, 600); root.Layout(gtx) })
	click(t, h, "更多 终端")
	click(t, h, "最大化")
	h.Frame()
	if d.Zoomed() != "term" || saved.Zoomed != "term" || !shown(h, "$ go test") || shown(h, "编辑器") || shown(h, "文件树") {
		t.Fatalf("zoom: zoomed=%q saved=%q", d.Zoomed(), saved.Zoomed)
	}
	if b := bounds(h, "$ go test"); b.Dx() < 800 {
		t.Fatalf("zoomed panel should span the 900dp dock: body %v", b)
	}
	h.Key(key.NameEscape, 0)
	h.Frame()
	if d.Zoomed() != "" || !shown(h, "编辑器") || !shown(h, "文件树") {
		t.Fatal("Esc should restore every panel")
	}
	// The zoom is part of the layout, and only for a shown panel.
	d.Zoom("files")
	data, _ := json.Marshal(d.Layout())
	d.Zoom("")
	var l DockLayout
	if err := json.Unmarshal(data, &l); err != nil || !d.SetLayout(l) || d.Zoomed() != "files" {
		t.Fatalf("restore zoom: %v %q", err, d.Zoomed())
	}
	d.SetVisible("files", false)
	if d.Zoomed() != "" {
		t.Fatal("closing the zoomed panel should end the zoom")
	}
	l.Hidden = []string{"files"}
	if !d.SetLayout(l) || d.Zoomed() != "" {
		t.Fatal("a hidden panel cannot be zoomed")
	}
}
