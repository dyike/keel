package widget

import (
	"gioui.org/io/key"
	"github.com/dyike/keel/ui/internal/uitest"
	"reflect"
	"testing"
)

func TestCheckboxMixedAndDisabled(t *testing.T) {
	var changes []bool
	c := Checkbox("全部", false).OnChange(func(v bool) { changes = append(changes, v) })
	c.SetIndeterminate(true)
	h := uitest.New(c)
	clickNamed(t, h, "全部")
	if c.Indeterminate() || !c.Value() || !reflect.DeepEqual(changes, []bool{true}) {
		t.Fatalf("mixed activation: %v %v %v", c.Indeterminate(), c.Value(), changes)
	}
	h.Key(key.NameSpace, 0)
	if c.Value() || len(changes) != 2 {
		t.Fatal("space did not toggle checkbox")
	}
	c.SetDisabled(true)
	c.SetIndeterminate(true)
	h.Frame()
	clickNamed(t, h, "全部")
	h.Key(key.NameReturn, 0)
	if c.Value() || !c.Indeterminate() || len(changes) != 2 {
		t.Fatal("disabled mixed checkbox changed")
	}
	c.SetValue(true)
	h.Frame()
	if c.Indeterminate() || len(changes) != 2 {
		t.Fatal("silent setter did not resolve mixed state")
	}
}
