package main

import (
	"flag"
	"fmt"
	"path/filepath"
)

// runProject runs the project's main package; arguments after -- go to
// the app.
func (c *cli) runProject(args []string) error {
	fs := flag.NewFlagSet("run", flag.ContinueOnError)
	fs.SetOutput(c.errw)
	fs.Usage = func() { fmt.Fprintln(c.errw, "Usage: keel run [-- app args]") }
	if err := fs.Parse(args); err != nil {
		return err
	}
	dir := c.wd()
	cfg, err := loadConfig(dir)
	if err != nil {
		return err
	}
	goArgs := append([]string{"run", "-ldflags", appIDFlag(cfg), "./" + filepath.ToSlash(filepath.Clean(cfg.Main))}, fs.Args()...)
	return c.command(dir, nil, "go", goArgs...)
}

// appIDFlag gives the binary the app ID Gio reports to Linux desktops
// (the Wayland app_id and X11 WM_CLASS), so a .desktop file matches it.
func appIDFlag(cfg *Config) string { return "-X gioui.org/app.ID=" + cfg.AppID }
