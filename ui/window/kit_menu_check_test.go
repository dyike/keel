package window

import (
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"testing"
)

func TestMenuCheckAgentState(t *testing.T) {
	m := kit.Menu().CheckItem("Visible", "", true, nil)
	m.Trigger(kit.Button("Open", m.Toggle))
	w := openTest(t, Options{Width: 400, Height: 300, Content: el.Root(m)})
	w.click(element(t, w, "Open").center())
	e := element(t, w, "Visible")
	if e.Role != "menuitemcheckbox" || e.Checked == nil || !*e.Checked {
		t.Fatal("missing checked state", e)
	}
	w.click(e.center())
	w.click(element(t, w, "Open").center())
	e = element(t, w, "Visible")
	if e.Checked == nil || *e.Checked {
		t.Fatal("stale checked state", e)
	}
}
