// SPDX-License-Identifier: Unlicense OR MIT

//go:build darwin && !ios && !nometal

package app

import (
	stdcontext "context"
	"fmt"
	"image/color"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/dyike/keel/gio/op"
	"github.com/dyike/keel/gio/op/paint"
	"github.com/dyike/keel/gio/unit"
	"github.com/ebitengine/purego/objc"
)

// Exercise applicationDidFinishLaunching in a real bundle: changing the
// policy after a window is shown can make accessory launchers lose focus.
func TestMacOSBundleActivationPolicy(t *testing.T) {
	if os.Getenv("KEEL_DESKTOP") != "1" {
		t.Skip("set KEEL_DESKTOP=1 for native bundle activation regression")
	}
	binary, err := os.ReadFile(os.Args[0])
	if err != nil {
		t.Fatal(err)
	}
	for _, accessory := range []bool{false, true} {
		t.Run(strconv.FormatBool(accessory), func(t *testing.T) {
			bundle := filepath.Join(t.TempDir(), "Activation.app")
			dir := filepath.Join(bundle, "Contents", "MacOS")
			if err := os.MkdirAll(dir, 0o755); err != nil {
				t.Fatal(err)
			}
			helper := filepath.Join(dir, "activation")
			if err := os.WriteFile(helper, binary, 0o755); err != nil {
				t.Fatal(err)
			}
			plist := fmt.Sprintf(`<?xml version="1.0"?><plist version="1.0"><dict>
<key>CFBundleExecutable</key><string>activation</string>
<key>CFBundleIdentifier</key><string>com.dyike.keel.activation-%v</string>
<key>CFBundlePackageType</key><string>APPL</string>
<key>LSUIElement</key><%v/>
</dict></plist>`, accessory, accessory)
			if err := os.WriteFile(filepath.Join(bundle, "Contents", "Info.plist"), []byte(plist), 0o644); err != nil {
				t.Fatal(err)
			}
			if out, err := exec.Command("codesign", "--force", "--sign", "-", bundle).CombinedOutput(); err != nil {
				t.Fatalf("codesign: %v\n%s", err, out)
			}
			ctx, cancel := stdcontext.WithTimeout(stdcontext.Background(), 15*time.Second)
			defer cancel()
			cmd := exec.CommandContext(ctx, helper)
			want := "0" // NSApplicationActivationPolicyRegular
			if accessory {
				want = "1" // NSApplicationActivationPolicyAccessory
			}
			cmd.Env = append(os.Environ(), "GIO_ACTIVATION_POLICY_HELPER="+want)
			if out, err := cmd.CombinedOutput(); err != nil {
				t.Fatalf("activation helper: %v\n%s", err, out)
			}
		})
	}
}

func activationPolicyWindowHelper() {
	want, _ := strconv.Atoi(os.Getenv("GIO_ACTIVATION_POLICY_HELPER"))
	w := new(Window)
	w.Option(Title("Activation policy regression"), Size(unit.Dp(320), unit.Dp(240)), TopMost(true))
	frames := make(chan struct{}, 1)
	go func() {
		var ops op.Ops
		for {
			switch e := w.Event().(type) {
			case DestroyEvent:
				os.Exit(1)
			case FrameEvent:
				ops.Reset()
				paint.Fill(&ops, color.NRGBA{R: 30, G: 40, B: 50, A: 255})
				e.Frame(&ops)
				select {
				case frames <- struct{}{}:
				default:
				}
			}
		}
	}()
	go func() {
		<-frames
		var got int
		w.Run(func() {
			app := objc.ID(objc.GetClass("NSApplication")).Send(objc.RegisterName("sharedApplication"))
			got = int(app.Send(objc.RegisterName("activationPolicy")))
		})
		if got != want {
			fmt.Fprintf(os.Stderr, "activation policy=%d, want %d\n", got, want)
			os.Exit(1)
		}
		os.Exit(0)
	}()
	Main()
}
