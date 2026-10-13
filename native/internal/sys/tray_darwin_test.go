//go:build darwin && !ios

package sys

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/png"
	"os"
	"os/exec"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"time"
)

// The helper owns the original OS main thread and a real AppKit event loop.
func TestMain(m *testing.M) {
	if os.Getenv("KEEL_TRAY_NATIVE_HELPER") == "1" {
		runtime.LockOSThread()
		runNativeTrayHelper()
		return
	}
	os.Exit(m.Run())
}

func TestRealTrayDarwin(t *testing.T) {
	if os.Getenv("KEEL_DESKTOP") != "1" {
		t.Skip("set KEEL_DESKTOP=1 to create a real status item")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, os.Args[0])
	cmd.Env = append(os.Environ(), "KEEL_TRAY_NATIVE_HELPER=1")
	out, err := cmd.CombinedOutput()
	if err != nil || !strings.Contains(string(out), "PASS") {
		t.Fatalf("native tray helper: %v\n%s", err, out)
	}
}

func runNativeTrayHelper() {
	check := func(ok bool, message string) {
		if !ok {
			fmt.Println("FAIL:", message)
			os.Exit(1)
		}
	}
	time.AfterFunc(10*time.Second, func() { fmt.Println("FAIL: native tray timeout"); os.Exit(1) })
	var data bytes.Buffer
	check(png.Encode(&data, image.NewNRGBA(image.Rect(0, 0, 32, 32))) == nil, "PNG")
	events := []uint32{}
	options := TrayOptions{PNG: data.Bytes(), Template: true, Tooltip: "Initial tooltip", Event: func(id uint32) { events = append(events, id) }, Menu: []TrayMenuItem{
		{ID: 1, Title: "Action", Checked: true},
		{ID: 2, Title: "Disabled", Disabled: true},
		{ID: 3, Title: "Submenu", Children: []TrayMenuItem{{ID: 4, Title: "Child"}}},
	}}
	backend, err := TrayOpen(options)
	check(err == nil, "create")
	tray := backend.(*darwinTray)
	button := send(tray.item, "button")
	check(sendBool(send(button, "image"), "isTemplate"), "template image")
	check(goString(send(button, "toolTip")) == "Initial tooltip", "tooltip")
	menu := send(tray.item, "menu")
	check(send(menu, "numberOfItems") == 3, "menu length")
	action := send(menu, "itemAtIndex:", 0)
	check(send(action, "state") == 1, "checked item")
	check(!sendBool(send(menu, "itemAtIndex:", 1), "isEnabled"), "disabled item")
	check(send(send(send(menu, "itemAtIndex:", 2), "submenu"), "numberOfItems") == 1, "submenu")
	// Exercise the registered Objective-C target/action callback and tag routing.
	send(tray.target, "keelTrayAction:", uintptr(action))
	send(tray.target, "keelTrayAction:", uintptr(button))
	check(reflect.DeepEqual(events, []uint32{1, 0}), "target/action routing")
	options.Tooltip = "Updated tooltip"
	options.Template = false
	options.Menu = []TrayMenuItem{{ID: 9, Title: "Updated item"}}
	check(tray.Update(options) == nil, "update enqueue")
	mainAsync(func() {
		check(goString(send(button, "toolTip")) == "Updated tooltip", "update applied")
		check(!sendBool(send(button, "image"), "isTemplate"), "template update")
		updated := send(send(tray.item, "menu"), "itemAtIndex:", 0)
		check(goString(send(updated, "title")) == "Updated item", "menu replacement")
		send(tray.target, "keelTrayAction:", uintptr(updated))
		check(reflect.DeepEqual(events, []uint32{1, 0, 9}), "replacement action")
		oldTarget := tray.target
		check(tray.Close() == nil, "close enqueue")
		mainAsync(func() {
			check(tray.item == 0 && tray.target == 0, "released objects")
			_, retained := darwinTrays.targets[oldTarget]
			check(!retained, "removed callback state")
			// Repeated creation reuses the registered Objective-C class/callback.
			for range 5 {
				next, err := TrayOpen(options)
				check(err == nil, "recreate")
				check(next.Close() == nil, "reclose")
			}
			mainAsync(func() {
				check(len(darwinTrays.targets) == 0, "retained recreated tray targets")
				fmt.Println("PASS")
				os.Exit(0)
			})
		})
	})
	TrayMain()
}
