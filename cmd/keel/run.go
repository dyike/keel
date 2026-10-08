package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
)

// runProject runs the project's main package; arguments after -- go to
// the app.
func (c *cli) runProject(args []string) error {
	fs := flag.NewFlagSet("run", flag.ContinueOnError)
	fs.SetOutput(c.errw)
	target := fs.String("target", "desktop", "desktop, ios or android")
	watch := fs.Bool("watch", true, "desktop: rebuild and restart on file changes")
	simulator := fs.String("simulator", "", "ios: simulator UDID (default: the sole booted device, otherwise the newest available iPhone)")
	serial := fs.String("serial", "", "android: adb device serial (required when multiple devices are connected)")
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
	if *serial != "" && *target != "android" {
		return fmt.Errorf("-serial requires -target android")
	}
	if *target == "android" {
		if *simulator != "" {
			return fmt.Errorf("-simulator requires -target ios; use -serial for Android")
		}
		return c.runAndroid(dir, cfg, *serial, fs.Args())
	}
	if *target == "ios" {
		return c.runIOS(dir, cfg, *simulator, fs.Args())
	}
	if *simulator != "" || (*target != "desktop" && *target != runtime.GOOS) {
		return fmt.Errorf("run target %q: use desktop, ios or android; -simulator requires ios", *target)
	}
	goArgs := append([]string{"run", "-ldflags", appIDFlag(cfg), "./" + filepath.ToSlash(filepath.Clean(cfg.Main))}, fs.Args()...)
	if c.dryRun {
		if runtime.GOOS == "darwin" {
			bundle := filepath.Join("<temporary>", cfg.Binary+".app")
			binary := filepath.Join(bundle, "Contents", "MacOS", cfg.Binary)
			if err := c.command(dir, nil, "go", "build", "-o", binary, "-ldflags", appIDFlag(cfg), "./"+filepath.ToSlash(filepath.Clean(cfg.Main))); err != nil {
				return err
			}
			fmt.Fprintln(c.out, "write development bundle metadata and icon to", bundle)
			return c.command(dir, nil, binary, fs.Args()...)
		}
		return c.command(dir, nil, "go", goArgs...)
	}
	if *watch {
		ctx, stop := signal.NotifyContext(context.Background(), runStopSignals()...)
		defer stop()
		return c.watchProject(ctx, dir, fs.Args())
	}
	if runtime.GOOS == "darwin" {
		// go run launches a bare temporary binary. Build once inside an app
		// bundle so non-watching runs have the same Dock identity as watch.
		ctx, stop := signal.NotifyContext(context.Background(), runStopSignals()...)
		defer stop()
		tmp, err := os.MkdirTemp("", "keel-run-")
		if err != nil {
			return err
		}
		defer os.RemoveAll(tmp)
		result := c.compileRun(ctx, dir, tmp, 1)
		defer result.cleanup()
		if result.err != nil {
			return result.err
		}
		return c.runOnce(ctx, dir, result, fs.Args())
	}
	icon, cleanup, err := prepareRunIcon(dir, cfg, runtime.GOOS)
	if err != nil {
		return err
	}
	defer cleanup()
	return c.command(dir, []string{"KEEL_RUN_ICON=" + icon, "KEEL_RUN_WATCH=0"}, "go", goArgs...)
}

func (c *cli) runOnce(ctx context.Context, dir string, build runBuild, args []string) error {
	cmd := exec.Command(build.binary, args...)
	cmd.Dir = dir
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, c.out, c.errw
	cmd.Env = append(os.Environ(), "KEEL_RUN_ICON="+build.icon, "KEEL_RUN_WATCH=0")
	configureRunProcess(cmd)
	if err := cmd.Start(); err != nil {
		return err
	}
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		stopRunProcess(cmd, done)
		return nil
	}
}

// appIDFlag gives the binary the app ID Gio reports to Linux desktops
// (the Wayland app_id and X11 WM_CLASS), so a .desktop file matches it.
func appIDFlag(cfg *Config) string { return "-X gioui.org/app.ID=" + cfg.AppID }

// prepareRunIcon passes a finished platform icon to the development
// executable. Keep it alive until the child exits.
func prepareRunIcon(dir string, cfg *Config, platform string) (string, func(), error) {
	noop := func() {}
	source := cfg.Icon
	if override := cfg.Icons[platform]; override != "" {
		source = override
	}
	if source == "" {
		return "", noop, nil
	}
	if _, err := os.Stat(filepath.Join(dir, source)); os.IsNotExist(err) && source == "appicon.png" && cfg.Icons[platform] == "" {
		// Existing projects without artwork still run with the system icon.
		return "", noop, nil
	}
	// An override is complete artwork; it need not have a base icon.
	var set *iconSet
	var err error
	if cfg.Icons[platform] != "" {
		set = &iconSet{dir: dir, cfg: cfg}
	} else {
		set, err = readIcons(dir, cfg)
		if err != nil {
			return "", noop, err
		}
	}
	size := 256
	if platform == "darwin" {
		size = 1024
	}
	img, err := set.icon(platform, size)
	if err != nil {
		return "", noop, err
	}
	tmp, err := os.MkdirTemp("", "keel-run-icon-")
	if err != nil {
		return "", noop, err
	}
	cleanup := func() { os.RemoveAll(tmp) }
	path := filepath.Join(tmp, "icon.png")
	if err := writePNG(path, img); err != nil {
		cleanup()
		return "", noop, err
	}
	return path, cleanup, nil
}
