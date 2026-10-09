//go:build !(darwin && !ios) && !windows && !(linux && !android)

package window

import "image"

func platformSetIcon(art image.Image, finished bool) {}
func iconWindowEvent(w *Window, e any)               {}
