package window

import (
	"bytes"
	"github.com/dyike/keel/third_party/gio/f32"
	"github.com/dyike/keel/third_party/gio/io/pointer"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"github.com/dyike/keel/ui/theme"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

func TestKitRadarAndSankeyPixelsAndAgent(t *testing.T) {
	old := theme.Current()
	defer core.Update(func() { theme.Apply(old) })
	for _, dark := range []bool{false, true} {
		palette := theme.Light()
		mode := "light"
		if dark {
			palette = theme.Dark()
			mode = "dark"
		}
		core.Update(func() { theme.Apply(palette) })
		radar := kit.RadarChart([]string{"Speed", "Quality", "Cost", "Reach"}, kit.Series{Name: "A", Values: []float64{90, 80, 50, 70}}, kit.Series{Name: "B", Values: []float64{60, 90, 80, 50}}).Title("Radar").RadarMax(100)
		sankey := kit.SankeyChart([]kit.SankeyNode{{Name: "Revenue"}, {Name: "Cost"}, {Name: "Profit"}}, []kit.SankeyLink{{Source: 0, Target: 1, Value: 60}, {Source: 0, Target: 2, Value: 40}}).Title("Flow")
		radar.TooltipContent(func(_ *el.Context, d kit.ChartTooltip) el.Element { return el.Text("hover:" + d.Label) })
		sankey.TooltipContent(func(_ *el.Context, d kit.SankeyTooltip) el.Element { return el.Text("hover:" + d.Node.Name) })
		for name, view := range map[string]el.View{"radar": radar, "sankey": sankey} {
			w := openTest(t, Options{Width: 600, Height: 420, Content: el.Root(el.ViewFunc(func(cx *el.Context) el.Element { return el.Div().P(16).Items(el.Stretch).Child(view.Render(cx)) }))})
			w.render()
			w.render()
			data, err := w.screenshot()
			if err != nil {
				t.Fatal(err)
			}
			im, err := png.Decode(bytes.NewReader(data))
			if err != nil {
				t.Fatal(err)
			}
			colored := 0
			for y := 80; y < 350; y++ {
				for x := 40; x < 560; x++ {
					r, g, b, _ := im.At(x, y).RGBA()
					if max(r, g, b)-min(r, g, b) > 3000 {
						colored++
					}
				}
			}
			if dir := os.Getenv("KEEL_CHART_SNAPSHOT_DIR"); dir != "" {
				if err := os.MkdirAll(dir, 0755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(dir, name+"-"+mode+".png"), data, 0644); err != nil {
					t.Fatal(err)
				}
			}
			if colored < 150 {
				t.Fatal(name, "chart marks absent", colored)
			}
			target := "Revenue"
			if name == "radar" {
				target = "Speed"
			}
			e := element(t, w, target)
			x, y := float32(e.X+e.Width/2), float32(e.Y+e.Height/2)
			if name == "radar" {
				x, y = float32(e.X+e.Width/2), float32(e.Y+e.Height+32)
			}
			w.virt.router.Queue(pointer.Event{Kind: pointer.Move, Source: pointer.Mouse, Position: f32.Pt(x, y)})
			w.render()
			w.render()
			if _, err := w.find("", "hover:"+target); err != nil {
				t.Fatal(name, "hover tooltip missing", err)
			}
			if name == "radar" {
				tip := element(t, w, "hover:"+target)
				if tip.Y <= int(y) || tip.X < int(x) {
					t.Fatalf("radar tooltip is detached from pointer (%v, %v): %+v", x, y, tip)
				}
			}
			core.Update(func() {
				if name == "radar" {
					radar.SetDisabled(true)
				} else {
					sankey.SetDisabled(true)
				}
			})
			w.render()
			if _, err := w.find("", "hover:"+target); err == nil {
				t.Fatal(name, "disabled tooltip remained")
			}
			core.Update(func() {
				if name == "radar" {
					radar.SetDisabled(false)
				} else {
					sankey.SetDisabled(false)
				}
			})
			w.render()
			w.click(element(t, w, "查看数据表").center())
			row := "Revenue | Cost | 60"
			if name == "radar" {
				row = "Speed | 90 | 60"
			}
			if roleOfName(w, row) != "row" {
				t.Fatal(name, "agent data row missing")
			}
		}
	}
}
