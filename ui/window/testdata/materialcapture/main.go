//go:build darwin && !ios && !nometal && keelnativeqa

package main

import (
	"fmt"
	"image/color"
	"image/png"
	"os"
	"time"

	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/internal/appkit"
	"github.com/dyike/keel/ui/theme"
	"github.com/dyike/keel/ui/window"
)

func main() {
	time.AfterFunc(15*time.Second, func() { fmt.Println("FAIL: material capture timeout"); os.Exit(1) })
	window.Open(window.Options{Title: "material source", Width: 900, Height: 700, Content: el.Root(el.ViewFunc(func(*el.Context) el.Element {
		return el.Div().Row().WFull().HFull().Child(el.Div().Grow().HFull().Bg(theme.RGB(0xd02030)), el.Div().Grow().HFull().Bg(theme.RGB(0x2040d0)))
	}))})
	front := window.Open(window.Options{Title: "material capture", Width: 480, Height: 320, Frameless: true, Glass: &window.GlassOptions{Style: window.GlassFrosted, CornerRadius: 16}, Content: el.Root(el.ViewFunc(func(*el.Context) el.Element {
		return el.Div().WFull().HFull().Bg(color.NRGBA{}).Child(el.Text("Native compositor material regression"), el.Div().W(el.Dp(40)).H(el.Dp(40)).Bg(theme.RGB(0x123456)))
	}))})
	go func() {
		time.Sleep(time.Second)
		front.Raise()
		appkit.MainSync(func() {
			windows := appkit.Send(appkit.App(), "windows")
			for i := uintptr(0); i < uintptr(appkit.Send(windows, "count")); i++ {
				w := appkit.Send(windows, "objectAtIndex:", i)
				if appkit.Equal(appkit.Send(w, "title"), "material capture") {
					appkit.Send(w, "setLevel:", 3)
				}
			}
		})
		// Capture after idle trimming, when only the presented full-resolution
		// surface should remain. Opaque content must survive along with blur.
		time.Sleep(1500 * time.Millisecond)
		file, err := os.CreateTemp("", "keel-material-*.png")
		if err != nil {
			panic(err)
		}
		path := file.Name()
		file.Close()
		defer os.Remove(path)
		if err := window.NativeQACaptureMaterial("material capture", path); err != nil {
			panic(err)
		}
		f, err := os.Open(path)
		if err != nil {
			panic(err)
		}
		img, err := png.Decode(f)
		f.Close()
		if err != nil {
			panic(err)
		}
		sz := img.Bounds().Size()
		left := color.NRGBAModel.Convert(img.At(sz.X/8, sz.Y/2)).(color.NRGBA)
		right := color.NRGBAModel.Convert(img.At(7*sz.X/8, sz.Y/2)).(color.NRGBA)
		if int(left.R)-int(left.B) < 10 || int(right.B)-int(right.R) < 10 {
			fmt.Printf("FAIL: backdrop colors missing: left=%v right=%v\n", left, right)
			os.Exit(1)
		}
		markers := 0
		for y := 0; y < sz.Y; y++ {
			for x := 0; x < sz.X; x++ {
				c := color.NRGBAModel.Convert(img.At(x, y)).(color.NRGBA)
				// ScreenCaptureKit's display profile can shift the marker's RGB.
				// Keep a conservative dark-blue range, distinct from either backdrop.
				if c.R >= 10 && c.R <= 40 && c.G >= 30 && c.G <= 70 && c.B >= 60 && c.B <= 110 {
					markers++
				}
			}
		}
		if markers < 100 {
			fmt.Printf("FAIL: presented content lost after idle: marker pixels=%d\n", markers)
			os.Exit(1)
		}
		fmt.Printf("PASS: native frosted composition transmits source colors: left=%v right=%v\n", left, right)
		window.Quit()
	}()
	window.Main()
}
