package window

import (
	"bytes"
	"image/color"
	"image/png"
	"testing"

	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
)

func TestKitAvatarSnapshot(t *testing.T) {
	w := openTest(t, Options{Content: el.Embed(kit.Avatar("Ada Lovelace").Status(kit.AvatarOnline))})
	if e := element(t, w, "Ada Lovelace"); e.Role != "avatar" || e.Value != "online" {
		t.Fatalf("invalid avatar: %+v", e)
	}
}

func TestKitAvatarGroupSnapshot(t *testing.T) {
	group := kit.AvatarGroup(kit.Avatar("Ada").Status(kit.AvatarOnline), kit.Avatar("Bob"), kit.Avatar("Carol")).Limit(2)
	w := openTest(t, Options{Width: 320, Height: 200, Content: el.Embed(group)})
	if e := element(t, w, "Ada"); e.Role != "avatar" || e.Value != "online" {
		t.Fatalf("member: %+v", e)
	}
	if e := element(t, w, "更多 1"); e.Role != "avatar" || e.Value != "1" {
		t.Fatalf("overflow: %+v", e)
	}
	for _, e := range w.snapshot() {
		if e.Name == "Carol" {
			t.Fatal("hidden member exposed")
		}
	}
	group.Ellipsis(true)
	if e := element(t, w, "更多 1"); e.Value != "1" {
		t.Fatal("ellipsis lost count")
	}
}

// Custom shape, colors and border change the drawn avatar; an avatar with no
// name shows its placeholder icon rather than a letter.
func TestKitAvatarAppearance(t *testing.T) {
	red := color.NRGBA{R: 0xd0, G: 0x20, B: 0x20, A: 0xff}
	green := color.NRGBA{R: 0x10, G: 0xa0, B: 0x40, A: 0xff}
	square := kit.Avatar("Team").Size(80).Rounded(0).Colors(red, color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}).Border(4, green)
	empty := kit.Avatar("").Size(80)
	w := openTest(t, Options{Width: 240, Height: 140, Content: el.Root(el.ViewFunc(func(cx *el.Context) el.Element {
		return el.Div().Row().Gap(16).P(16).Items(el.Start).Child(square.Render(cx), empty.Render(cx))
	}))})
	w.render()
	shot, err := w.screenshot()
	if err != nil {
		t.Fatal(err)
	}
	img, err := png.Decode(bytes.NewReader(shot))
	if err != nil {
		t.Fatal(err)
	}
	b := element(t, w, "Team")
	at := func(x, y int) color.NRGBA { return color.NRGBAModel.Convert(img.At(x, y)).(color.NRGBA) }
	near := func(c, want color.NRGBA) bool {
		d := func(a, b uint8) int { return max(int(a)-int(b), int(b)-int(a)) }
		return d(c.R, want.R) < 40 && d(c.G, want.G) < 40 && d(c.B, want.B) < 40
	}
	if c := at(b.X+1, b.Y+1); !near(c, green) {
		t.Errorf("square corner is not the green border: %v", c)
	}
	if c := at(b.X+10, b.Y+10); !near(c, red) {
		t.Errorf("custom background missing: %v", c)
	}
	e := element(t, w, "Avatar")
	ink := 0
	for y := e.Y + 20; y < e.Y+60; y++ {
		for x := e.X + 20; x < e.X+60; x++ {
			if c := at(x, y); int(c.R)+int(c.G)+int(c.B) < 300 {
				ink++
			}
		}
	}
	if ink < 80 {
		t.Errorf("no placeholder icon drawn: %d dark pixels", ink)
	}
}
