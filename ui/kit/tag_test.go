package kit

import (
	"github.com/dyike/keel/third_party/gio/io/key"
	"github.com/dyike/keel/ui/el"
	"math"
	"testing"
)

func TestTagWrapsAndUpdates(t *testing.T) {
	for _, scale := range []int{1, 2} {
		v := Tag("Release 123 中文长标签需要换行").Tone(ToneInfo)
		h := renderView(v, 90, scale)
		n, ok := semanticNode(h, "tag:info")
		if !ok || n.Desc.Bounds.Dx() > 90*scale || n.Desc.Bounds.Dy() < 30*scale {
			t.Fatalf("tag did not wrap: %+v", n)
		}
		v.SetText("完成")
		v.Tone(ToneSuccess)
		h.Frame()
		if n, ok = semanticNode(h, "tag:success"); !ok || n.Desc.Label != "完成" {
			t.Fatal("tag did not update")
		}
	}
}

func TestTagToneNames(t *testing.T) {
	for c, name := range map[Tone]string{ToneNeutral: "neutral", ToneInfo: "info", ToneSuccess: "success", ToneWarning: "warning", ToneDanger: "danger"} {
		h := renderView(Tag("标签").Tone(c), 120, 1)
		if _, ok := semanticNode(h, "tag:"+name); !ok {
			t.Fatal(name)
		}
	}
}

func TestTagSelectionRemovalAndDisabled(t *testing.T) {
	changes, removes := 0, 0
	v := Tag("标签").Selectable().OnChange(func(bool) { changes++ }).OnRemove(func() { removes++ })
	h := renderView(v, 180, 1)
	v.SetValue(true)
	h.Frame()
	if changes != 0 {
		t.Fatal("program change fired callback")
	}
	click(t, h, "标签")
	if v.Value() || changes != 1 {
		t.Fatal("selection failed")
	}
	click(t, h, "移除 标签")
	if removes != 1 || changes != 1 {
		t.Fatal("remove also selected")
	}
	h.Frame()
	if _, ok := node(h, "移除 标签"); !ok {
		t.Fatal("tag hid itself")
	}
	v.SetDisabled(true)
	h.Frame()
	click(t, h, "移除 标签")
	if removes != 1 {
		t.Fatal("disabled remove")
	}
	v.SetDisabled(false)
	h.Frame()
	click(t, h, "移除 标签")
	if removes != 2 {
		t.Fatal("restore failed")
	}
}

func TestTagRichContentSizeAndInheritedDisable(t *testing.T) {
	for _, scale := range []int{1, 2} {
		changes, removes := 0, 0
		disabled := false
		v := Tag("Rich").Outline(true).Rounded(0).Size(32).Content(text("Custom label")).Selectable().OnChange(func(bool) { changes++ }).OnRemove(func() { removes++ })
		h := renderView(viewFunc(func(cx *el.Context) el.Element { return el.Div().Disabled(disabled).Child(v.Render(cx)) }), 110, scale)
		n, ok := semanticNode(h, "tag:neutral")
		if !ok || n.Desc.Bounds.Dx() > 110*scale || n.Desc.Bounds.Dy() < 32*scale || !shown(h, "Custom label") {
			t.Fatal("rich layout", n)
		}
		click(t, h, "切换 Rich")
		h.Key(key.NameSpace, 0)
		if changes != 2 || v.Value() {
			t.Fatal("rich keyboard selection")
		}
		click(t, h, "移除 Rich")
		if removes != 1 || changes != 2 {
			t.Fatal("rich removal selected")
		}
		disabled = true
		h.Frame()
		click(t, h, "切换 Rich")
		click(t, h, "移除 Rich")
		if changes != 2 || removes != 1 {
			t.Fatal("inherited disable")
		}
		disabled = false
		v.SetValue(true)
		v.Content(nil)
		h.Frame()
		if shown(h, "Custom label") || changes != 2 {
			t.Fatal("reset or setter callback")
		}
		v.Size(float32(math.NaN())).Size(-1).Rounded(-1).Rounded(float32(math.Inf(1)))
		if v.height != 32 || *v.radius != 0 {
			t.Fatal("invalid style accepted")
		}
		v.Size(20)
		h.Frame()
		n, _ = semanticNode(h, "tag:neutral")
		if n.Desc.Bounds.Dy() >= 32*scale {
			t.Fatal("small size did not shrink", n.Desc.Bounds)
		}
	}
}
