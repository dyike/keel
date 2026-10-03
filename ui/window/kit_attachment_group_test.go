package window

import (
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"testing"
)

func TestAttachmentGroupAgentActions(t *testing.T) {
	opened := 0
	a := kit.Attachment("report.pdf", 1024).OnOpen(func() { opened++ })
	group := kit.AttachmentGroup(a).Name("Files")
	w := openTest(t, Options{Width: 400, Height: 240, Content: el.Root(group)})
	e := element(t, w, "Files")
	if e.Role != "group" {
		t.Fatal("group semantics", e)
	}
	w.click(element(t, w, "report.pdf").center())
	if opened != 1 {
		t.Fatal("agent activation")
	}
	group.SetDisabled(true)
	w.render()
	w.click(element(t, w, "report.pdf").center())
	if opened != 1 {
		t.Fatal("disabled group activation")
	}
	group.SetItems()
	w.render()
	for _, e := range w.snapshot() {
		if e.Name == "report.pdf" {
			t.Fatal("removed item retained")
		}
	}
}
