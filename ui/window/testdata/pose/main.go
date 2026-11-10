// A native window checks that pose motion paints without rebuilding the view,
// can be retargeted while resizing, and stops drawing after settling.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/theme"
	"github.com/dyike/keel/ui/window"
)

type probe struct {
	root            *el.RootWidget
	target          el.Pose
	width           float32
	title           string
	count           int
	full            bool
	hero, inert     bool
	pixels          atomic.Uint64
	frames, renders atomic.Int64
	ready           chan struct{}
	once            sync.Once
}

func (p *probe) Render(*el.Context) el.Element {
	p.renders.Add(1)
	owner, x, size := "thumbnail", float32(0), float32(40)
	radius, alpha, col := float32(4), float32(1), theme.Primary
	title, textWidth, textSize := "Keel 共享文字", float32(180), float32(18)
	if p.hero {
		owner, x, size = "detail", 240, 100
		radius, alpha, col = 30, .7, theme.Muted
		title, textWidth, textSize = "共享文字 / Keel", 300, 24
	}
	hero := el.Div().H(el.Dp(120)).Items(el.Start).Child(el.Div().ID(owner).W(el.Dp(size)).H(el.Dp(110)).Translate(x, 0).Inert(p.inert).ScrollY().Items(el.Start).Child(
		el.Div().Name("Shared element").Size(el.Dp(size)).Rounded(radius).Opacity(alpha).BgGradient(theme.Gradient{From: col, To: theme.Success, Angle: radius * 4}).Shadow(theme.Elevation{Offset: radius / 3, Blur: radius}).SharedLayout("native-hero", el.Tween(600*time.Millisecond, el.Linear)).Fly(),
	))
	return el.Div().P(40).Items(el.Start).Child(
		el.Div().ID("card").W(el.Dp(p.width)).H(el.Dp(80)).P(12).Rounded(12).Bg(theme.Primary).
			AnimatePose(p.target, el.Tween(600*time.Millisecond, el.Linear)).AnimateLayout(el.Tween(600*time.Millisecond, el.Linear)).Child(el.Text(p.title).ID("title").TextColor(theme.OnColor).ContentTransition(el.ContentTransition{Morph: true, Animation: el.Tween(600*time.Millisecond, el.Linear)})),
		el.Text(fmt.Sprintf("%04d", p.count)).ID("count").TextSize(32).ContentTransition(el.ContentTransition{Numeric: true, Animation: el.Tween(600*time.Millisecond, el.Linear)}),
		hero,
		el.Div().ID("text-"+owner).Translate(x, 0).Inert(p.inert).Items(el.Start).
			SharedLayout("native-shared-text", el.Tween(600*time.Millisecond, el.Linear)).MatchText().Fly().Child(
			el.Text(title).ID("matched-title").W(el.Dp(textWidth)).H(el.Dp(60)).TextSize(textSize),
		),
	)
}
func (p *probe) FillsWindow() bool { return true }
func (p *probe) RetainFrames() {
	if !p.full {
		p.root.RetainFrames()
	}
}
func (p *probe) Layout(c core.C) core.D {
	d := p.root.Layout(c)
	p.frames.Add(1)
	p.pixels.Store(uint64(c.Constraints.Max.X)<<32 | uint64(uint32(c.Constraints.Max.Y)))
	p.once.Do(func() { close(p.ready) })
	return d
}
func fail(message string) { fmt.Println("FAIL:", message); os.Exit(1) }
func main() {
	time.AfterFunc(20*time.Second, func() { fail("native pose timeout; requires a visible drawing window") })
	full := flag.Bool("full", false, "rebuild the same animated view on every frame for comparison")
	flag.Parse()
	p := &probe{ready: make(chan struct{}), width: 160, title: "Keel 动画 / pose", count: 19, full: *full}
	p.root = el.Root(p)
	w := window.Open(window.Options{Title: "Keel native pose", Width: 640, Height: 400, Content: p})
	go func() {
		<-p.ready
		time.Sleep(300 * time.Millisecond)
		for i := range 7 {
			p.record(fmt.Sprintf("phase-%d", i))
			frames, renders := p.frames.Load(), p.renders.Load()
			if i == 1 {
				if err := w.Resize(720, 460); err != nil {
					fail(err.Error())
				}
			}
			done := make(chan struct{})
			core.Update(func() {
				switch i {
				case 0:
					p.target = el.Pose{X: 180, Y: 20, Scale: 1.1, Rotate: 10, Fade: .1}
				case 1:
					p.target = el.Pose{}
					p.width = 240
				case 2:
					p.title = "动画 / Keel text 👩🏽‍💻"
				case 3:
					p.count = 20
				case 4:
					p.hero = true
				case 5:
					p.hero = false
				case 6:
					p.hero, p.inert = true, true
				}
				close(done)
			})
			<-done
			time.Sleep(1200 * time.Millisecond)
			f, r := p.frames.Load()-frames, p.renders.Load()-renders
			if f < 10 || !p.full && r*2 >= f || p.full && r < f {
				fail(fmt.Sprintf("phase %d: %d layouts, %d renders; window occluded or animation rebuilt", i, f, r))
			}
			fmt.Printf("phase %d: layouts=%d renders=%d\n", i, f, r)
		}
		p.record("idle")
		idle := p.frames.Load()
		time.Sleep(3 * time.Second)
		if n := p.frames.Load() - idle; n > 1 {
			fail(fmt.Sprintf("settled animation drew %d extra frames", n))
		}
		p.record("done")
		fmt.Println("PASS: native pose/layout/text/shared motion, resize/retarget and settled idle")
		w.Close()
	}()
	window.Main()
}

func (p *probe) record(phase string) {
	json.NewEncoder(os.Stdout).Encode(map[string]any{"phase": phase, "time": time.Now(), "frames": p.frames.Load(), "renders": p.renders.Load(), "pixels": p.pixels.Load(), "full": p.full})
}
