package stroke

import (
	"encoding/binary"
	"math/rand"
	"slices"
	"testing"

	"github.com/dyike/keel/gio/internal/f32"
	"github.com/dyike/keel/gio/internal/ops"
	"github.com/dyike/keel/gio/internal/scene"
)

func strokeCommands(commands []scene.Command, contours []uint32) []byte {
	data := make([]byte, len(commands)*(scene.CommandSize+4))
	for i, c := range commands {
		dst := data[i*(scene.CommandSize+4):]
		binary.LittleEndian.PutUint32(dst, contours[i])
		ops.EncodeCommand(dst[4:], c)
	}
	return data
}

func TestScratchMatchesOriginalStroke(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	point := func() f32.Point { return f32.Pt(rng.Float32()*100, rng.Float32()*100) }
	var scratch Scratch
	for round := range 100 {
		var commands []scene.Command
		var contours []uint32
		for c := uint32(1); c < 4; c++ {
			start := point()
			from := start
			for j := range 4 {
				to := point()
				if j == 3 && round%2 == 0 {
					to = start // Closed contours exercise both winding directions.
				}
				var cmd scene.Command
				switch (round + j) % 3 {
				case 0:
					cmd = scene.Line(from, to)
				case 1:
					cmd = scene.Quad(from, point(), to)
				case 2:
					cmd = scene.Cubic(from, point(), point(), to)
				}
				commands = append(commands, cmd)
				contours = append(contours, c)
				from = to
			}
		}
		data := strokeCommands(commands, contours)
		style := StrokeStyle{Width: .25 + rng.Float32()*12}
		want := StrokePathCommands(style, data)
		for range 3 {
			got := scratch.Commands(style, data)
			if !slices.Equal(got, want) {
				t.Fatalf("stroke differs at round %d: got %d, want %d quads", round, len(got), len(want))
			}
		}
	}
}

func TestScratchWarmCommandsDoNotAllocate(t *testing.T) {
	data := strokeCommands([]scene.Command{scene.Quad(f32.Pt(0, 0), f32.Pt(50, 90), f32.Pt(100, 0))}, []uint32{1})
	var scratch Scratch
	style := StrokeStyle{Width: 3}
	scratch.Commands(style, data)
	if n := testing.AllocsPerRun(100, func() { scratch.Commands(style, data); scratch.Trim() }); n != 0 {
		t.Fatalf("warm stroke allocated %g times", n)
	}
	scratch.output = make(StrokeQuads, 16385)
	scratch.Trim()
	if scratch.output != nil {
		t.Fatal("oversized scratch retained")
	}
}
