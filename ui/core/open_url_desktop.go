//go:build !js

package core

import (
	"fmt"
	"os/exec"
	"runtime"
)

func openURL(raw string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", raw)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", raw)
	case "linux", "freebsd", "openbsd":
		cmd = exec.Command("xdg-open", raw)
	default:
		return fmt.Errorf("open URL: unsupported platform %s", runtime.GOOS)
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	go func() { _ = cmd.Wait() }()
	return nil
}
