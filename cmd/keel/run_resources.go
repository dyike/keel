package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/tc-hib/winres"
)

func runLinkFlags(cfg *Config, platform string) string {
	flags := appIDFlag(cfg)
	if platform == "windows" {
		flags += " -H=windowsgui"
	}
	return flags
}

// Development executables need the same DPI awareness and native theme
// activation as release builds. The Go linker reads resources from the main
// package, so keep this object there only while compiling, then remove it.
func prepareRunResources(dir string, cfg *Config, platform, arch string) (func(), error) {
	noop := func() {}
	if platform != "windows" {
		return noop, nil
	}
	mainDir := filepath.Join(dir, filepath.FromSlash(cfg.Main))
	others, err := filepath.Glob(filepath.Join(mainDir, "*.syso"))
	if err != nil {
		return noop, err
	}
	if len(others) > 0 {
		return noop, fmt.Errorf("%w: %s", errOtherSyso, others[0])
	}
	// Projects created before icon support still get a manifest and version.
	var icons *iconSet
	if cfg.Icons[platform] != "" {
		icons = &iconSet{dir: dir, cfg: cfg}
	} else if _, err := os.Stat(filepath.Join(dir, cfg.Icon)); cfg.Icon != "" && !(os.IsNotExist(err) && cfg.Icon == "appicon.png") {
		var iconErr error
		icons, iconErr = readIcons(dir, cfg)
		if iconErr != nil {
			return noop, iconErr
		}
	}
	rs, err := windowsResources(cfg, icons)
	if err != nil {
		return noop, err
	}
	var data bytes.Buffer
	if err := rs.WriteObject(&data, winres.Arch(arch)); err != nil {
		return noop, err
	}
	path := filepath.Join(mainDir, "zz_keel_run_windows_"+arch+".syso")
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return noop, err
	}
	cleanup := func() { os.Remove(path) }
	_, writeErr := file.Write(data.Bytes())
	closeErr := file.Close()
	if writeErr != nil {
		cleanup()
		return noop, writeErr
	}
	if closeErr != nil {
		cleanup()
		return noop, closeErr
	}
	return cleanup, nil
}
