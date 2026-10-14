package el

import (
	"image"
	"strings"
	"testing"

	"github.com/dyike/keel/gio/op"
	"github.com/dyike/keel/gio/unit"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/internal/uitest"
)

func TestCachedLayoutMatchesFreshAcrossConstraintsAndMetrics(t *testing.T) {
	var cached, fresh Element
	size := float32(14)
	value := strings.Repeat("Cache 中文 words ", 6)
	makeText := func() Element { return Text(value) }
	retained := Root(ViewFunc(func(cx *Context) Element {
		cached = cx.Cache("text", makeText)
		return Div().TextSize(size).Child(cached)
	}))
	rebuilt := Root(ViewFunc(func(*Context) Element {
		fresh = makeText()
		return Div().TextSize(size).Child(fresh)
	}))
	var ops op.Ops
	for _, tc := range []struct {
		width, scale int
		textSize     float32
	}{
		{240, 1, 14}, {240, 1, 14}, {80, 1, 14}, {80, 2, 14}, {240, 2, 20}, {240, 1, 14},
	} {
		size = tc.textSize
		gtx := core.C{Ops: &ops, Metric: unit.Metric{PxPerDp: float32(tc.scale), PxPerSp: float32(tc.scale)}}
		gtx.Constraints.Max = image.Pt(tc.width, 1000)
		ops.Reset()
		retained.Layout(gtx)
		ops.Reset()
		rebuilt.Layout(gtx)
		if cached.node().size != fresh.node().size {
			t.Fatalf("width=%d scale=%d text=%g: cached %v, fresh %v", tc.width, tc.scale, size, cached.node().size, fresh.node().size)
		}
	}
}

func TestKeyHintWithoutActionHandler(t *testing.T) {
	const action = "test.hint-only"
	if err := core.Bind(action, "mod+s"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { core.Bind(action) })
	for _, overlay := range []bool{false, true} {
		chord := ""
		root := Root(ViewFunc(func(cx *Context) Element {
			hint := KeyHint(action, "target", func(s string) Element {
				chord = s
				return Text(s)
			})
			if overlay {
				cx.Overlay("hint", Modal(hint))
				return Div().ID("target")
			}
			return Div().Child(hint, Div().ID("target"))
		}))
		uitest.New(root)
		if chord != "mod+s" {
			t.Fatalf("overlay=%v: hint = %q", overlay, chord)
		}
	}
}

func TestLazyInputEditorPreservesFocusAndText(t *testing.T) {
	value := ""
	var button, field Element
	root := Root(ViewFunc(func(cx *Context) Element {
		button = Div().ID("button").Size(Dp(30)).OnClick(func() { cx.Focus("field") })
		field = Input().ID("field").Bind(&value).W(Dp(200))
		return Div().Child(button, field)
	}))
	h := uitest.New(root)
	buttonState := root.store.states[button.node().key]
	fieldState := root.store.states[field.node().key]
	if buttonState.editor != nil || fieldState.editor == nil {
		t.Fatal("only inputs should allocate an editor")
	}
	editor := fieldState.editor
	h.Click(5, 5)
	h.Frame()
	h.Type("你好 Keel")
	if value != "你好 Keel" || !h.Router.Source().Focused(editor) {
		t.Fatalf("value=%q focused=%v", value, h.Router.Source().Focused(editor))
	}
	h.Frame()
	if root.store.states[field.node().key].editor != editor {
		t.Fatal("editor event tag changed across frames")
	}
}

func TestActionTargetsKeepIndependentAncestry(t *testing.T) {
	var cx *Context
	var got []string
	root := Root(ViewFunc(func(c *Context) Element {
		cx = c
		c.ActionAt("left", "save", func() { got = append(got, "left") })
		c.ActionAt("right", "save", func() { got = append(got, "right") })
		return Div().Child(
			Div().ID("left").KeyContext("Left").Child(Div().ID("left-child").Child(Text("left"))),
			Div().ID("right").KeyContext("Right").Child(Div().ID("right-child").Child(Text("right"))),
		)
	}))
	var ops op.Ops
	gtx := core.C{Ops: &ops, Metric: unit.Metric{PxPerDp: 1, PxPerSp: 1}}
	gtx.Constraints.Max = image.Pt(400, 300)
	root.Layout(gtx)
	if !cx.Perform("left-child", "save") || !cx.Perform("right-child", "save") {
		t.Fatal("ancestor action did not resolve")
	}
	if len(got) != 2 || got[0] != "left" || got[1] != "right" {
		t.Fatalf("actions = %v", got)
	}
	if contexts := cx.bindingTargets["left-child"].contexts; len(contexts) != 1 || contexts[0] != "Left" {
		t.Fatalf("left scopes = %v", contexts)
	}
	if contexts := cx.bindingTargets["right-child"].contexts; len(contexts) != 1 || contexts[0] != "Right" {
		t.Fatalf("right scopes = %v", contexts)
	}
}
