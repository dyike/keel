package svgicon

import (
	"bytes"
	"image/png"
	"os"
	"testing"
)

// The icon renders from its SVG source: transparent rounded corners, the
// blue gradient behind, white sails, hull and keel.
func TestRenderIcon(t *testing.T) {
	svg, err := os.ReadFile("../../docs/images/keel.svg")
	if err != nil {
		t.Fatal(err)
	}
	data, err := Render(svg, 512)
	if err != nil {
		t.Fatal(err)
	}
	if out := os.Getenv("KEEL_ICON_PNG"); out != "" {
		os.WriteFile(out, data, 0o644)
	}
	img, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	at := func(x, y float64) (r, g, b, a uint32) {
		r, g, b, a = img.At(int(x*8), int(y*8)).RGBA()
		return r >> 8, g >> 8, b >> 8, a >> 8
	}
	if _, _, _, a := at(0.3, 0.3); a != 0 {
		t.Errorf("rounded corner is not transparent: alpha %d", a)
	}
	if r, g, b, a := at(5, 5); a != 255 || b < 200 || r > 90 {
		t.Errorf("background near top left is not light blue: %d %d %d %d", r, g, b, a)
	}
	if _, _, b, _ := at(59, 59); b > 200 {
		t.Errorf("background near bottom right is not the darker blue: b=%d", b)
	}
	for name, p := range map[string][2]float64{"mainsail": {38, 28}, "hull": {32, 40}, "keel": {32, 52}} {
		if r, g, b, _ := at(p[0], p[1]); r < 240 || g < 240 || b < 240 {
			t.Errorf("%s is not white: %d %d %d", name, r, g, b)
		}
	}
	if r, _, b, _ := at(26, 30); r > 220 || b < 240 {
		t.Errorf("jib is not light blue: r=%d b=%d", r, b)
	}
	if _, err := Render(svg, 16); err != nil {
		t.Fatal(err)
	}
}

func TestPathTokens(t *testing.T) {
	got := pathTokens("M34 9c6-6 12 15.5.5 24H34z")
	want := []string{"M", "34", "9", "c", "6", "-6", "12", "15.5", ".5", "24", "H", "34", "z"}
	if len(got) != len(want) {
		t.Fatalf("tokens %q", got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("tokens %q, want %q", got, want)
		}
	}
}
