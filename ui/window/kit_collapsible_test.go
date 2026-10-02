package window

import (
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/kit"
	"github.com/dyike/keel/ui/theme"
	"testing"
)

func TestKitCollapsibleAgentToggle(t *testing.T) {
	old := theme.ReducedMotion
	core.Update(func() { theme.SetReducedMotion(true) })
	defer core.Update(func() { theme.SetReducedMotion(old) })
	v := kit.Collapsible("Details", kit.Input("saved"))
	w := openTest(t, kitPage(v))
	if roleOfName(w, "Details") != "disclosure" {
		t.Fatal("disclosure semantics")
	}
	w.click(element(t, w, "Details").center())
	if !v.Value() || element(t, w, "Details").Value != "expanded" {
		t.Fatal("agent toggle")
	}
	v.SetDisabled(true)
	w.render()
	w.click(element(t, w, "Details").center())
	if !v.Value() {
		t.Fatal("disabled disclosure toggled")
	}
}
