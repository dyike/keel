// SPDX-License-Identifier: Unlicense OR MIT

//go:build darwin && !ios && !nometal

package app

import (
	"fmt"
	"image/color"
	"os"
	"testing"
	"time"
	"unsafe"

	"github.com/dyike/keel/gio/font"
	"github.com/dyike/keel/gio/font/gofont"
	"github.com/dyike/keel/gio/io/key"
	"github.com/dyike/keel/gio/io/system"
	"github.com/dyike/keel/gio/op"
	"github.com/dyike/keel/gio/op/paint"
	"github.com/dyike/keel/gio/text"
	"github.com/dyike/keel/gio/unit"
	"github.com/dyike/keel/gio/widget"
	"github.com/ebitengine/purego/objc"
)

// TestMacOSTextInputClient drives GioView's NSTextInputClient methods the
// way an input method does, through objc_msgSend, under either binding
// (go test, and go test -tags purego). Keel patch: the purego binding
// implements these methods as Go callbacks with struct arguments and
// results, so the ranges and rectangles cross the native ABI both ways.
func TestMacOSTextInputClient(t *testing.T) {
	if os.Getenv("KEEL_DESKTOP") != "1" {
		t.Skip("set KEEL_DESKTOP=1 for native text input regression")
	}
	runMetalHelper(t, "GIO_TEXT_INPUT_HELPER", 20*time.Second)
}

type tiRange struct{ Location, Length uintptr }
type tiRect struct{ X, Y, Width, Height float64 }

const tiNotFound = 1<<63 - 1

func textInputHelper() {
	fail := func(format string, args ...any) {
		fmt.Fprintf(os.Stderr, format+"\n", args...)
		os.Exit(1)
	}
	frames := make(chan string, 32)
	w := new(Window)
	w.Option(Title("Text input regression"), Size(unit.Dp(320), unit.Dp(120)), TopMost(true))
	go func() {
		var (
			ops     op.Ops
			ed      widget.Editor
			focused bool
		)
		shaper := text.NewShaper(text.WithCollection(gofont.Collection()))
		material := func(c color.NRGBA) op.CallOp {
			m := op.Record(&ops)
			paint.ColorOp{Color: c}.Add(&ops)
			return m.Stop()
		}
		for {
			switch e := w.Event().(type) {
			case DestroyEvent:
				if e.Err != nil {
					fail("%v", e.Err)
				}
				os.Exit(0)
			case FrameEvent:
				gtx := NewContext(&ops, e)
				if !focused {
					gtx.Execute(key.FocusCmd{Tag: &ed})
					focused = true
				}
				paint.Fill(gtx.Ops, color.NRGBA{R: 255, G: 255, B: 255, A: 255})
				ed.Layout(gtx, shaper, font.Font{}, unit.Sp(16), material(color.NRGBA{A: 255}), material(color.NRGBA{B: 255, A: 80}))
				e.Frame(gtx.Ops)
				select {
				case frames <- ed.Text():
				default:
				}
			}
		}
	}()
	// waitText waits for a frame showing want.
	waitText := func(want string) {
		deadline := time.NewTimer(3 * time.Second)
		defer deadline.Stop()
		var got string
		for {
			w.Invalidate()
			select {
			case got = <-frames:
				if got == want {
					return
				}
			case <-deadline.C:
				exitIfOccluded(w, "during text input")
				fail("editor shows %q, want %q", got, want)
			}
		}
	}
	go func() {
		// Two frames: the first focuses the editor, the second reports it
		// focused to the window.
		<-frames
		<-frames
		var errs []string
		check := func(ok bool, format string, args ...any) {
			if !ok {
				errs = append(errs, fmt.Sprintf(format, args...))
			}
		}
		str := func(s string) objc.ID {
			return objc.ID(objc.GetClass("NSString")).Send(objc.RegisterName("stringWithUTF8String:"), s+"\x00")
		}
		var view objc.ID
		w.Run(func() {
			view = objc.ID(uintptr(w.driver.(*window).view))
			check(view.Send(objc.RegisterName("inputContext")) != 0, "GioView has no input context; NSTextInputClient missing")
			pool := objc.ID(objc.GetClass("NSAutoreleasePool")).Send(objc.RegisterName("new"))
			defer pool.Send(objc.RegisterName("drain"))
			hasMarked := func() bool {
				return objc.Send[bool](view, objc.RegisterName("hasMarkedText"))
			}
			// Compose "zhong" with the caret at its end, then commit 中.
			objc.Send[objc.ID](view, objc.RegisterName("setMarkedText:selectedRange:replacementRange:"), str("zhong"), tiRange{5, 0}, tiRange{tiNotFound, 0})
			check(hasMarked(), "no marked text after setMarkedText")
			if got := objc.Send[tiRange](view, objc.RegisterName("markedRange")); got != (tiRange{0, 5}) {
				check(false, "markedRange = %+v, want {0 5}", got)
			}
			if got := objc.Send[tiRange](view, objc.RegisterName("selectedRange")); got != (tiRange{5, 0}) {
				check(false, "selectedRange while composing = %+v, want {5 0}", got)
			}
			objc.Send[objc.ID](view, objc.RegisterName("insertText:replacementRange:"), str("中"), tiRange{tiNotFound, 0})
			check(!hasMarked(), "marked text left after commit")
			if got := objc.Send[tiRange](view, objc.RegisterName("markedRange")); got.Location != tiNotFound {
				check(false, "markedRange after commit = %+v, want NSNotFound", got)
			}
			// A surrogate pair counts two UTF-16 units.
			objc.Send[objc.ID](view, objc.RegisterName("insertText:replacementRange:"), str("😀"), tiRange{tiNotFound, 0})
			if got := objc.Send[tiRange](view, objc.RegisterName("selectedRange")); got != (tiRange{3, 0}) {
				check(false, "selectedRange after emoji = %+v, want {3 0}", got)
			}
			// Cancel a composition: the marked text goes, the text stays.
			objc.Send[objc.ID](view, objc.RegisterName("setMarkedText:selectedRange:replacementRange:"), str("ni"), tiRange{2, 0}, tiRange{tiNotFound, 0})
			objc.Send[objc.ID](view, objc.RegisterName("setMarkedText:selectedRange:replacementRange:"), str(""), tiRange{0, 0}, tiRange{tiNotFound, 0})
			objc.Send[objc.ID](view, objc.RegisterName("unmarkText"))
			check(!hasMarked(), "marked text left after cancel")
		})
		waitText("中😀")
		w.Run(func() {
			pool := objc.ID(objc.GetClass("NSAutoreleasePool")).Send(objc.RegisterName("new"))
			defer pool.Send(objc.RegisterName("drain"))
			actual := new(tiRange)
			sub := objc.Send[objc.ID](view, objc.RegisterName("attributedSubstringForProposedRange:actualRange:"), tiRange{0, 3}, actual)
			got := objc.Send[objc.ID](sub, objc.RegisterName("string"))
			if s := nsDescription(got); s != "中😀" {
				check(false, "attributedSubstring = %q, want 中😀", s)
			}
			check(*actual == tiRange{0, 3}, "substring actualRange = %+v, want {0 3}", *actual)
			// The candidate window goes at the caret, in screen coordinates
			// inside the window.
			r := objc.Send[tiRect](view, objc.RegisterName("firstRectForCharacterRange:actualRange:"), tiRange{3, 0}, actual)
			frame := objc.Send[tiRect](view.Send(objc.RegisterName("window")), objc.RegisterName("frame"))
			check(r.Height > 0 && r.X > frame.X && r.X < frame.X+frame.Width && r.Y > frame.Y && r.Y < frame.Y+frame.Height,
				"caret rect %+v outside window frame %+v", r, frame)
			check(*actual == tiRange{3, 0}, "rect actualRange = %+v, want {3 0}", *actual)
			// NSNotFound: the view does not map points to characters.
			idx := objc.Send[uintptr](view, objc.RegisterName("characterIndexForPoint:"), struct{ X, Y float64 }{r.X, r.Y})
			check(idx == tiNotFound, "characterIndexForPoint = %d", idx)
		})
		if len(errs) > 0 {
			for _, e := range errs {
				fmt.Fprintln(os.Stderr, e)
			}
			os.Exit(1)
		}
		fmt.Println("PASS: marked text, commit, cancel, UTF-16 ranges and caret rectangle through NSTextInputClient")
		w.Perform(system.ActionClose)
	}()
	Main()
}

// nsDescription returns an NSString's UTF-8 contents.
func nsDescription(s objc.ID) string {
	p := objc.Send[*byte](s, objc.RegisterName("UTF8String"))
	if p == nil {
		return ""
	}
	var b []byte
	for q := p; *q != 0; q = (*byte)(unsafe.Add(unsafe.Pointer(q), 1)) {
		b = append(b, *q)
	}
	return string(b)
}
