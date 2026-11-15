//go:build !keel_lazy_highlight

package golang

import "github.com/dyike/keel/ui/core"

func init() {
	core.SetHighlighter(Chroma{})
}
