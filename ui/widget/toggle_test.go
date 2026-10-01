package widget

import (
	"gioui.org/io/key"
	"github.com/dyike/keel/ui/internal/uitest"
	"testing"
)

func TestToggleInteraction(t *testing.T) {
	var changes []bool
	v := Toggle("固定", false).OnChange(func(b bool) { changes = append(changes, b) })
	h := uitest.New(v)
	clickNamed(t, h, "固定")
	if !v.Value() || len(changes) != 1 {
		t.Fatal("click did not toggle")
	}
	h.Key(key.NameSpace, 0)
	if v.Value() || len(changes) != 2 {
		t.Fatal("space did not toggle")
	}
	h.Key(key.NameReturn, 0)
	if !v.Value() || len(changes) != 3 {
		t.Fatal("enter did not toggle")
	}
	v.SetValue(false)
	h.Frame()
	if len(changes) != 3 {
		t.Fatal("setter notified callback")
	}
	v.SetDisabled(true)
	h.Frame()
	clickNamed(t, h, "固定")
	h.Key(key.NameReturn, 0)
	h.Key(key.NameSpace, 0)
	if v.Value() || len(changes) != 3 {
		t.Fatal("disabled toggle activated")
	}
	v.SetDisabled(false)
	h.Frame()
	clickNamed(t, h, "固定")
	if !v.Value() {
		t.Fatal("re-enabled toggle unresponsive")
	}
}
