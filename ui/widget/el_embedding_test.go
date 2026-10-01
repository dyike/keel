package widget

import (
	"testing"

	"gioui.org/io/key"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/internal/uitest"
)

type embeddedWidgetView struct{ widget core.Widget }

func (v embeddedWidgetView) Render(*el.Context) el.Element {
	return el.Div().Child(el.Widget(v.widget))
}

func TestSelectStaysOpenDuringElMeasurement(t *testing.T) {
	s := Select("状态", "待处理", "已完成")
	h := uitest.New(el.Root(embeddedWidgetView{s}))
	clickNamed(t, h, "状态")
	for frame := 0; frame < 3; frame++ {
		h.Frame()
		if !s.open {
			t.Fatalf("measurement closed select on frame %d", frame)
		}
		if bounds := namedBounds(h, "已完成"); bounds.Empty() {
			t.Fatalf("option disappeared on frame %d", frame)
		}
	}
	clickNamed(t, h, "已完成")
	if s.Value() != "已完成" {
		t.Fatal("visible option could not be selected")
	}
}

func TestTableKeepsFocusDuringElMeasurement(t *testing.T) {
	tb := Table(Col("名称", 1))
	tb.SetRows([][]string{{"甲"}, {"乙"}})
	h := uitest.New(el.Root(embeddedWidgetView{tb}))
	clickNamed(t, h, "甲")
	for frame := 0; frame < 3; frame++ {
		h.Frame()
		if !tb.focused || !h.Router.Source().Focused(tb) {
			t.Fatalf("measurement lost table focus on frame %d", frame)
		}
	}
	h.Key(key.NameDownArrow, 0)
	if tb.Selected() != 1 {
		t.Fatal("focused table did not accept keyboard navigation")
	}
	tb.SetDisabled(true)
	if tb.focused {
		t.Fatal("disabling table did not clear saved focus")
	}
}
