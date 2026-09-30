// Inspect reads display metadata and permission status without starting a GUI.
package main

import (
	"fmt"
	"github.com/dyike/keel/permissions"
	"github.com/dyike/keel/screen"
)

func main() {
	displays, err := screen.New().Displays()
	fmt.Printf("displays: %+v; error: %v\n", displays, err)
	p := permissions.New()
	for _, kind := range []permissions.Kind{permissions.Accessibility, permissions.ScreenRecording, permissions.InputMonitoring} {
		s, err := p.Check(kind)
		fmt.Printf("permission %d: %d; error: %v\n", kind, s, err)
	}
}
