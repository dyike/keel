package kit

import (
	"testing"
	"time"

	"gioui.org/io/key"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/locale"
)

func TestTimeFieldSegmentsKeysAndPrecision(t *testing.T) {
	calls := 0
	v := TimeField("Clock").Seconds().Hour12(false).OnChange(func(time.Duration) { calls++ })
	v.SetValue(23*time.Hour + 59*time.Minute + 59*time.Second)
	h := page(v)
	clickClass(t, h, "Editor", "秒 Clock")
	h.Key(key.NameUpArrow, 0)
	if v.Value() != 0 || calls != 1 {
		t.Fatalf("midnight %v callbacks %d", v.Value(), calls)
	}
	clickClass(t, h, "Editor", "时 Clock")
	h.Key(key.NameUpArrow, 0)
	clickClass(t, h, "Editor", "分 Clock")
	h.Key(key.NamePageUp, 0)
	if v.Value() != time.Hour+10*time.Minute || calls != 3 {
		t.Fatalf("segment steps %v %d", v.Value(), calls)
	}
	v.SetValue(-time.Second)
	if v.Value() != 24*time.Hour-time.Second || calls != 3 {
		t.Fatal("program value precision or callback")
	}
}

func TestTimeFieldSegmentDraftBlurAndDisable(t *testing.T) {
	for _, mode := range []string{"blur", "disabled", "ancestor"} {
		t.Run(mode, func(t *testing.T) {
			v := TimeField("Clock").Segmented().Hour12(false)
			calls := 0
			v.OnChange(func(time.Duration) { calls++ })
			disabled := false
			h := render(func(cx *el.Context) el.Element {
				return el.Div().Child(el.Div().Disabled(disabled).Child(v.Render(cx)), Button("Other", func() {}).Render(cx))
			})
			clickClass(t, h, "Editor", "时 Clock")
			h.Key(key.NameEnd, 0)
			h.Key(key.NameDeleteBackward, 0)
			h.Key(key.NameDeleteBackward, 0)
			h.Type("09")
			switch mode {
			case "blur":
				click(t, h, "Other")
			case "disabled":
				v.SetDisabled(true)
			case "ancestor":
				disabled = true
			}
			h.Frame()
			h.Frame()
			want := time.Duration(0)
			n := 0
			if mode == "blur" {
				want = 9 * time.Hour
				n = 1
			}
			if v.Value() != want || calls != n {
				t.Fatalf("value %v calls %d", v.Value(), calls)
			}
			if mode != "blur" && v.parts[0] != "00" {
				t.Fatal("disabled retained draft")
			}
		})
	}
}

func TestTimeFieldLocalePeriodAndInvalidInput(t *testing.T) {
	old := locale.Current()
	defer locale.Apply(old)
	locale.Apply(locale.English())
	v := TimeField("Clock").Seconds()
	v.SetValue(12*time.Hour + 30*time.Minute + 5*time.Second)
	h := page(v)
	if v.parts[0] != "12" || !shown(h, "PM") {
		t.Fatal("noon locale")
	}
	click(t, h, "Period Clock")
	if v.Value() != 30*time.Minute+5*time.Second || v.parts[0] != "12" {
		t.Fatal("midnight period")
	}
	clickClass(t, h, "Editor", "Minute Clock")
	h.Key(key.NameEnd, 0)
	h.Key(key.NameDeleteBackward, 0)
	h.Key(key.NameDeleteBackward, 0)
	h.Type("99")
	h.Key(key.NameReturn, 0)
	if v.parts[1] != "30" {
		t.Fatal("invalid minute did not revert")
	}
	h.Key(key.NameEnd, 0)
	h.Key(key.NameDeleteBackward, 0)
	h.Key(key.NameDeleteBackward, 0)
	h.Type("45")
	h.Router.MoveFocus(key.FocusForward)
	h.Frame()
	h.Key(key.NameUpArrow, 0)
	if v.Value() != 45*time.Minute+6*time.Second {
		t.Fatalf("tab to seconds %v", v.Value())
	}
	locale.Apply(locale.Chinese())
	h.Frame()
	if v.parts[0] != "00" || shown(h, "AM") || !shown(h, "秒 Clock") {
		t.Fatal("runtime locale switch")
	}
}

func TestTimeFieldPeriodCommitsDraftOnce(t *testing.T) {
	calls := 0
	v := TimeField("Clock").Hour12(true).OnChange(func(time.Duration) { calls++ })
	h := page(v)
	clickClass(t, h, "Editor", "分 Clock")
	h.Key(key.NameEnd, 0)
	h.Key(key.NameDeleteBackward, 0)
	h.Key(key.NameDeleteBackward, 0)
	h.Type("45")
	click(t, h, "时段 Clock")
	h.Frame()
	h.Frame()
	if v.Value() != 12*time.Hour+45*time.Minute || calls != 1 {
		t.Fatalf("period draft %v callbacks %d", v.Value(), calls)
	}
}
