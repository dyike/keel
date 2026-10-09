package kit

import (
	"testing"
	"time"

	"gioui.org/io/key"
)

func TestNumberAndTimeArrowKeys(t *testing.T) {
	n := NumberInput("数量").Range(0, 100).Step(5)
	tf := TimeField("开始")
	tf.SetValue(23*time.Hour + 59*time.Minute)
	h := page(n, tf)
	clickClass(t, h, "Editor", "数量")
	for _, k := range []key.Name{key.NameUpArrow, key.NameUpArrow, key.NameDownArrow, key.NamePageUp} {
		h.Key(k, 0)
	}
	if n.Value() != 55 {
		t.Fatalf("number keys: %v", n.Value())
	}
	clickClass(t, h, "Editor", "开始")
	h.Key(key.NameUpArrow, 0) // wraps past midnight
	h.Key(key.NamePageUp, 0)
	if tf.Value() != time.Hour {
		t.Fatalf("time keys: %v", tf.Value())
	}
}

func TestComboboxArrowHighlight(t *testing.T) {
	c := Combobox("城市", "北京", "上海", "深圳")
	h := page(c)
	clickClass(t, h, "Editor", "城市")
	h.Key(key.NameDownArrow, 0) // opens on the first match
	h.Key(key.NameDownArrow, 0)
	h.Key(key.NameDownArrow, 0)
	h.Key(key.NameUpArrow, 0)
	h.Key(key.NameReturn, 0)
	h.Frame()
	if c.Value() != "上海" {
		t.Fatalf("highlighted choice %q", c.Value())
	}
}

func TestSelectSearchDownEntersList(t *testing.T) {
	s := Select("城市", "北京", "上海").Searchable()
	h := page(s)
	clickRole(t, h, "select", "城市")
	h.Frame()
	h.Key(key.NameDownArrow, 0) // search box → first option
	h.Key(key.NameDownArrow, 0)
	h.Key(key.NameReturn, 0)
	h.Frame()
	if s.Value() != "上海" {
		t.Fatalf("value %q", s.Value())
	}
}
