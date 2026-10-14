package window

import (
	"bytes"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"math"
	"time"

	"github.com/dyike/keel/ui/core"
)

// MenuIcon is an immutable, reusable bitmap. Template icons use their alpha
// mask with the current text color on AppKit and Keel-rendered menus.
type MenuIcon struct {
	pixels   *image.NRGBA
	png      []byte
	template bool
}

// NewMenuIcon copies art; later changes to the caller's image have no effect.
// Icons are displayed at 16dp; supply higher-resolution artwork for high DPI.
func NewMenuIcon(art image.Image, template bool) (*MenuIcon, error) {
	if art == nil {
		return nil, fmt.Errorf("window: nil menu icon")
	}
	size := art.Bounds().Size()
	if size.X <= 0 || size.Y <= 0 || size.X > 256 || size.Y > 256 {
		return nil, fmt.Errorf("window: menu icon must be 1..256 pixels per axis")
	}
	copy := image.NewNRGBA(image.Rectangle{Max: size})
	draw.Draw(copy, copy.Bounds(), art, art.Bounds().Min, draw.Src)
	var data bytes.Buffer
	if err := png.Encode(&data, copy); err != nil {
		return nil, err
	}
	return &MenuIcon{pixels: copy, png: data.Bytes(), template: template}, nil
}

// Image returns a copy for applications that draw their own menu.
func (icon *MenuIcon) Image() image.Image {
	if icon == nil || icon.pixels == nil {
		return nil
	}
	copy := image.NewNRGBA(icon.pixels.Bounds())
	copy.Pix = append(copy.Pix[:0], icon.pixels.Pix...)
	return copy
}

func (icon *MenuIcon) Template() bool { return icon != nil && icon.template }

// MenuAnimation controls opening Keel-rendered window menus. The zero value
// keeps immediate opening. Native menu-bar opening is owned by the OS.
type MenuAnimation struct {
	Duration time.Duration
	Offset   float32 // downward opening displacement in dp, eased to zero
}

func (m *MenuBar) SetWindowAnimation(animation MenuAnimation) error {
	if animation.Duration < 0 || animation.Duration > 10*time.Second || animation.Offset < 0 ||
		animation.Offset > 128 || math.IsNaN(float64(animation.Offset)) || math.IsInf(float64(animation.Offset), 0) {
		return fmt.Errorf("window: invalid menu animation")
	}
	m.mu.Lock()
	m.animation = animation
	m.mu.Unlock()
	go core.Update(func() {})
	return nil
}

func (m *MenuBar) WindowAnimation() MenuAnimation {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.animation
}

// CancelTracking closes this installed menu in all windows. AppKit honors
// animated; Win32 dismisses immediately. It is safe from a UI callback.
func (m *MenuBar) CancelTracking(animated bool) {
	bar, generation := currentMenu()
	if bar != m {
		return
	}
	// Native tracking may suspend UI frames. Post cancellation directly to the
	// OS thread; never make it depend on draining the frame update queue.
	if !offScreen() {
		platformCancelMenuTracking(generation, animated)
	}
	go core.Update(func() {
		bar, current := currentMenu()
		if bar != m || current != generation {
			return
		}
		application.Lock()
		for w := range application.windows {
			if w.menuView.bar == m {
				w.menuView.reset()
			}
		}
		application.Unlock()
	})
}

// MenuFeatures describes the actual menu presentation selected for a window.
type MenuFeatures struct {
	Native, Icons, TemplateIcons, WindowAnimation, NativeDismissAnimation, CancelTracking bool
}

func (w *Window) MenuFeatures() MenuFeatures {
	if w.opts.MenuDisplay == MenuDisplayHidden {
		return MenuFeatures{}
	}
	if w.opts.MenuDisplay == MenuDisplayWindow || platformDrawApplicationMenu(w) {
		return MenuFeatures{Icons: true, TemplateIcons: true, WindowAnimation: true, CancelTracking: true}
	}
	if offScreen() {
		return MenuFeatures{}
	}
	return platformMenuFeatures()
}
