//go:build darwin && !ios

// Two real windows exercise native font identities and immutable mask pages
// while alternating glyphs, fractional sizes, path clips and opacity layers.
package main

import (
	"fmt"
	"image"
	"image/color"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"github.com/dyike/keel/gio/font"
	"github.com/dyike/keel/gio/op"
	"github.com/dyike/keel/gio/op/clip"
	"github.com/dyike/keel/gio/op/paint"
	"github.com/dyike/keel/gio/text"
	"github.com/dyike/keel/gio/unit"
	"github.com/dyike/keel/gio/widget"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/theme"
	"github.com/dyike/keel/ui/window"
	"golang.org/x/image/math/fixed"
)

type probe struct {
	sh    *text.Shaper
	step  int
	drawn atomic.Int64
}

func (p *probe) Layout(gtx core.C) core.D {
	theme.BeginTextFrame()
	face := font.Font{Typeface: "Menlo, Hiragino Sans GB, Apple Symbols, Apple Color Emoji"}
	samples := []string{"中文终端 ⠋⠙⠹ ┌─┐ █ ░ ⚡ 😀", "مرحبا हिन्दी שלום á é 🧑‍💻 🇨🇳 👍🏽"}
	value := fmt.Sprintf("Font cache / frame %d\n%s", p.step, strings.Repeat(samples[p.step%2]+"\n", 5))
	p.sh.LayoutString(text.Parameters{Font: face, PxPerEm: fixed.I(20), MaxWidth: 2000}, value)
	for g, ok := p.sh.NextGlyph(); ok; g, ok = p.sh.NextGlyph() {
		if g.ID != 0 {
			if _, _, _, ok := p.sh.GlyphNativeFont(g.ID); !ok {
				panic("Go font parsing replaced the native path")
			}
		}
	}
	paint.Fill(gtx.Ops, color.NRGBA{R: 28, G: 32, B: 44, A: 255})
	border := clip.UniformRRect(image.Rectangle{Max: gtx.Constraints.Max}, 15).Push(gtx.Ops)
	alpha := float32(1)
	if p.step%3 == 0 {
		alpha = .7
	}
	layer := paint.PushOpacity(gtx.Ops, alpha)
	col := color.NRGBA{R: 210, G: 228, B: 248, A: 255}
	dims := (widget.Label{}).LayoutWithColor(gtx, p.sh, face, unit.Sp(18+float32(p.step%4)*.25), value, paintOp(gtx, col), col)
	layer.Pop()
	border.Pop()
	p.drawn.Store(int64(p.step + 1))
	return dims
}
func paintOp(gtx core.C, col color.NRGBA) (call op.CallOp) {
	m := op.Record(gtx.Ops)
	paint.ColorOp{Color: col}.Add(gtx.Ops)
	return m.Stop()
}
func fail(message string) { fmt.Println("FAIL:", message); os.Exit(1) }
func main() {
	time.AfterFunc(20*time.Second, func() { fail("native text cache timeout") })
	set, err := theme.PrepareFontFiles(nil, "/System/Library/Fonts/Menlo.ttc", "/System/Library/Fonts/Hiragino Sans GB.ttc")
	if err != nil {
		fail(err.Error())
	}
	set.Use()
	probes := []*probe{{sh: theme.NewShaper()}, {sh: theme.NewShaper()}}
	var windows []*window.Window
	for i, p := range probes {
		windows = append(windows, window.Open(window.Options{Title: fmt.Sprintf("Keel native text cache %d", i), Width: 680, Height: 300, Frameless: true, Content: p}))
	}
	go func() {
		for step := 0; step < 12; step++ {
			next := step
			done := make(chan struct{})
			core.Update(func() {
				for _, p := range probes {
					p.step = next
				}
				close(done)
			})
			<-done
			width := 680 + (step%3)*70
			for _, w := range windows {
				if err := w.Resize(width, 300+(step%2)*30); err != nil {
					fail(err.Error())
				}
			}
			deadline := time.Now().Add(3 * time.Second)
			for {
				ready := true
				for i, p := range probes {
					w, h := windows[i].Size()
					ready = ready && p.drawn.Load() >= int64(step+1) && w == width && h == 300+(step%2)*30
				}
				if ready {
					break
				}
				if time.Now().After(deadline) {
					fail("native redraw or resize stalled")
				}
				time.Sleep(20 * time.Millisecond)
			}
		}
		fmt.Println("PASS: two native windows, 12 glyph/size/clip/opacity changes and resizes")
		for _, w := range windows {
			w.Close()
		}
	}()
	window.Main()
}
