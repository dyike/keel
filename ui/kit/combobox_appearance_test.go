package kit

import (
	"github.com/dyike/keel/third_party/gio/io/key"
	"github.com/dyike/keel/ui/locale"
	"math"
	"testing"
)

func TestComboboxSizePreservesFocusAndRows(t *testing.T) {
	for _, scale := range []int{1, 2} {
		c := Combobox("Choice", "a", "b").Clearable(true)
		c.SetValue("a")
		h := renderView(c, 300, scale)
		clear := locale.Current().Name(locale.Current().Clear, "Choice")
		base := bounds(h, clear).Dy()
		clickClass(t, h, "Editor", "Choice")
		c.Size(48)
		h.Frame()
		if bounds(h, clear).Dy() <= base {
			t.Fatal("large clear control not scaled")
		}
		h.Key(key.NameDownArrow, 0)
		h.Frame()
		large := bounds(h, "b").Dy()
		c.Size(28)
		h.Frame()
		h.Frame()
		small := bounds(h, "b").Dy()
		if small >= large || bounds(h, clear).Dy() >= base {
			t.Fatal("small control/row", small, large)
		}
		h.Key(key.NameDownArrow, 0)
		h.Key(key.NameReturn, 0)
		if c.Value() != "b" {
			t.Fatal("resize lost keyboard input", c.Value())
		}
		c.Size(float32(math.NaN())).Size(-1).Size(float32(math.Inf(1)))
		if c.height != 28 {
			t.Fatal("invalid size accepted")
		}
		c.Size(0)
		h.Frame()
		if bounds(h, clear).Dy() != base {
			t.Fatal("default not restored")
		}
	}
}

func TestComboboxCheckIconSnapshotAndLayout(t *testing.T) {
	icon := Icon(IconPlus).Size(22)
	c := Combobox("Choice", "a", "b").CheckIcon(icon)
	icon.Size(40)
	if c.checkIcon.size != 22 {
		t.Fatal("icon configuration not copied")
	}
	c.SetValue("a")
	h := page(c)
	clickClass(t, h, "Editor", "Choice")
	h.Key(key.NameDownArrow, 0)
	h.Frame()
	before := bounds(h, "b")
	c.CheckIcon(Icon(IconNone).Size(22))
	h.Frame()
	if bounds(h, "b") != before {
		t.Fatal("hidden check changed row geometry")
	}
	c.CheckIcon(nil)
	h.Frame()
	if c.checkIcon != nil {
		t.Fatal("default check not restored")
	}
	h.Key(key.NameDownArrow, 0)
	h.Key(key.NameReturn, 0)
	if c.Value() != "b" {
		t.Fatal("icon change lost keyboard identity")
	}
}

func TestComboboxSizeMultipleTagRemoval(t *testing.T) {
	c := Combobox("Tags", "a", "b").Multiple().Size(48).Clearable(true)
	c.SetValues([]string{"a", "b"})
	h := renderView(c, 260, 2)
	click(t, h, locale.Current().Name(locale.Current().Remove, "a"))
	h.Frame()
	if len(c.Values()) != 1 || c.Value() != "b" {
		t.Fatal("large tag removal", c.Values())
	}
	c.Size(28)
	h.Frame()
	click(t, h, locale.Current().Name(locale.Current().Clear, "Tags"))
	h.Frame()
	if len(c.Values()) != 0 {
		t.Fatal("small multi clear", c.Values())
	}
}
