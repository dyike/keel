package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strconv"
	"strings"
)

type iosSimulator struct {
	Name        string `json:"name"`
	UDID        string `json:"udid"`
	State       string `json:"state"`
	IsAvailable bool   `json:"isAvailable"`
	runtime     string
}

func selectIOSSimulator(data []byte, requested string) (iosSimulator, error) {
	var inventory struct {
		Devices map[string][]iosSimulator `json:"devices"`
	}
	if err := json.Unmarshal(data, &inventory); err != nil {
		return iosSimulator{}, fmt.Errorf("simulator inventory: %w", err)
	}
	var candidates, booted []iosSimulator
	for runtime, devices := range inventory.Devices {
		if !strings.Contains(runtime, ".iOS-") {
			continue
		}
		for _, device := range devices {
			if !device.IsAvailable {
				continue
			}
			device.runtime = runtime
			if requested != "" && strings.EqualFold(device.UDID, requested) {
				return device, nil
			}
			candidates = append(candidates, device)
			if device.State == "Booted" {
				booted = append(booted, device)
			}
		}
	}
	if requested != "" {
		return iosSimulator{}, fmt.Errorf("iOS simulator %q is unavailable; see xcrun simctl list devices available", requested)
	}
	if len(booted) > 1 {
		return iosSimulator{}, errors.New("multiple iOS simulators are booted; select one with -simulator <UDID>")
	}
	if len(booted) == 1 {
		return booted[0], nil
	}
	if len(candidates) == 0 {
		return iosSimulator{}, errors.New("no available iOS simulator; install an iOS runtime in Xcode Settings > Components")
	}
	slices.SortFunc(candidates, func(a, b iosSimulator) int {
		if phoneA, phoneB := strings.HasPrefix(a.Name, "iPhone"), strings.HasPrefix(b.Name, "iPhone"); phoneA != phoneB {
			if phoneA {
				return -1
			}
			return 1
		}
		if order := compareIOSVersions(a.runtime, b.runtime); order != 0 {
			return -order
		}
		return strings.Compare(a.UDID, b.UDID)
	})
	return candidates[0], nil
}

func compareIOSVersions(a, b string) int {
	parse := func(value string) []string {
		if _, version, found := strings.Cut(value, ".iOS-"); found {
			value = strings.ReplaceAll(version, "-", ".")
		}
		return strings.Split(value, ".")
	}
	x, y := parse(a), parse(b)
	for i := 0; i < max(len(x), len(y)); i++ {
		var first, second int
		if i < len(x) {
			first, _ = strconv.Atoi(x[i])
		}
		if i < len(y) {
			second, _ = strconv.Atoi(y[i])
		}
		if first < second {
			return -1
		}
		if first > second {
			return 1
		}
	}
	return 0
}

func (c *cli) runIOS(dir string, cfg *Config, requested string, args []string) error {
	if runtime.GOOS != "darwin" && !c.dryRun {
		return errors.New("iOS simulators require a Mac with full Xcode")
	}
	device := iosSimulator{UDID: requested, State: "Shutdown"}
	developer := "<Xcode-developer-directory>"
	if c.dryRun {
		if device.UDID == "" {
			device.UDID = "<simulator-udid>"
		}
		c.command(dir, nil, "xcrun", "simctl", "list", "devices", "available", "--json")
	} else {
		inventory, err := toolOutput(dir, "xcrun", "simctl", "list", "devices", "available", "--json")
		if err != nil {
			return err
		}
		device, err = selectIOSSimulator([]byte(inventory), requested)
		if err != nil {
			return err
		}
		if compareIOSVersions(device.runtime, cfg.iosMinimumVersion()) < 0 {
			return fmt.Errorf("%s is older than ios.minimum_version %s; select a newer simulator", device.Name, cfg.iosMinimumVersion())
		}
		developer, err = toolOutput(dir, "xcode-select", "-p")
		if err != nil {
			return err
		}
		fmt.Fprintf(c.out, "Running on %s (%s).\n", device.Name, device.UDID)
	}
	icons, err := loadIcons(dir, cfg)
	if err != nil {
		return err
	}
	c.release = false // Development runs keep debug information.
	output := filepath.Join(dir, "dist")
	main := "./" + filepath.ToSlash(filepath.Clean(cfg.Main))
	if err := c.buildIOS(dir, cfg, icons, output, main, "", "", "", false); err != nil {
		return err
	}
	if device.State != "Booted" {
		if err := c.command(dir, nil, "xcrun", "simctl", "boot", device.UDID); err != nil {
			return err
		}
	}
	if err := c.command(dir, nil, "xcrun", "simctl", "bootstatus", device.UDID, "-b"); err != nil {
		return err
	}
	simulator := filepath.Join(developer, "Applications", "Simulator.app")
	if !c.dryRun {
		var err error
		simulator, err = iosSimulatorApp(developer)
		if err != nil {
			return err
		}
	}
	if err := c.command(dir, nil, "open", "-a", simulator); err != nil {
		return err
	}
	app := filepath.Join(output, "ios", cfg.Binary+".app")
	if err := c.command(dir, nil, "xcrun", "simctl", "install", device.UDID, app); err != nil {
		return err
	}
	launch := []string{"simctl", "launch", "--terminate-running-process", "--console-pty", device.UDID, cfg.AppID}
	return c.command(dir, nil, "xcrun", append(launch, args...)...)
}

// Recent Xcode versions may provide DeviceHub instead of Simulator.app.
func iosSimulatorApp(developer string) (string, error) {
	for _, candidate := range []string{
		filepath.Join(developer, "Applications", "Simulator.app"),
		filepath.Join(developer, "..", "Applications", "Simulator.app"),
		filepath.Join(developer, "..", "Applications", "DeviceHub.app"),
	} {
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return filepath.Clean(candidate), nil
		}
	}
	return "", fmt.Errorf("no simulator viewer (Simulator.app or DeviceHub.app) in the selected Xcode at %s", developer)
}
