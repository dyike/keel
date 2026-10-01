package widget

import (
	"gioui.org/io/key"
	"github.com/dyike/keel/ui/core"
	"testing"

	"github.com/dyike/keel/ui/internal/uitest"
)

func TestLinkClick(t *testing.T) {
	n := 0
	h := uitest.New(Link("文档", func() { n++ }))
	h.Click(5, 5)
	if n != 1 {
		t.Fatalf("clicked %d times", n)
	}
}

func TestLinkDisabledRecovery(t *testing.T) {
	n := 0
	l := Link("文档", func() { n++ })
	h := uitest.NewFunc(func(gtx core.C) {
		gtx.Execute(key.FocusCmd{Tag: &l.click})
		l.Layout(gtx)
	})
	clickNamed(t, h, "文档")
	l.SetDisabled(true)
	h.Frame()
	clickNamed(t, h, "文档")
	h.Key(key.NameSpace, 0)
	h.Key(key.NameReturn, 0)
	if n != 1 || h.Router.Source().Focused(&l.click) {
		t.Fatal("disabled link activated or focused")
	}
	l.SetDisabled(false)
	h.Frame()
	clickNamed(t, h, "文档")
	h.Key(key.NameSpace, 0)
	if n != 3 {
		t.Fatalf("link did not recover: %d", n)
	}
}
