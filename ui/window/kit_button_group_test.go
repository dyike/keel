package window

import (
	"testing"

	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
)

func TestKitButtonGroupAndRadioSnapshot(t *testing.T) {
	group := kit.ButtonGroup(kit.Button("列表", nil).Selected(true), kit.Button("看板", nil)).Name("视图")
	radio := kit.Radio("快递配送").Checked(true)
	w := openTest(t, Options{Width: 320, Height: 200, Content: el.Root(el.ViewFunc(func(cx *el.Context) el.Element {
		return el.Div().P(16).Gap(12).Items(el.Start).Child(group.Render(cx), radio.Render(cx))
	}))})
	if e := element(t, w, "视图"); e.Role != "group" {
		t.Fatalf("group: %+v", e)
	}
	if e := element(t, w, "列表"); e.Role != "button" || e.Selected == nil || !*e.Selected {
		t.Fatalf("selected button: %+v", e)
	}
	if e := element(t, w, "快递配送"); e.Role != "radio" {
		t.Fatalf("radio: %+v", e)
	}
}
