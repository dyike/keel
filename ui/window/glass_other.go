//go:build !(darwin && !ios && cgo && !nometal)

package window

func platformGlassSupported() bool       { return false }
func platformLiquidGlassSupported() bool { return false }
func platformGlassEvent(*Window, any)    {}
