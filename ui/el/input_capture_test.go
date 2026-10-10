package el

import (
	"strings"
	"testing"

	"gioui.org/io/key"
	"github.com/dyike/keel/ui/internal/uitest"
)

// A text area can send on Enter and still break lines with Shift+Enter.
func TestTextAreaCapturesPlainEnter(t *testing.T) {
	value, sent := "", 0
	root := Root(viewFunc(func(*Context) Element {
		return Div().Child(TextArea().ID("draft").Bind(&value).W(Dp(200)).CaptureKeys(string(key.NameReturn)).OnKey(func(e KeyEvent) bool {
			if e.State == KeyPress {
				sent++
			}
			return true
		}))
	}))
	h := uitest.New(root)
	h.Click(20, 20)
	h.Type("hi")
	h.Key(key.NameReturn, 0)
	h.Frame()
	if sent != 1 || value != "hi" {
		t.Fatalf("plain Enter: sent=%d value=%q", sent, value)
	}
	h.Key(key.NameReturn, key.ModShift)
	h.Frame()
	if sent != 1 || len(value) != 3 || !strings.Contains(value, "\n") {
		t.Fatalf("Shift+Enter: sent=%d value=%q", sent, value)
	}
}
