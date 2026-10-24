// SPDX-License-Identifier: Unlicense OR MIT

//go:build linux && !android && !novulkan

package headless

import "testing"

// TestRenderPixelsVulkan renders through the Vulkan fallback context: the
// EGL context Linux selects first is skipped. Keel patch: it covers the
// Vulkan binding under either cgo or purego.
func TestRenderPixelsVulkan(t *testing.T) {
	primary := newContextPrimary
	newContextPrimary = nil
	defer func() { newContextPrimary = primary }()
	checkRenderPixels(t)
}
