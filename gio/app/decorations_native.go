// SPDX-License-Identifier: Unlicense OR MIT

//go:build (darwin && !ios) || windows

package app

import (
	"image"

	"github.com/dyike/keel/gio/op"
	"github.com/dyike/keel/gio/unit"
)

// AppKit and Win32 implement Decorated directly. They never need Gio's
// client-side title bar, including when a window is explicitly undecorated.
type fallbackDecorations struct{}

func (*Window) initDecorations() unit.Dp                 { return 0 }
func (*Window) updateDecorationMode(WindowMode)          {}
func (*Window) fallbackDecorate() bool                   { return false }
func (*Window) decorate(FrameEvent, *op.Ops) image.Point { return image.Point{} }
