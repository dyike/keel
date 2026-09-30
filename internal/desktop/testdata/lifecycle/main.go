//go:build darwin && cgo

// Native regression probe: exercise AppKit application actions while idle.
package main

/*
#cgo CFLAGS: -x objective-c -fobjc-arc
#cgo LDFLAGS: -framework AppKit -framework Carbon
void startLifecycleProbe(void);
int lifecycleProbePassed(void);
*/
import "C"

import (
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/dyike/keel"
	"github.com/dyike/keel/ui"
)

func main() {
	// Keep results available when Launch Services launches the test bundle.
	if executable, err := os.Executable(); err == nil && filepath.Base(executable) == "KeelLifecycleProbe" {
		file, err := os.Create(filepath.Join(filepath.Dir(executable), "..", "lifecycle.log"))
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()
		log.SetOutput(file)
	}
	go func() { time.Sleep(15 * time.Second); log.Print("FAIL: native lifecycle timed out"); os.Exit(1) }()
	kit := keel.New()
	w, err := kit.Window.New(keel.WindowOptions{Title: "Keel lifecycle probe", UI: ui.NewPage("Main", ui.Text("Native lifecycle regression"))})
	if err != nil {
		log.Fatal(err)
	}
	_, err = kit.Window.New(keel.WindowOptions{Title: "Keel lifecycle second window", UI: ui.NewPage("Second", ui.Text("Independent window"))})
	if err != nil {
		log.Fatal(err)
	}
	_, err = kit.Window.New(keel.WindowOptions{Title: "Hidden utility", Hidden: true, UI: ui.NewPage("Utility", ui.Text("Keep hidden"))})
	if err != nil {
		log.Fatal(err)
	}
	if err := w.OnReady(func() { C.startLifecycleProbe() }); err != nil {
		log.Fatal(err)
	}
	if err := kit.Run(); err != nil {
		log.Fatal(err)
	}
	if C.lifecycleProbePassed() == 0 {
		log.Fatal("FAIL: native application lifecycle")
	}
	if w.Host().Ctx().Err() == nil {
		log.Fatal("FAIL: window cleanup skipped")
	}
	for _, host := range kit.App.Windows() {
		if host.Ctx().Err() == nil {
			log.Fatal("FAIL: another window remained alive after quit")
		}
	}
	log.Print("PASS: hide, Apple Event reopen/quit while idle, hidden utility preserved, window cleanup")
}
