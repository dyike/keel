// SPDX-License-Identifier: Unlicense OR MIT

//go:build darwin && ios && !nometal

package app

// metalIdleTrims is false on iOS: an idle window only releases its scratch
// textures and spare buffers, as it did before the debounce.
const metalIdleTrims = false

func trimMetalLayer(layer uintptr) {}
