package el

import (
	"bytes"
	"testing"

	"github.com/dyike/keel/ui/internal/uitest"
)

type reloadView struct {
	name, secret, note string
	placeholder        string
}

func (v *reloadView) Render(cx *Context) Element {
	rows := Div()
	for range 40 {
		rows.Child(Div().H(Dp(20)).Child(Text("row")))
	}
	return Div().Child(
		Input().Name("Name").Placeholder(v.placeholder).Bind(&v.name).W(Dp(200)),
		Input().Name("Secret").Password().Bind(&v.secret).W(Dp(200)),
		Div().ID("list").H(Dp(100)).ScrollY().Child(rows),
	)
}

func scrollOf(r *RootWidget, id string) int {
	var y int
	var walk func(n *Node)
	walk = func(n *Node) {
		if n.id == id {
			if st := r.store.states[n.key]; st != nil {
				y = st.scrollY
			}
		}
		for _, c := range n.children {
			walk(c.node())
		}
	}
	walk(r.mainTree)
	return y
}

// A reload keeps scroll offsets and input text, never a password, and only
// for the same element.
func TestReloadStateRoundTrip(t *testing.T) {
	v := &reloadView{placeholder: "Your name"}
	root := Root(v)
	h := uitest.New(root)
	h.Click(20, 10)
	h.Type("Ada")
	v.secret = "hunter2" // as if typed
	h.Frame()
	h.Scroll(50, 120, 60)
	if v.name != "Ada" || scrollOf(root, "list") == 0 {
		t.Fatalf("setup: name %q scroll %d", v.name, scrollOf(root, "list"))
	}
	saved := root.SaveReloadState()
	if bytes.Contains(saved, []byte("hunter2")) {
		t.Fatal("a password reached the saved state")
	}

	again := &reloadView{placeholder: "Your name"}
	restored := Root(again)
	restored.RestoreReloadState(saved)
	uitest.New(restored)
	if again.name != "Ada" || again.secret != "" || scrollOf(restored, "list") != scrollOf(root, "list") {
		t.Fatalf("restored name %q secret %q scroll %d, want Ada, empty, %d", again.name, again.secret, scrollOf(restored, "list"), scrollOf(root, "list"))
	}

	// The input changed in code: its old text stays out.
	edited := &reloadView{placeholder: "Full name"}
	other := Root(edited)
	other.RestoreReloadState(saved)
	uitest.New(other)
	if edited.name != "" {
		t.Fatalf("text restored into a changed input: %q", edited.name)
	}
	if scrollOf(other, "list") == 0 {
		t.Fatal("the unchanged list lost its offset")
	}
}
