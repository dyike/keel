//go:build !js || !keel_gallery_small

package main

// The category builder omits this optional provider only from browser bundles
// without code/Markdown examples. Desktop and complete galleries retain it.
import _ "github.com/dyike/keel/ui/highlight/golang"
