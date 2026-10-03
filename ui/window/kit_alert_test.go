package window

import (
	"bytes"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"github.com/dyike/keel/ui/theme"
	"image/color"
	"image/png"
	"testing"
)

func TestKitAlertSnapshot(t *testing.T) {
	a := kit.Alert("保存成功").Description("订单 SO-123 已保存").Tone(kit.ToneSuccess)
	w := openTest(t, Options{Content: el.Embed(a)})
	e := element(t, w, "保存成功")
	if e.Role != "alert" || e.Value != "success" {
		t.Fatalf("alert semantics: %+v", e)
	}
	if e := element(t, w, "订单 SO-123 已保存"); e.Role != "text" {
		t.Fatal("missing description")
	}
	a.SetTitle("保存失败")
	a.Tone(kit.ToneDanger)
	if e := element(t, w, "保存失败"); e.Value != "danger" {
		t.Fatal("stale alert")
	}
}

func TestKitAlertKeyboardDismiss(t *testing.T) {
	n := 0
	a := kit.Alert("提示").OnClose(func() { n++ })
	w := openTest(t, Options{Content: el.Embed(a)})
	if err := w.press("tab"); err != nil {
		t.Fatal(err)
	}
	if err := w.press("enter"); err != nil {
		t.Fatal(err)
	}
	if a.Visible() || n != 1 {
		t.Fatal("keyboard close failed")
	}
}

func TestAlertColorsFollowRuntimePalette(t *testing.T) {
	old := theme.Current()
	defer core.Update(func() { theme.Apply(old) })
	v := kit.Alert("提示").Description("说明")
	w := openTest(t, Options{Width: 260, Height: 150, Content: el.Embed(v)})
	for _, p := range []theme.Palette{theme.Light(), theme.Dark()} {
		core.Update(func() { theme.Apply(p) })
		for _, sample := range []struct {
			tone  kit.Tone
			color color.NRGBA
		}{{kit.ToneInfo, p.Info}, {kit.ToneSuccess, p.Success}, {kit.ToneWarning, p.Warning}, {kit.ToneDanger, p.DangerText}} {
			v.Tone(sample.tone)
			data, err := w.screenshot()
			if err != nil {
				t.Fatal(err)
			}
			im, err := png.Decode(bytes.NewReader(data))
			if err != nil {
				t.Fatal(err)
			}
			count := 0
			for y := 0; y < im.Bounds().Dy(); y++ {
				for x := 0; x < im.Bounds().Dx(); x++ {
					if color.NRGBAModel.Convert(im.At(x, y)) == sample.color {
						count++
					}
				}
			}
			if count < 30 {
				t.Fatalf("missing runtime tone %v", sample.tone)
			}
		}
	}
}

func TestKitAlertBannerRichContentSnapshot(t *testing.T) {
	actions, closes := 0, 0
	a := kit.Alert("Service").Banner(true).Size(kit.AlertSizeSmall).Icon(kit.IconCalendar).Content(kit.Button("Retry service", func() { actions++ })).OnClose(func() { closes++ })
	w := openTest(t, kitPage(a))
	if e := element(t, w, "Service"); e.Role != "alert" {
		t.Fatalf("banner semantics %+v", e)
	}
	w.click(element(t, w, "Retry service").center())
	if actions != 1 || closes != 0 {
		t.Fatal("body action")
	}
	a.SetDisabled(true)
	if e := element(t, w, "Retry service"); !e.Disabled {
		t.Fatal("body disabled semantics")
	}
	a.SetDisabled(false)
	w.click(element(t, w, "关闭 Service").center())
	if a.Visible() || closes != 1 {
		t.Fatal("close")
	}
	a.SetVisible(true)
	if roleOfName(w, "Retry service") != "button" {
		t.Fatal("restore rich body")
	}
}
