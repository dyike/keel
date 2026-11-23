package main

import (
	"embed"
	"image"
	"image/png"
	"math"
	"sync"

	"github.com/dyike/keel/gio/f32"
	"github.com/dyike/keel/gio/op"
	"github.com/dyike/keel/gio/op/clip"
	"github.com/dyike/keel/gio/op/paint"
	"github.com/dyike/keel/ui/theme"
)

const filmName = "Keel Film"
const filmSeconds = 27

// The title plates preserve the Chinese typography in WebAssembly. The
// interface plates are captures of the hello, components and video examples.
//
//go:embed film/*.png
var filmArt embed.FS

var filmImages = func() map[string]func() paint.ImageOp {
	entries, err := filmArt.ReadDir("film")
	if err != nil {
		panic(err)
	}
	images := make(map[string]func() paint.ImageOp, len(entries))
	for _, entry := range entries {
		name := entry.Name()
		images[name] = sync.OnceValue(func() paint.ImageOp {
			f, err := filmArt.Open("film/" + name)
			if err != nil {
				panic(err)
			}
			defer f.Close()
			img, err := png.Decode(f)
			if err != nil {
				panic(err)
			}
			return paint.NewImageOp(img)
		})
	}
	return images
}()

func place(ops *op.Ops, x, y, scale float32) op.TransformStack {
	return op.Affine(f32.Affine2D{}.Scale(f32.Point{}, f32.Pt(scale, scale)).Offset(f32.Pt(x, y))).Push(ops)
}

func filmImage(ops *op.Ops, name string, x, y, scale float32) {
	defer place(ops, x, y, scale).Pop()
	filmImages[name]().Add(ops)
	paint.PaintOp{}.Add(ops)
}

func easeInShot(t float64) float32 {
	u := float32(1 - min(max(t/.48, 0), 1))
	return u * u * u
}

func filmCursor(ops *op.Ops, x, y float32, p palette) {
	defer place(ops, x, y, 1).Pop()
	var path clip.Path
	path.Begin(ops)
	path.MoveTo(f32.Pt(0, 0))
	path.LineTo(f32.Pt(2, 28))
	path.LineTo(f32.Pt(9, 21))
	path.LineTo(f32.Pt(15, 32))
	path.LineTo(f32.Pt(21, 29))
	path.LineTo(f32.Pt(15, 18))
	path.LineTo(f32.Pt(25, 17))
	path.Close()
	spec := path.End()
	paint.FillShape(ops, p.dark, clip.Outline{Path: spec}.Op())
	paint.FillShape(ops, white, clip.Stroke{Path: spec, Width: 2}.Op())
}

func keelFilm(ops *op.Ops, p palette, phase float64) {
	// Author in 1280 × 720 coordinates inside the player's 640 × 360 canvas.
	defer place(ops, -320, -180, .5).Pop()
	seconds := phase * filmSeconds / (2 * math.Pi)
	starts := [...]float64{0, 4, 10, 16, 22}
	titles := [...]string{"title0.png", "title1.png", "title2.png", "title3.png", "title4.png"}
	shot := 0
	for shot+1 < len(starts) && seconds >= starts[shot+1] {
		shot++
	}
	t := seconds - starts[shot]
	enter := easeInShot(t)
	dark := theme.RGB(0x101c32)
	if shot == 1 || shot == 2 {
		paint.Fill(ops, p.paper)
	} else {
		paint.Fill(ops, dark)
	}
	switch shot {
	case 0, 4:
		x := float32(915)
		size := float32(2.0)
		if shot == 4 {
			x = 1010
			size = 1.28
		}
		transform := place(ops, x+enter*55, 362, size)
		orbit(ops, p, seconds*.96)
		transform.Pop()
	case 1:
		x, y := float32(545)+enter*52, float32(158)
		const scale = 1.03
		name := "form0.png"
		if t >= 1.25 {
			name = "form1.png"
		}
		if t >= 3.45 {
			name = "form2.png"
		}
		filmImage(ops, name, x, y, scale)
		progress := float32(min(max((t-1.8)/1.4, 0), 1))
		progress = progress * progress * (3 - 2*progress)
		filmCursor(ops, x+(204-122*progress)*scale, y+(207+89*progress)*scale, p)
	case 2:
		panel := place(ops, 530+enter*52, 105, .94)
		filmImage(ops, "chart.png", 0, 0, 1)
		progress := min(max((t-.25)/2.7, 0), 1)
		if progress < 1 {
			edge := 115 + int(483*progress)
			cover := clip.Rect(image.Rect(edge, 214, 598, 478)).Push(ops)
			filmImage(ops, "grid.png", 115, 214, 1)
			cover.Pop()
		}
		panel.Pop()
	case 3:
		x, y := float32(493)+enter*52, float32(130)
		const scale = .885
		filmImage(ops, "studio.png", x, y, scale)
		// Match the live artwork to the player rectangle in the captured UI.
		inset := place(ops, x+10*scale, y+149*scale, scale)
		previewBounds := clip.Rect(image.Rect(0, 0, 594, 334)).Push(ops)
		paint.Fill(ops, palettes[0].paper)
		model := place(ops, 297, 167, 334.0/360)
		orbit(ops, palettes[0], math.Mod(t, 4)*math.Pi/2)
		model.Pop()
		previewBounds.Pop()
		inset.Pop()
	}
	filmImage(ops, titles[shot], 0, enter*22, 1)
	// A fast coloured wipe covers the cut between each pair of shots.
	for _, cut := range starts[1:] {
		if delta := math.Abs(seconds - cut); delta < .20 {
			width := float32((1 - delta/.20) * 1280)
			left := float32(0)
			if seconds >= cut {
				left = 1280 - width
			}
			paint.FillShape(ops, p.base, clip.Rect{Min: image.Pt(int(left), 0), Max: image.Pt(int(left+width), 720)}.Op())
		}
	}
}
