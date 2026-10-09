//go:build darwin && !ios

package main

import (
	"fmt"
	"math"
	"os"
	"strings"

	"github.com/dyike/keel/ui/internal/appkit"
)

func checkPositions(move, report bool, moved bool) bool {
	ok := true
	appkit.MainSync(func() {
		count := 0
		windows := appkit.Send(appkit.App(), "windows")
		for i := uintptr(0); i < uintptr(appkit.Send(windows, "count")); i++ {
			window := appkit.Send(windows, "objectAtIndex:", i)
			title := appkit.GoString(appkit.Send(window, "title"))
			if !strings.HasPrefix(title, "Center ") {
				continue
			}
			count++
			r := appkit.MsgRect(appkit.Send(window, "screen"), appkit.Sel("visibleFrame"))
			frame := appkit.MsgRect(window, appkit.Sel("frame"))
			expected := appkit.Point{
				X: r.Origin.X + r.Size.Width/2 - frame.Size.Width/2,
				Y: r.Origin.Y + r.Size.Height/2 - frame.Size.Height/2,
			}
			if move {
				appkit.MsgSetPoint(window, appkit.Sel("setFrameOrigin:"), appkit.Point{X: expected.X + 35, Y: expected.Y - 25})
				continue
			}
			if moved {
				expected.X += 35
				expected.Y -= 25
			}
			if math.Abs(frame.Origin.X-expected.X) > 1 || math.Abs(frame.Origin.Y-expected.Y) > 1 {
				ok = false
				if report {
					fmt.Fprintf(os.Stderr, "%s: actual %.1f,%.1f expected %.1f,%.1f\n", title, frame.Origin.X, frame.Origin.Y, expected.X, expected.Y)
				}
			}
		}
		if count != 2 {
			ok = false
			if report {
				fmt.Fprintf(os.Stderr, "expected two windows, got %d\n", count)
			}
		}
	})
	return ok
}
