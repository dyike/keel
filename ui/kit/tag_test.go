package kit

import "testing"

func TestTagWrapsAndUpdates(t *testing.T) {
	for _, scale := range []int{1, 2} {
		v := Tag("Release 123 中文长标签需要换行").Tone(Info)
		h := renderView(v, 90, scale)
		n, ok := semanticNode(h, "tag:info")
		if !ok || n.Desc.Bounds.Dx() > 90*scale || n.Desc.Bounds.Dy() < 30*scale {
			t.Fatalf("tag did not wrap: %+v", n)
		}
		v.SetText("完成")
		v.Tone(Success)
		h.Frame()
		if n, ok = semanticNode(h, "tag:success"); !ok || n.Desc.Label != "完成" {
			t.Fatal("tag did not update")
		}
	}
}

func TestTagToneNames(t *testing.T) {
	for c, name := range map[Tone]string{Neutral: "neutral", Info: "info", Success: "success", Warning: "warning", Danger: "danger"} {
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
