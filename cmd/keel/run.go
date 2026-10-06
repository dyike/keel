package main

import (
	"flag"
	"fmt"
	"path/filepath"
	"runtime"
)

// runProject runs the project's main package; arguments after -- go to
// the app.
func (c *cli) runProject(args []string) error {
	fs := flag.NewFlagSet("run", flag.ContinueOnError)
	fs.SetOutput(c.errw)
	target := fs.String("target", "desktop", "desktop or ios")
	simulator := fs.String("simulator", "", "ios: simulator UDID (default: the sole booted device, otherwise the newest available iPhone)")
	fs.BoolVar(&c.dryRun, "n", false, "print the commands instead of running them")
	fs.Usage = func() {
		fmt.Fprintln(c.errw, "Usage: keel run [flags] [-- app args]")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	dir := c.wd()
	cfg, err := loadConfig(dir)
	if err != nil {
		return err
	}
	if *target == "ios" {
		return c.runIOS(dir, cfg, *simulator, fs.Args())
	}
	if *simulator != "" || (*target != "desktop" && *target != runtime.GOOS) {
		return fmt.Errorf("run target %q: use desktop or ios; -simulator requires ios", *target)
	}
	goArgs := append([]string{"run", "-ldflags", appIDFlag(cfg), "./" + filepath.ToSlash(filepath.Clean(cfg.Main))}, fs.Args()...)
	return c.command(dir, nil, "go", goArgs...)
}

// appIDFlag gives the binary the app ID Gio reports to Linux desktops
// (the Wayland app_id and X11 WM_CLASS), so a .desktop file matches it.
func appIDFlag(cfg *Config) string { return "-X gioui.org/app.ID=" + cfg.AppID }
