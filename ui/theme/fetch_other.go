//go:build !js

package theme

import "errors"

// FetchFonts is for WebAssembly builds, where it downloads font files with
// the browser. Elsewhere it returns an error: read the files and LoadFonts.
func FetchFonts(urls ...string) error {
	return errors.New("theme.FetchFonts works in a browser; use LoadFonts")
}
