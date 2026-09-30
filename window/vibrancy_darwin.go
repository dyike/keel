//go:build darwin && cgo

package window

import "github.com/go-gui-org/go-gui/gui"

func setGUIVibrancy(d *guiDriver, v Vibrancy) error {
	material := gui.VibrancyMaterial(0)
	switch v {
	case VibrancySidebar:
		material = gui.VibrancySidebar
	case VibrancyHUD:
		material = gui.VibrancyMaterial(3)
	}
	return d.enqueue(func(w *gui.Window) { w.SetWindowVibrancy(material) })
}
