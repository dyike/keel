package widget

import (
	"github.com/dyike/keel/ui/internal/uitest"
	"reflect"
	"testing"
)

func TestInputAdornmentsClearFocusAndReadOnly(t *testing.T) {
	var changes []string
	f := Input("搜索").Prefix(Icon(IconSearch)).Suffix(Kbd("mod+k").Plain()).Clearable().OnChange(func(v string) { changes = append(changes, v) })
	h := uitest.New(f)
	// The label delegates focus and is not an extra keyboard stop.
	h.Click(10, 8)
	h.Frame()
	h.Type("你好")
	if f.Value() != "你好" {
		t.Fatal("label did not focus editor")
	}
	clickNamed(t, h, "清空搜索")
	if f.Value() != "" || !reflect.DeepEqual(changes, []string{"你好", ""}) {
		t.Fatalf("clear callback mismatch: %q %v", f.Value(), changes)
	}
	h.Frame()
	h.Type("再试")
	if f.Value() != "再试" {
		t.Fatal("clear did not return focus")
	}
	f.SetReadOnly(true)
	h.Frame()
	clickNamed(t, h, "清空搜索")
	h.Type("x")
	if f.Value() != "再试" {
		t.Fatal("read-only field cleared or edited")
	}
	f.SetReadOnly(false)
	f.SetDisabled(true)
	h.Frame()
	clickNamed(t, h, "清空搜索")
	h.Type("y")
	if f.Value() != "再试" {
		t.Fatal("disabled field changed")
	}
}
