package window

import (
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"github.com/dyike/keel/ui/theme"
	"testing"
)

func TestKitAccordionSizeAndBorderSnapshot(t *testing.T) {
	old := theme.ReducedMotion
	core.Update(func() { theme.SetReducedMotion(true) })
	defer core.Update(func() { theme.SetReducedMotion(old) })
	body := el.ViewFunc(func(*el.Context) el.Element { return el.Text("Details") })
	v := kit.Accordion().Add("Section", body).Size(kit.AccordionSizeSmall).Bordered(false)
	w := openTest(t, kitPage(v))
	small := element(t, w, "Section")
	if small.Role != "disclosure" || small.Value != "collapsed" {
		t.Fatalf("semantics %+v", small)
	}
	v.Size(kit.AccordionSizeLarge).Bordered(true)
	large := element(t, w, "Section")
	if large.Height <= small.Height {
		t.Fatal("size not reflected")
	}
	w.click(large.center())
	if e := element(t, w, "Section"); e.Value != "expanded" {
		t.Fatal("not expanded")
	}
	if roleOfName(w, "Details") != "text" {
		t.Fatal("body missing")
	}
}
