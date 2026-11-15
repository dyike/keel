//go:build !js || !keel_gallery_small

package main

import (
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/highlight/golang"
)

// The category builder omits this optional provider from browser bundles
// without code/Markdown examples. Desktop and code sections enable it.
func enableHighlight() { core.SetHighlighter(golang.Chroma{}) }
