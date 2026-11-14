package kit

import (
	"github.com/dyike/keel/gio/f32"
	"github.com/dyike/keel/gio/io/pointer"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/internal/uitest"
	"image"
	"image/color"
	"io"
	"strings"
	"testing"

	"github.com/dyike/keel/gio/io/key"
	"github.com/dyike/keel/gio/io/transfer"
	"github.com/dyike/keel/ui/el"
)

func TestTerminalInputOutputResizeAndNoEcho(t *testing.T) {
	v := Terminal().Name("Session").Height(180)
	var input string
	resizes := 0
	v.OnData(func(s string) { input += s }).OnResize(func(cols, rows int) {
		resizes++
		if cols < 1 || rows < 1 {
			t.Fatal(cols, rows)
		}
	})
	h := renderView(v, 400, 1)
	v.Focus()
	h.Frame()
	typeAt(h, "hello")
	h.Frame()
	if input != "hello" || strings.Contains(v.Text(), "hello") {
		t.Fatal("unexpected echo", input, v.Text())
	}
	h.Key(key.NameReturn, 0)
	h.Key("C", key.ModCtrl)
	h.Key(key.NameUpArrow, 0)
	if input != "hello\r\x03\x1b[A" {
		t.Fatal([]byte(input))
	}
	v.WriteString("\x1b[?1h")
	h.Key(key.NameUpArrow, 0)
	if !strings.HasSuffix(input, "\x1bOA") {
		t.Fatal("application cursor")
	}
	for range 4 {
		h.Frame()
	}
	if resizes != 1 {
		t.Fatal("unchanged size notified", resizes)
	}
	v.WriteString("\x1b[31moutput\x1b[0m")
	h.Frame()
	if !strings.Contains(v.Text(), "output") {
		t.Fatal(v.Text())
	}
	v.SetDisabled(true)
	h.Frame()
	typeAt(h, "blocked")
	h.Key(key.NameReturn, 0)
	if strings.Contains(input, "blocked") {
		t.Fatal("disabled input")
	}
	v.WriteString(" accepted")
	if !strings.Contains(v.Text(), "accepted") {
		t.Fatal("disabled output blocked")
	}
}
func TestTerminalSelectionCopyPreservesWideText(t *testing.T) {
	v := Terminal().Height(180)
	h := renderView(v, 400, 1)
	v.WriteString("a中bc")
	h.Frame()
	cw, lh := v.painter.metrics.space, v.painter.metrics.lh
	h.Drag(0, float32(lh)/2, float32(5*cw), float32(lh)/2)
	if v.Selection() != "a中bc" {
		t.Fatal(v.Selection())
	}
	h.Key("C", key.ModCtrl|key.ModShift)
	_, data, ok := h.Router.WriteClipboard()
	if !ok || string(data) != "a中bc" {
		t.Fatal(string(data), ok)
	}
	var raw string
	v.OnData(func(s string) { raw += s })
	h.Key("C", key.ModCtrl)
	if raw != "\x03" {
		t.Fatal("Ctrl+C intercepted", raw)
	}
}
func TestTerminalCompositionCommitsOnce(t *testing.T) {
	v := Terminal()
	var input string
	v.OnData(func(s string) { input += s })
	h := renderView(v, 400, 1)
	v.Focus()
	h.Frame()
	h.Router.Queue(key.CompositionEvent{Start: 0, End: 2}, key.EditEvent{Text: "ni"})
	h.Frame()
	if input != "" || string(v.draft) != "ni" {
		t.Fatal("preedit leaked", input, string(v.draft))
	}
	h.Router.Queue(key.EditEvent{Range: key.Range{Start: 0, End: 2}, Text: "你"}, key.CompositionEvent{Start: -1, End: -1})
	h.Frame()
	if input != "你" || len(v.draft) != 0 {
		t.Fatal("commit", input, string(v.draft))
	}
	h.Frame()
	if input != "你" {
		t.Fatal("duplicate commit")
	}
}
func TestTerminalBracketedPasteAndAncestorDisabled(t *testing.T) {
	v := Terminal()
	var input string
	v.OnData(func(s string) { input += s })
	disabled := false
	h := renderView(viewFunc(func(cx *el.Context) el.Element { return el.Div().Disabled(disabled).Child(v.Render(cx)) }), 400, 1)
	v.Focus()
	h.Frame()
	v.WriteString("\x1b[?2004h")
	h.Key("V", key.ModCtrl|key.ModShift)
	h.Router.Queue(transfer.DataEvent{Type: "application/text", Open: func() io.ReadCloser { return io.NopCloser(strings.NewReader("a\nb")) }})
	h.Frame()
	if input != "\x1b[200~a\nb\x1b[201~" {
		t.Fatal([]byte(input))
	}
	disabled = true
	h.Frame()
	typeAt(h, "blocked")
	h.Key(key.NameReturn, 0)
	if strings.Contains(input, "blocked") {
		t.Fatal("ancestor input")
	}
}
func TestTerminalPaletteColors(t *testing.T) {
	fallback := color.NRGBA{1, 2, 3, 255}
	cases := []struct {
		code uint32
		want color.NRGBA
	}{
		{0, fallback}, {2, color.NRGBA{205, 0, 0, 255}}, {13, color.NRGBA{92, 92, 255, 255}},
		{17, color.NRGBA{0, 0, 0, 255}}, {232, color.NRGBA{255, 255, 255, 255}}, {233, color.NRGBA{8, 8, 8, 255}},
		{1<<24 | 0x123456, color.NRGBA{0x12, 0x34, 0x56, 255}},
	}
	for _, tc := range cases {
		if got := terminalColor(tc.code, fallback); got != tc.want {
			t.Fatal(tc.code, got, tc.want)
		}
	}
}

func TestTerminalTouchTapFocusAndSelectionSurvivesOutput(t *testing.T) {
	v := Terminal().Height(180)
	input := ""
	v.OnData(func(s string) { input += s })
	h := renderView(v, 400, 1)
	for _, kind := range []pointer.Kind{pointer.Press, pointer.Release} {
		h.Router.Queue(pointer.Event{Kind: kind, Source: pointer.Touch, PointerID: 1, Position: f32.Pt(20, 20)})
		h.Frame()
	}
	typeAt(h, "typed")
	if input != "typed" {
		t.Fatal("touch did not focus", input)
	}
	v.WriteString("first\r\n")
	h.Frame()
	cw, lh := v.painter.metrics.space, v.painter.metrics.lh
	h.Drag(0, float32(lh)/2, float32(cw*5), float32(lh)/2)
	if v.Selection() != "first" {
		t.Fatal(v.Selection())
	}
	v.WriteString("unrelated")
	if v.Selection() != "first" {
		t.Fatal("output cleared unchanged selection", v.Selection())
	}
	v.WriteString("\x1b[Hchanged")
	if v.Selection() != "" {
		t.Fatal("changed source retained stale selection")
	}
}
func TestTerminalMeasurementDoesNotResizeOrNotify(t *testing.T) {
	v := Terminal()
	v.WriteString("preserved")
	calls := 0
	v.OnResize(func(int, int) { calls++ })
	uitest.NewFunc(func(gtx core.C) {
		gtx.Constraints.Max = image.Pt(5, 5)
		v.layoutViewport(gtx, image.Rectangle{})
	})
	cols, rows := v.Size()
	if cols != 80 || rows != 24 || calls != 0 || !strings.HasPrefix(v.Text(), "preserved") {
		t.Fatal(cols, rows, calls, v.Text())
	}
}
