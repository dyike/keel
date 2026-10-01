package widget

import (
	"gioui.org/io/key"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/internal/uitest"
	"reflect"
	"testing"
)

func TestToggleGroupSingleMultipleKeyboard(t *testing.T) {
	var changes [][]string
	g := ToggleGroup("左", "中", "右").OnChange(func(v []string) { changes = append(changes, v) })
	h := uitest.New(g)
	clickNamed(t, h, "左")
	clickNamed(t, h, "中")
	if !reflect.DeepEqual(g.Values(), []string{"中"}) {
		t.Fatal(g.Values())
	}
	g.SetItemDisabled(2, true)
	h.Frame()
	h.Key(key.NameRightArrow, 0)
	h.Key(key.NameSpace, 0)
	if !reflect.DeepEqual(g.Values(), []string{"左"}) {
		t.Fatalf("arrow did not skip disabled: %v", g.Values())
	}
	h.Key(key.NameEnd, 0)
	h.Key(key.NameReturn, 0)
	if !reflect.DeepEqual(g.Values(), []string{"中"}) {
		t.Fatal("end key selected wrong item")
	}
	g.Multiple()
	h.Frame()
	clickNamed(t, h, "左")
	if !reflect.DeepEqual(g.Values(), []string{"左", "中"}) {
		t.Fatal("multiple selection failed")
	}
	before := len(changes)
	g.SetValues("中", "missing")
	h.Frame()
	if len(changes) != before || !reflect.DeepEqual(g.Values(), []string{"中"}) {
		t.Fatal("setter was not silent")
	}
	changes[len(changes)-1][0] = "mutated"
	if !reflect.DeepEqual(g.Values(), []string{"中"}) {
		t.Fatal("callback mutated state")
	}
	g.SetDisabled(true)
	h.Frame()
	clickNamed(t, h, "左")
	h.Key(key.NameSpace, 0)
	if !reflect.DeepEqual(g.Values(), []string{"中"}) {
		t.Fatal("disabled group changed")
	}
}

func TestToggleGroupValueAndDisabledRecovery(t *testing.T) {
	changes := 0
	g := ToggleGroup("甲", "乙", "丙").OnChange(func([]string) { changes++ })
	g.SetValue([]string{"甲"})
	g.SetItemDisabled(2, true)
	h := uitest.NewFunc(func(gtx core.C) {
		if g.disabled {
			gtx.Execute(key.FocusCmd{Tag: &g.items[1].click})
		}
		g.Layout(gtx)
	})
	value := g.Value()
	value[0] = "丙"
	if !reflect.DeepEqual(g.Value(), []string{"甲"}) || changes != 0 {
		t.Fatal("value or silent setter failed")
	}
	g.SetDisabled(true)
	h.Frame()
	clickNamed(t, h, "乙")
	h.Key(key.NameSpace, 0)
	h.Key(key.NameEnd, 0)
	if !reflect.DeepEqual(g.Value(), []string{"甲"}) || changes != 0 || h.Router.Source().Focused(&g.items[1].click) {
		t.Fatal("disabled group changed or focused")
	}
	g.SetDisabled(false)
	h.Frame()
	clickNamed(t, h, "乙")
	if !reflect.DeepEqual(g.Value(), []string{"乙"}) || changes != 1 {
		t.Fatal("group did not recover")
	}
	clickNamed(t, h, "丙")
	if !reflect.DeepEqual(g.Value(), []string{"乙"}) {
		t.Fatal("item disabled state lost")
	}
}
