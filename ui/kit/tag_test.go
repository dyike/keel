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

func TestTagColorNames(t *testing.T) {
	for c, name := range map[TagColor]string{TagDefault: "default", TagPrimary: "primary", TagSuccess: "success", TagWarning: "warning", TagDanger: "danger"} {
		h := renderView(Tag("标签").Color(c), 120, 1)
		if _, ok := semanticNode(h, "tag:"+name); !ok {
			t.Fatal(name)
		}
	}
}
