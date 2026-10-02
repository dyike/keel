package kit

import "testing"

func TestTextAreaClearsStaleErrorOnUserEdit(t *testing.T) {
	for _, scale := range []int{1, 2} {
		calls := 0
		v := TextArea("备注").OnChange(func(string) { calls++ })
		v.SetError("请补充说明")
		h := renderView(v, 240, scale)
		if !shown(h, "请补充说明") {
			t.Fatal("missing validation error")
		}
		clickClass(t, h, "Editor", "备注")
		h.Type("已补充\n第二行 123")
		h.Frame()
		if v.Error() != "" || shown(h, "请补充说明") || calls != 1 {
			t.Fatal("stale error after user correction", v.Error(), calls)
		}
		v.SetError("服务器拒绝")
		v.SetDisabled(true)
		h.Frame()
		h.Type("ignored")
		h.Frame()
		if v.Error() != "服务器拒绝" || calls != 1 {
			t.Fatal("disabled editing cleared error")
		}
	}
}
