package window

import (
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"strings"
	"testing"
)

func TestTerminalAgentInputAndResize(t *testing.T) {
	term := kit.Terminal().Name("Console").Height(260)
	var input string
	term.OnData(func(s string) { input += s; term.WriteString(s) })
	w := openTest(t, Options{Width: 390, Height: 400, Content: el.Root(term)})
	e := element(t, w, "Console")
	if e.Role != "textbox" || e.Height < 260 {
		t.Fatal(e)
	}
	w.click(e.center())
	if err := w.typeText("echo 中文"); err != nil {
		t.Fatal(err)
	}
	if err := w.press("enter"); err != nil {
		t.Fatal(err)
	}
	if input != "echo 中文\r" || !strings.Contains(term.Text(), "echo 中文") {
		t.Fatal(input, term.Text())
	}
	term.SetDisabled(true)
	if !element(t, w, "Console").Disabled {
		t.Fatal("disabled semantics")
	}
}
