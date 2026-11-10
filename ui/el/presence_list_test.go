package el

import (
	"slices"
	"testing"
	"time"

	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/internal/uitest"
)

func TestPresenceListExitOrderingAndReclamation(t *testing.T) {
	for _, wait := range []bool{false, true} {
		t.Run(map[bool]string{false: "together", true: "ordered"}[wait], func(t *testing.T) {
			now := time.Unix(1800, 0)
			items := []int{1, 2}
			declare := true
			var frames []PresentItem[int]
			root := Root(ViewFunc(func(cx *Context) Element {
				if declare {
					frames = PresentItems(cx, "list", items, func(v int) any { return v }, PresenceListOptions{PresenceOptions: PresenceOptions{Enter: Pose{Fade: 1}, Exit: Pose{Fade: 1}, Animation: Tween(time.Second, Linear)}, ExitBeforeEnter: wait})
				}
				return Div()
			}))
			h := uitest.NewFunc(func(gtx core.C) { gtx.Now = now; root.Layout(gtx) })
			now = now.Add(time.Second)
			h.Frame()
			items = []int{2, 3}
			h.Frame()
			keys := func() []int {
				var out []int
				for _, f := range frames {
					out = append(out, f.Value)
				}
				return out
			}
			want := []int{1, 2, 3}
			if wait {
				want = []int{1, 2}
			}
			if !slices.Equal(keys(), want) || !frames[0].Frame.Exiting {
				t.Fatalf("exit order=%v", keys())
			}
			now = now.Add(time.Second)
			h.Frame()
			if !slices.Equal(keys(), []int{2, 3}) {
				t.Fatalf("after exit=%v", keys())
			}
			if wait && (frames[1].Frame.Pose.Fade != 1 || frames[1].Frame.Settled) {
				t.Fatal("ordered entry did not start at the end of exit")
			}
			items = []int{3, 2}
			h.Frame()
			if !slices.Equal(keys(), items) {
				t.Fatalf("source reorder=%v", keys())
			}
			declare = false
			h.Frame()
			if len(root.presences) != 0 || len(root.presenceLists) != 0 {
				t.Fatal("omitted group retained values")
			}
		})
	}
}

func TestPresenceListInterruptedExitKeepsIdentity(t *testing.T) {
	type item struct {
		id    int
		label string
	}
	now := time.Unix(1900, 0)
	items := []item{{1, "first"}}
	var frames []PresentItem[item]
	root := Root(ViewFunc(func(cx *Context) Element {
		frames = PresentItems(cx, "list", items, func(v item) any { return v.id }, PresenceListOptions{PresenceOptions: PresenceOptions{Exit: Pose{X: 100, Fade: 1}, Animation: Tween(time.Second, Linear)}, ExitBeforeEnter: true})
		return Div()
	}))
	h := uitest.NewFunc(func(gtx core.C) { gtx.Now = now; root.Layout(gtx) })
	items = nil
	h.Frame()
	now = now.Add(500 * time.Millisecond)
	h.Frame()
	if len(frames) != 1 || frames[0].Value.label != "first" || frames[0].Frame.Pose.X != 50 {
		t.Fatalf("retained removed item: %+v", frames)
	}
	items = []item{{1, "updated"}}
	h.Frame()
	if len(frames) != 1 || frames[0].Value.label != "updated" || frames[0].Frame.Pose.X != 50 || frames[0].Frame.Exiting {
		t.Fatalf("reentry: %+v", frames)
	}
	now = now.Add(time.Second)
	h.Frame()
	items = nil
	h.Frame()
	now = now.Add(time.Second)
	h.Frame()
	if len(frames) != 0 || len(root.presences) != 0 || len(root.presenceLists) != 0 {
		t.Fatal("finished collection retained state")
	}
}
