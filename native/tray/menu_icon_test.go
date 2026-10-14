package tray

import (
	"bytes"
	"errors"
	"image"
	"image/color"
	"image/png"
	"testing"

	"github.com/dyike/keel/native"
	"github.com/dyike/keel/native/internal/sys"
)

func TestMenuIconSnapshotAndValidation(t *testing.T) {
	art := image.NewNRGBA(image.Rect(4, 5, 20, 21))
	art.SetNRGBA(4, 5, color.NRGBA{R: 200, A: 255})
	tray, backend := testTray(t, Options{Menu: []MenuItem{{Title: "Action", Icon: art, IconTemplate: true}}})
	art.SetNRGBA(4, 5, color.NRGBA{B: 255, A: 255})
	item := backend.options.Menu[0]
	decoded, err := png.Decode(bytes.NewReader(item.PNG))
	if err != nil || color.NRGBAModel.Convert(decoded.At(0, 0)).(color.NRGBA).R != 200 {
		t.Fatal("image aliases caller or lost nonzero origin", err)
	}
	if !item.Template {
		t.Fatal("template flag lost")
	}
	for _, menu := range [][]MenuItem{{{Separator: true, Icon: art}}, {{Title: "Invalid", IconTemplate: true}}, {{Title: "Invalid", Icon: image.NewNRGBA(image.Rect(0, 0, 257, 1))}}} {
		if err := tray.SetMenu(menu); err == nil {
			t.Fatal("invalid icon accepted")
		}
		if !bytes.Equal(backend.options.Menu[0].PNG, item.PNG) {
			t.Fatal("failed update changed image")
		}
	}
	if err := tray.SetMenuAnimation(MenuAnimation(255)); !errors.Is(err, native.ErrInvalidArgument) {
		t.Fatal(err)
	}
	if !Capabilities().DisableOpeningAnimation {
		if err := tray.SetMenuAnimation(MenuAnimationNone); !errors.Is(err, native.ErrUnsupported) {
			t.Fatal(err)
		}
	}
	if err := tray.CancelMenuTracking(false); !errors.Is(err, native.ErrUnsupported) {
		t.Fatal("fake backend claimed cancellation", err)
	}
}

type animationTray struct {
	fakeTray
	animation           uint8
	animated            bool
	sets, cancellations int
}

func (f *animationTray) SetMenuAnimation(animation uint8) error {
	if f.fail != nil {
		return f.fail
	}
	f.animation = animation
	f.sets++
	return nil
}
func (f *animationTray) CancelMenuTracking(animated bool) error {
	f.animated = animated
	f.cancellations++
	return f.fail
}

func TestMenuAnimationPolicyAndCancellation(t *testing.T) {
	backend := new(animationTray)
	tray, err := newTray(Options{Icon: image.NewNRGBA(image.Rect(0, 0, 16, 16))}, func(options sys.TrayOptions) (sys.TrayBackend, error) {
		backend.options = options
		return backend, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { tray.Close() })
	for _, animated := range []bool{false, true} {
		if err := tray.CancelMenuTracking(animated); err != nil || backend.animated != animated {
			t.Fatal("cancellation argument lost", err)
		}
	}
	if backend.cancellations != 2 {
		t.Fatal("cancellation routing")
	}
	if err := tray.SetMenuAnimation(MenuAnimationSystem); err != nil || backend.sets != 1 {
		t.Fatal("policy setter bypassed", err)
	}
	if Capabilities().DisableOpeningAnimation {
		if err := tray.SetMenuAnimation(MenuAnimationNone); err != nil || backend.animation != 1 {
			t.Fatal("disable animation", err)
		}
	}
	old := tray.options.MenuAnimation
	backend.fail = native.ErrFailed
	if err := tray.SetMenuAnimation(MenuAnimationSystem); !errors.Is(err, native.ErrFailed) || tray.options.MenuAnimation != old {
		t.Fatal("failed policy update was not atomic", err)
	}
	backend.fail = nil
	if err := tray.Close(); err != nil {
		t.Fatal(err)
	}
	if err := tray.CancelMenuTracking(false); !errors.Is(err, native.ErrFailed) || backend.cancellations != 2 {
		t.Fatal("closed tray dispatched cancellation", err)
	}
}
