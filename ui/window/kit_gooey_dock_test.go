package window

import (
	"github.com/dyike/keel/gio/f32"
	"github.com/dyike/keel/gio/io/pointer"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"os"
	"path/filepath"
	"testing"
	"time"
)

type gooeyTestRoot struct{ core.Widget }

func (gooeyTestRoot) FillsWindow() bool { return true }

func TestGooeyDockSnapshot(t *testing.T) {
	now := time.Unix(100, 0)
	names := []string{"Home", "Messages", "Search", "Music", "Notifications", "Profile", "Settings"}
	files := []string{"house", "message-circle", "search", "music", "bell", "user", "settings"}
	items := make([]kit.GooeyDockItem, len(names))
	calls := 0
	for i, name := range names {
		icon, err := kit.SVGIconFile("../../examples/components/assets/lucide/" + files[i] + ".svg")
		if err != nil {
			t.Fatal(err)
		}
		items[i] = kit.GooeyDockItem{ID: files[i], Label: name, IconView: icon, Active: i == 0, Action: func() { calls++ }}
	}
	dock := kit.GooeyDock(items...)
	root := el.Root(el.ViewFunc(func(cx *el.Context) el.Element {
		return el.Div().Child(el.Div().WFull().H(el.Dp(200)).Justify(el.End).Items(el.Center).Pb(32).Child(dock.Render(cx)))
	}))
	w := openTest(t, Options{Width: 520, Height: 240, Content: gooeyTestRoot{core.Func(func(gtx core.C) core.D { gtx.Now = now; return root.Layout(gtx) })}})
	save := func(name string) {
		t.Helper()
		b, err := w.screenshot()
		if err != nil {
			t.Fatal(err)
		}
		if dir := os.Getenv("KEEL_GOOEY_CAPTURE"); dir != "" {
			if err = os.WriteFile(filepath.Join(dir, name+".png"), b, 0600); err != nil {
				t.Fatal(err)
			}
		}
	}
	save("gooey-rest")
	home := element(t, w, "Home")
	if home.Role != "button" {
		t.Fatalf("home role: %s", home.Role)
	}
	music := element(t, w, "Music")
	w.virt.router.Queue(pointer.Event{Kind: pointer.Move, Source: pointer.Mouse, Position: f32.Pt(float32(music.X+music.Width/2), float32(music.Y+music.Height/2))})
	w.snapshot()
	for range 120 {
		now = now.Add(time.Second / 60)
		w.snapshot()
	}
	save("gooey-hover")
	music = element(t, w, "Music")
	if music.Width <= 40 {
		t.Fatalf("not enlarged: %+v", music)
	}
	w.click(music.center())
	if calls != 1 {
		t.Fatalf("clicks: %d", calls)
	}
}
