package kit

import (
	"github.com/dyike/keel/third_party/gio/io/key"
	"github.com/dyike/keel/third_party/gio/io/transfer"
	"github.com/dyike/keel/ui/el"
	"io"
	"strings"
	"testing"
	"time"
)

func TestTimeFieldQuickSegmentsAdvanceAndCycle(t *testing.T) {
	calls := 0
	v := TimeField("Clock").Seconds().Hour12(false).SegmentKeys(true).OnChange(func(time.Duration) { calls++ })
	var cx *el.Context
	h := render(func(c *el.Context) el.Element { cx = c; return v.Render(c) })
	clickClass(t, h, "Editor", "时 Clock")
	h.Router.Queue(key.EditEvent{Range: key.Range{Start: 0, End: 2}, Text: "09"})
	h.Frame()
	h.Frame()
	h.Frame()
	if v.Value() != 9*time.Hour || calls != 1 || !cx.Focused(v.segmentID(1)) {
		t.Fatal("hour did not commit and advance", v.Value(), calls)
	}
	h.Router.Queue(key.EditEvent{Range: key.Range{Start: 0, End: 2}, Text: "59"})
	h.Frame()
	h.Frame()
	h.Frame()
	if v.Value() != 9*time.Hour+59*time.Minute || !cx.Focused(v.segmentID(2)) {
		t.Fatal("minute did not advance", v.Value())
	}
	h.Key(key.NameLeftArrow, 0)
	h.Frame()
	h.Key(key.NameUpArrow, 0)
	if v.Value() != 9*time.Hour {
		t.Fatal("minute carried into hour", v.Value())
	}
	h.Key(key.NameDownArrow, 0)
	if v.Value() != 9*time.Hour+59*time.Minute {
		t.Fatal("minute did not wrap backward", v.Value())
	}
	h.Key(key.NameDeleteBackward, 0)
	if v.Value() != 9*time.Hour {
		t.Fatal("delete did not reset minute")
	}
}

func TestTimeFieldQuickPeriodInvalidAndDisabled(t *testing.T) {
	v := TimeField("Clock").Hour12(true).SegmentKeys(true)
	v.SetValue(13*time.Hour + 30*time.Minute)
	disabled := false
	h := render(func(cx *el.Context) el.Element { return el.Div().Disabled(disabled).Child(v.Render(cx)) })
	clickClass(t, h, "Editor", "时 Clock")
	h.Router.Queue(key.EditEvent{Range: key.Range{Start: 0, End: 2}, Text: "99"})
	h.Frame()
	h.Frame()
	if v.Value() != 13*time.Hour+30*time.Minute {
		t.Fatal("invalid hour committed")
	}
	h.Key(key.NameDeleteBackward, 0)
	if v.Value() != 12*time.Hour+30*time.Minute {
		t.Fatal("12-hour reset lost period", v.Value())
	}
	h.Key("A", 0)
	if v.Value() != 30*time.Minute {
		t.Fatal("AM shortcut", v.Value())
	}
	h.Key("P", 0)
	if v.Value() != 12*time.Hour+30*time.Minute {
		t.Fatal("PM shortcut", v.Value())
	}
	disabled = true
	h.Frame()
	h.Key(key.NameUpArrow, 0)
	if v.Value() != 12*time.Hour+30*time.Minute {
		t.Fatal("disabled quick segment changed")
	}
}

func TestTimeFieldSizesPreserveValue(t *testing.T) {
	for _, segmented := range []bool{false, true} {
		for _, scale := range []int{1, 2} {
			v := TimeField("")
			if segmented {
				v.Seconds().Hour12(false)
			}
			v.SetValue(9*time.Hour + 30*time.Minute)
			var cx *el.Context
			h := renderView(el.ViewFunc(func(c *el.Context) el.Element { cx = c; return v.Render(c) }), 400, scale)
			previous := float32(0)
			for _, size := range []TimeFieldSize{TimeFieldSizeXSmall, TimeFieldSizeSmall, TimeFieldSizeMedium, TimeFieldSizeLarge} {
				v.Size(size)
				h.Frame()
				_, height := cx.LastSize(autoID("time", v))
				if height <= previous {
					t.Fatal("size heights not increasing", segmented, size, height, previous)
				}
				previous = height
				if v.Value() != 9*time.Hour+30*time.Minute {
					t.Fatal("size changed time")
				}
			}
		}
	}
}

func TestTimeFieldQuickPasteAdvancesOnce(t *testing.T) {
	calls := 0
	v := TimeField("Clock").SegmentKeys(true).Hour12(false).OnChange(func(time.Duration) { calls++ })
	var cx *el.Context
	h := render(func(c *el.Context) el.Element { cx = c; return v.Render(c) })
	clickClass(t, h, "Editor", "时 Clock")
	h.Frame()
	h.Key("V", key.ModShortcut)
	h.Router.Queue(transfer.DataEvent{Type: "application/text", Open: func() io.ReadCloser { return io.NopCloser(strings.NewReader("17")) }})
	h.Frame()
	h.Frame()
	if v.Value() != 17*time.Hour || calls != 1 || !cx.Focused(v.segmentID(1)) {
		t.Fatal("paste failed to commit and advance", v.Value(), calls)
	}
}
