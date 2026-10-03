//go:build !((linux && !android) || freebsd || openbsd || netbsd)

package window

// clientDecorations: macOS, Windows and mobile platforms draw their own.
var clientDecorations = false
