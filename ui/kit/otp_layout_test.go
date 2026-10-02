package kit

import (
	"github.com/dyike/keel/ui/el"
	"testing"
)

func TestOtpNarrowDigitsAndInputShareBounds(t *testing.T) {
	for _, scale := range []int{1, 2} {
		completed := ""
		v := OtpInput("验证码", 6).OnComplete(func(s string) { completed = s })
		v.SetValue("123456")
		h := renderView(viewFunc(func(cx *el.Context) el.Element { return el.Div().W(el.Dp(224)).Child(v.Render(cx)) }), 224, scale)
		last := 0
		for _, digit := range []string{"1", "2", "3", "4", "5", "6"} {
			b := bounds(h, digit)
			if b.Empty() || b.Min.X < last || b.Max.X > 224*scale {
				t.Fatalf("scale %d: digit %s outside narrow field: %v", scale, digit, b)
			}
			last = b.Max.X
		}
		v.SetValue("")
		h.Frame()
		clickClass(t, h, "Editor", "验证码")
		h.Type("654321")
		h.Frame()
		if completed != "654321" || v.Value() != completed {
			t.Fatal(completed, v.Value())
		}
	}
}
