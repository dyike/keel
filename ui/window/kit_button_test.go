package window

import (
	"bytes"
	"gioui.org/f32"
	"gioui.org/io/pointer"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"github.com/dyike/keel/ui/theme"
	"image/color"
	"image/png"
	"testing"
)

func TestKitButtonSnapshotAndKeyboard(t *testing.T) {
	calls := 0
	v := kit.Button("保存 123", func() { calls++ }).Icon(kit.IconPlus)
	w := openTest(t, Options{Content: el.Embed(v)})
	before := element(t, w, "保存 123")
	if before.Role != "button" || before.Disabled || before.Value != "" {
		t.Fatalf("button semantics: %+v", before)
	}
	for _, chord := range []string{"tab", "space", "enter"} {
		if err := w.press(chord); err != nil {
			t.Fatal(err)
		}
	}
	if calls != 2 {
		t.Fatalf("keyboard calls %d", calls)
	}
	v.SetLoading(true)
	busy := element(t, w, "保存 123")
	if busy.Role != "button" || busy.Disabled || busy.Value != "loading" || busy.Width != before.Width || busy.Height != before.Height {
		t.Fatalf("loading semantics/bounds: %+v", busy)
	}
	if len(w.snapshot()) != 1 {
		t.Fatalf("loading decoration leaked into snapshot: %+v", w.snapshot())
	}
	w.click(busy.center())
	w.press("space")
	if calls != 2 {
		t.Fatal("loading button activated")
	}
	v.SetLoading(false)
	v.SetDisabled(true)
	if e := element(t, w, "保存 123"); !e.Disabled || e.Value != "" {
		t.Fatalf("disabled semantics %+v", e)
	}
	v.SetDisabled(false)
	v.SetText("继续")
	w.click(element(t, w, "继续").center())
	if calls != 3 {
		t.Fatal("button did not recover")
	}
}

func TestKitButtonRuntimeColorsAndPointerStyles(t *testing.T) {
	old := theme.Current()
	defer core.Update(func() { theme.Apply(old) })
	v := kit.Button("按钮", func() {})
	w := openTest(t, Options{Width: 240, Height: 140, Content: el.Embed(v)})
	sample := func() color.NRGBA {
		t.Helper()
		e := element(t, w, "按钮")
		data, err := w.screenshot()
		if err != nil {
			t.Fatal(err)
		}
		im, err := png.Decode(bytes.NewReader(data))
		if err != nil {
			t.Fatal(err)
		}
		return color.NRGBAModel.Convert(im.At(e.X+8, e.Y+5)).(color.NRGBA)
	}
	for _, p := range []theme.Palette{theme.Light(), theme.Dark()} {
		core.Update(func() { theme.Apply(p) })
		for _, tc := range []struct {
			variant   kit.ButtonVariant
			bg, hover color.NRGBA
		}{
			{kit.ButtonPrimary, p.Primary, p.PrimaryHover}, {kit.ButtonSecondary, p.Subtle, p.SubtleHover},
			{kit.ButtonGhost, p.Bg, p.Subtle}, {kit.ButtonDanger, p.Danger, p.DangerHover},
		} {
			v.Variant(tc.variant)
			w.virt.router.Queue(pointer.Event{Kind: pointer.Move, Source: pointer.Mouse, Position: f32.Pt(230, 130)})
			if got := sample(); got != tc.bg {
				t.Fatalf("variant %d normal %v want %v", tc.variant, got, tc.bg)
			}
			e := element(t, w, "按钮")
			pos := e.center()
			w.virt.router.Queue(pointer.Event{Kind: pointer.Move, Source: pointer.Mouse, Position: pos})
			if got := sample(); got != tc.hover {
				t.Fatalf("variant %d hover %v want %v", tc.variant, got, tc.hover)
			}
			w.virt.router.Queue(pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, Position: pos, Buttons: pointer.ButtonPrimary})
			if got := sample(); got == tc.hover {
				t.Fatalf("variant %d missing active style", tc.variant)
			}
			w.virt.router.Queue(pointer.Event{Kind: pointer.Release, Source: pointer.Mouse, Position: pos})
			w.render()
			v.SetLoading(true)
			if got := sample(); got != tc.bg {
				t.Fatalf("busy hover: %v want %v", got, tc.bg)
			}
			w.virt.router.Queue(pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, Position: pos, Buttons: pointer.ButtonPrimary})
			if got := sample(); got != tc.bg {
				t.Fatalf("busy active: %v want %v", got, tc.bg)
			}
			w.virt.router.Queue(pointer.Event{Kind: pointer.Release, Source: pointer.Mouse, Position: pos})
			w.render()
			v.SetLoading(false)
			v.SetDisabled(true)
			want := p.Subtle
			if tc.variant == kit.ButtonGhost {
				want = p.Bg
			}
			if got := sample(); got != want {
				t.Fatalf("disabled background: %v want %v", got, want)
			}
			v.SetDisabled(false)
		}
	}
}

func TestKitButtonLoadingDecorationColors(t *testing.T) {
	old := theme.Current()
	defer core.Update(func() { theme.Apply(old) })
	core.Update(func() { theme.Apply(theme.Light()) })
	for _, icon := range []bool{false, true} {
		v := kit.Button("加载 ABC", nil).Loading(true)
		if icon {
			v.Icon(kit.IconPlus)
		}
		w := openTest(t, Options{Width: 240, Height: 100, Content: el.Embed(v)})
		e := element(t, w, "加载 ABC")
		data, err := w.screenshot()
		if err != nil {
			t.Fatal(err)
		}
		im, err := png.Decode(bytes.NewReader(data))
		if err != nil {
			t.Fatal(err)
		}
		white, muted := 0, 0
		for y := e.Y; y < e.Y+e.Height; y++ {
			for x := e.X; x < e.X+e.Width; x++ {
				c := color.NRGBAModel.Convert(im.At(x, y))
				if c == theme.OnColor {
					white++
				}
				if c == theme.Muted {
					muted++
				}
			}
		}
		if white == 0 || muted != 0 {
			t.Fatalf("loading colors icon=%v: white=%d muted=%d", icon, white, muted)
		}
	}
}

func TestKitButtonBusyKeepsKeyboardFocus(t *testing.T) {
	calls := 0
	var v *kit.ButtonView
	v = kit.Button("save", func() { calls++; v.SetLoading(true) })
	next := 0
	w := openTest(t, Options{Content: el.Embed(el.ViewFunc(func(cx *el.Context) el.Element {
		return el.Div().Child(v.Render(cx), kit.Button("next", func() { next++ }).Render(cx))
	}))})
	w.press("tab")
	w.press("enter")
	w.press("enter")
	if calls != 1 {
		t.Fatalf("busy activated: %d", calls)
	}
	v.SetLoading(false)
	w.press("enter")
	if calls != 2 {
		t.Fatalf("busy lost focus: %d", calls)
	}
	w.press("tab")
	w.press("enter")
	if next != 1 {
		t.Fatal("Tab did not advance from busy button")
	}
	w.press("shift+tab")
	v.SetLoading(false)
	w.press("enter")
	if calls != 3 {
		t.Fatal("busy button was removed from Tab order")
	}
	v.SetDisabled(true)
	if e := element(t, w, "save"); !e.Disabled {
		t.Fatal("disabled must override busy")
	}
}
