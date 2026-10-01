package widget

import (
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/internal/uitest"
	"github.com/dyike/keel/ui/layout"
	"testing"
)

func TestAssociatedLabelFocus(t *testing.T) {
	f := Input("")
	label := Text("外部标签").For(f)
	h := uitest.NewFunc(func(gtx core.C) {
		layout.Column(label, f).Layout(gtx)
	})
	clickNamed(t, h, "外部标签")
	h.Frame()
	h.Type("关联成功")
	if f.Value() != "关联成功" {
		t.Fatal("associated label did not focus field")
	}
}
