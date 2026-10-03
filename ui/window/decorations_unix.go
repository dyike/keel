//go:build (linux && !android) || freebsd || openbsd || netbsd

package window

// clientDecorations: on Wayland the compositor may leave the title bar to us.
var clientDecorations = true
