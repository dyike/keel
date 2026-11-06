//go:build keel_legacy_svg

package oksvg

import (
	"golang.org/x/net/html/charset"
	"io"
)

func svgCharsetReader(label string, input io.Reader) (io.Reader, error) {
	return charset.NewReaderLabel(label, input)
}
