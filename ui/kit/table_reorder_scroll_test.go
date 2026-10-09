package kit

import (
	"fmt"
	"image"
	"slices"
	"testing"
	"time"

	"github.com/dyike/keel/third_party/gio/f32"
	"github.com/dyike/keel/third_party/gio/io/pointer"
	"github.com/dyike/keel/third_party/gio/unit"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/internal/uitest"
)

func TestTableHeaderStationaryEdgeScroll(t *testing.T) {
	for _, scale := range []int{1, 2} {
		for _, frozen := range []bool{false, true} {
			for _, ending := range []string{"release", "cancel", "outside", "disabled"} {
				t.Run(fmt.Sprint(scale, frozen, ending), func(t *testing.T) {
					cols := make([]*ColumnSpec, 8)
					for i := range cols {
						cols[i] = Col(fmt.Sprint(i)).Width(100)
					}
					v := Table(cols...).Height(80)
					if frozen {
						v.FrozenColumns(1, 1)
					}
					now := time.Unix(1000, 0)
					var cx *el.Context
					root := el.Embed(viewFunc(func(c *el.Context) el.Element { cx = c; return v.Render(c) }))
					h := uitest.NewFunc(func(gtx core.C) {
						gtx.Now = now
						gtx.Metric = unit.Metric{PxPerDp: float32(scale), PxPerSp: float32(scale)}
						gtx.Constraints.Max = image.Pt(400*scale, 1000*scale)
						root.Layout(gtx)
					})
					frames := func(n int) {
						for i := 0; i < n; i++ {
							now = now.Add(20 * time.Millisecond)
							h.Frame()
						}
					}
					source := 0
					edge := float32(392)
					left := float32(8)
					if frozen {
						source = 1
						edge = 292
						left = 108
					}
					start := f32.Pt(float32(source*100+30)*float32(scale), 15*float32(scale))
					end := f32.Pt(edge*float32(scale), start.Y)
					h.Router.Queue(pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, Buttons: pointer.ButtonPrimary, Position: start})
					h.Frame()
					h.Router.Queue(pointer.Event{Kind: pointer.Move, Source: pointer.Mouse, Buttons: pointer.ButtonPrimary, Position: end})
					h.Frame()
					frames(200)
					offset, view, content := cx.ScrollStateX(autoID("table", v))
					if offset != content-view {
						t.Fatalf("stationary edge did not reach end: %v/%v", offset, content-view)
					}
					if v.columnDrag == nil || !v.columnDrag.valid {
						t.Fatal("missing drop after scrolling")
					}
					// Move back to the left edge without releasing: reverse while source is offscreen.
					end.X = left * float32(scale)
					h.Router.Queue(pointer.Event{Kind: pointer.Move, Source: pointer.Mouse, Buttons: pointer.ButtonPrimary, Position: end})
					h.Frame()
					frames(200)
					if x, _, _ := cx.ScrollStateX(autoID("table", v)); x != 0 {
						t.Fatal("left scroll failed", x, v.columnDrag, v.headerGeometry[source])
					}
					end.X = edge * float32(scale)
					h.Router.Queue(pointer.Event{Kind: pointer.Move, Source: pointer.Mouse, Buttons: pointer.ButtonPrimary, Position: end})
					h.Frame()
					frames(200)
					calls := 0
					v.OnColumnMove(func(int, int, int) { calls++ })
					kind := pointer.Release
					switch ending {
					case "cancel":
						kind = pointer.Cancel
					case "outside":
						end.Y += 100 * float32(scale)
					case "disabled":
						v.SetDisabled(true)
					}
					h.Router.Queue(pointer.Event{Kind: kind, Source: pointer.Mouse, Position: end})
					h.Frame()
					if ending == "release" && (slices.Index(v.columns, source) <= source || calls != 1) {
						t.Fatal("drop did not move to revealed column", v.columns, calls)
					}
					if ending != "release" && (slices.Index(v.columns, source) != source || calls != 0) {
						t.Fatal("canceled drag committed", v.columns, calls)
					}
					x, _, _ := cx.ScrollStateX(autoID("table", v))
					frames(10)
					if after, _, _ := cx.ScrollStateX(autoID("table", v)); after != x || v.columnDrag != nil {
						t.Fatal("scroll survived release")
					}
				})
			}
		}
	}
}
