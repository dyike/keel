//go:build !keel_legacy_svg

package oksvg

import (
	"fmt"
	"io"
	"strings"
)

// encoding/xml handles UTF-8 itself. Accept its common alias too; all other
// encodings need the optional legacy decoder tables rather than silent fallback.
func svgCharsetReader(label string, input io.Reader) (io.Reader, error) {
	if strings.EqualFold(strings.TrimSpace(label), "utf8") {
		return input, nil
	}
	return nil, fmt.Errorf("unsupported SVG charset %q: convert to UTF-8 or build with -tags keel_legacy_svg", label)
}
