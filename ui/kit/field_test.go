package kit

import (
	"fmt"
	"testing"

	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/theme"
)

// Every single-line field draws the same frame, theme.ControlHeight tall, so
// fields side by side line up. Each is measured between two marker lines.
func TestFieldsShareControlHeight(t *testing.T) {
	var q string
	fields := []struct {
		name string
		view el.View
	}{
		{"Input", Input("").Placeholder("p")},
		{"Input with prefix", Input("").Prefix(text("¥"))},
		{"NumberInput", NumberInput("")},
		{"Select", Select("", "a", "b")},
		{"Combobox", Combobox("", "a", "b")},
		{"TimeField", TimeField("")},
		{"DatePicker", DatePicker("")},
		{"InputGroup", InputGroup("", Input(""))},
		{"el.Input", viewFunc(func(*el.Context) el.Element { return el.Input().Bind(&q) })},
	}
	h := renderView(viewFunc(func(cx *el.Context) el.Element {
		col := el.Div().W(el.Dp(320)).Items(el.Stretch)
		for i, f := range fields {
			col.Child(el.Text(fmt.Sprint("m", i)), el.Div().NoShrink().Items(el.Stretch).Child(f.view.Render(cx)))
		}
		return col.Child(el.Text(fmt.Sprint("m", len(fields))))
	}), 320, 1)
	h.Frame()
	want := int(theme.ControlHeight)
	for i, f := range fields {
		above, below := bounds(h, fmt.Sprint("m", i)), bounds(h, fmt.Sprint("m", i+1))
		if got := below.Min.Y - above.Max.Y; got != want {
			t.Errorf("%s: %ddp tall, want %d", f.name, got, want)
		}
	}
}
