package el

import (
	"github.com/dyike/keel/third_party/gio/io/key"
	"testing"

	"github.com/dyike/keel/ui/internal/uitest"
)

func TestDragReportsPositionInDp(t *testing.T) {
	var got []DragEvent
	disabled := false
	h := uitest.New(Embed(ViewFunc(func(cx *Context) Element {
		return Div().P(10).Child(Div().ID("track").Name("track").W(Dp(200)).H(Dp(20)).Disabled(disabled).
			OnDrag(func(e DragEvent) { got = append(got, e) }))
	})))
	h.Drag(60, 20, 160, 25)
	if len(got) < 3 || got[0].Kind != DragStart || got[len(got)-1].Kind != DragEnd {
		t.Fatalf("events %+v", got)
	}
	if got[0].X != 50 || got[0].W != 200 || got[0].H != 20 {
		t.Fatalf("start %+v", got[0])
	}
	if last := got[len(got)-1]; last.X != 150 {
		t.Fatalf("end %+v", last)
	}
	got, disabled = nil, true
	h.Frame()
	h.Drag(60, 20, 160, 25)
	if len(got) != 0 {
		t.Fatal("disabled element received a drag")
	}
}

func TestDisabledSubtreeDisablesInput(t *testing.T) {
	text := ""
	h := uitest.New(Embed(ViewFunc(func(cx *Context) Element {
		return Div().Disabled(true).Child(Input().ID("in").Name("in").W(Dp(200)).Bind(&text))
	})))
	h.Click(center(nodeBounds(h, "in")))
	h.Type("abc")
	if text != "" {
		t.Fatalf("disabled input accepted %q", text)
	}
}

func TestInputMaxLenFilterReadOnly(t *testing.T) {
	a, b := "", "keep"
	h := uitest.New(Embed(ViewFunc(func(cx *Context) Element {
		return Div().Gap(8).Child(
			Input().ID("a").Name("a").W(Dp(200)).MaxLen(3).Filter("0123456789").Bind(&a),
			Input().ID("b").Name("b").W(Dp(200)).ReadOnly(true).Bind(&b),
		)
	})))
	h.Click(center(nodeBounds(h, "a")))
	h.Type("1a2b34")
	if a != "123" {
		t.Fatalf("filtered input %q", a)
	}
	h.Click(center(nodeBounds(h, "b")))
	h.Type("x")
	if b != "keep" {
		t.Fatalf("read-only input %q", b)
	}
}

func TestSingleLineInputOnKeyTakesArrows(t *testing.T) {
	var keys []string
	text := "ab"
	h := uitest.New(Embed(ViewFunc(func(cx *Context) Element {
		return Div().Child(Input().ID("in").Name("in").W(Dp(200)).Bind(&text).OnKey(func(e KeyEvent) bool {
			if e.State == KeyPress {
				keys = append(keys, e.Name)
			}
			return true
		}))
	})))
	h.Click(center(nodeBounds(h, "in")))
	for _, k := range []key.Name{key.NameUpArrow, key.NameDownArrow, key.NamePageDown} {
		h.Key(k, 0)
	}
	h.Type("c") // other keys still edit
	if len(keys) != 3 || keys[0] != string(key.NameUpArrow) {
		t.Fatalf("keys %v", keys)
	}
	if text != "abc" && text != "cab" && text != "acb" {
		t.Fatalf("typing broken: %q", text)
	}
}
