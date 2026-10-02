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

func TestOtpGroupsAndSizes(t *testing.T) {
	for _, scale := range []int{1, 2} {
		v := OtpInput("Code", 7).Groups(3).Size(40)
		v.SetValue("1234567")
		h := renderView(v, 224, scale)
		previous := 0
		for _, digit := range []string{"1", "2", "3", "4", "5", "6", "7"} {
			b := bounds(h, digit)
			if b.Empty() || b.Min.X < previous || b.Max.X > 224*scale {
				t.Fatalf("digit %s: %v", digit, b)
			}
			previous = b.Max.X
		}
		x := func(s string) int { b := bounds(h, s); return (b.Min.X + b.Max.X) / 2 }
		if x("4")-x("3") <= x("3")-x("2")+2*scale || x("6")-x("5") <= x("5")-x("4")+2*scale {
			t.Fatal("group boundaries missing")
		}
		v.Groups(1).Size(56)
		h.Frame()
		if b := bounds(h, "Code"); b.Dx() > 224*scale {
			t.Fatal("size change overflow", b)
		}
		if v.Value() != "1234567" {
			t.Fatal("layout configuration changed value")
		}
	}
}
