package kit

import (
	"github.com/dyike/keel/ui/el"
	"testing"
)

func TestFormValidatesControlsWithoutErrorPresenter(t *testing.T) {
	checked := Checkbox("同意条款", false)
	calls := 0
	f := Form().Field("条款", checked, func() string {
		calls++
		if !checked.Value() {
			return "必须同意"
		}
		return ""
	})
	var cx *el.Context
	page(viewFunc(func(c *el.Context) el.Element { cx = c; return f.Render(c) }))
	if f.Validate(cx) || calls != 1 {
		t.Fatal("validator silently skipped for Checkbox")
	}
	checked.SetValue(true)
	if !f.Validate(cx) || calls != 2 {
		t.Fatal("valid field rejected")
	}
}

func TestRequiredWhitespace(t *testing.T) {
	for _, s := range []string{"", " \t\r\n", "\u3000", "\u00a0"} {
		if Required(s, "required") == "" {
			t.Fatalf("blank value accepted: %q", s)
		}
	}
	if Required("订单 123", "required") != "" {
		t.Fatal("nonblank rejected")
	}
}
