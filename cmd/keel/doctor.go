package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
)

// doctor reports what each target needs on this machine.
func (c *cli) doctor(args []string) error {
	flags := flag.NewFlagSet("doctor", flag.ContinueOnError)
	flags.SetOutput(c.errw)
	target := flags.String("target", runtime.GOOS, "target to check; ios checks Xcode, android checks SDK, NDK, Java and adb")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if *target != runtime.GOOS && *target != "ios" && *target != "android" {
		return fmt.Errorf("cannot check target %q on this machine", *target)
	}
	ok := true
	check := func(name string, pass bool, detail string) {
		mark := "ok  "
		if !pass {
			mark, ok = "fix ", false
		}
		fmt.Fprintf(c.out, "%s %s", mark, name)
		if detail != "" {
			fmt.Fprintf(c.out, ": %s", detail)
		}
		fmt.Fprintln(c.out)
	}
	goVersion, err := exec.Command("go", "env", "GOVERSION").Output()
	v := strings.TrimSpace(string(goVersion))
	check("Go "+minGoVersion+" or newer", err == nil && goAtLeast(v, minGoVersion), v)
	if *target == "android" {
		c.doctorAndroid(check)
		if !ok {
			return fmt.Errorf("fix the items above")
		}
		return nil
	}
	if *target == "ios" {
		check("macOS host", runtime.GOOS == "darwin", "iOS builds require full Xcode on a Mac")
		if runtime.GOOS == "darwin" {
			for _, sdk := range []string{"iphonesimulator", "iphoneos"} {
				location, err := toolOutput(c.wd(), "xcrun", "--sdk", sdk, "--show-sdk-path")
				check(sdk+" SDK", err == nil, hint(err, "install full Xcode and select it with xcode-select")+location)
			}
			data, err := toolOutput(c.wd(), "xcrun", "simctl", "list", "runtimes", "--json")
			var inventory struct {
				Runtimes []struct {
					Identifier  string
					IsAvailable bool
				}
			}
			if err == nil {
				err = json.Unmarshal([]byte(data), &inventory)
			}
			available := false
			for _, item := range inventory.Runtimes {
				available = available || item.IsAvailable && strings.Contains(item.Identifier, ".iOS-")
			}
			detail := ""
			if err != nil || !available {
				detail = "install runtimes in Xcode Settings > Components"
			}
			check("iOS simulator runtime", err == nil && available, detail)
			developer, err := toolOutput(c.wd(), "xcode-select", "-p")
			if err == nil {
				_, err = iosSimulatorApp(developer)
			}
			check("simulator viewer", err == nil, hint(err, "install full Xcode and select it with xcode-select"))
		}
		if !ok {
			return fmt.Errorf("fix the items above")
		}
		return nil
	}
	switch runtime.GOOS {
	case "darwin":
		_, err := exec.Command("xcode-select", "-p").Output()
		check("Xcode command line tools (cgo, iconutil)", err == nil, hint(err, "xcode-select --install"))
		_, err = exec.LookPath("iconutil")
		check("iconutil for .app icons", err == nil, "")
	case "linux":
		for _, pkg := range []string{"wayland-client", "x11", "xkbcommon", "egl"} {
			err := exec.Command("pkg-config", "--exists", pkg).Run()
			check("pkg-config "+pkg, err == nil, hint(err, "install the -dev / -devel package for "+pkg))
		}
	case "windows":
		check("Windows builds need no cgo", true, "")
	}
	fmt.Fprintln(c.out, "\nTargets from this machine:")
	fmt.Fprintln(c.out, "  windows, js  any OS (no cgo)")
	switch runtime.GOOS {
	case "darwin":
		fmt.Fprintln(c.out, "  darwin       here")
		fmt.Fprintln(c.out, "  ios          full Xcode required; check with keel doctor -target ios")
	case "linux":
		fmt.Fprintln(c.out, "  linux        here")
	}
	fmt.Fprintln(c.out, "  android      SDK, NDK and Java required; check with keel doctor -target android")
	if !ok {
		return fmt.Errorf("fix the items above")
	}
	return nil
}

func hint(err error, fix string) string {
	if err == nil {
		return ""
	}
	return fix
}

// goAtLeast reports whether toolchain v ("go1.26.1", "go1.27rc1") is at
// least release want ("1.26.1"). A prerelease counts as below its first
// release: go1.27rc1 is newer than 1.26.9 but older than 1.27.0.
func goAtLeast(v, want string) bool {
	got, pre, ok := goRelease(strings.TrimPrefix(v, "go"))
	min, _, ok2 := goRelease(want)
	if !ok || !ok2 {
		return false
	}
	for i := range got {
		if got[i] != min[i] {
			return got[i] > min[i]
		}
	}
	return !pre
}

// goRelease splits "1.26.1" or "1.27rc1" into numbers and a prerelease flag.
func goRelease(v string) (n [3]int, pre bool, ok bool) {
	parts := strings.SplitN(v, ".", 3)
	if len(parts) < 2 {
		return n, false, false
	}
	for i, p := range parts {
		digits := p
		if j := strings.IndexFunc(p, func(r rune) bool { return r < '0' || r > '9' }); j >= 0 {
			digits, pre = p[:j], true
		}
		x, err := strconv.Atoi(digits)
		if err != nil {
			return n, false, false
		}
		n[i] = x
	}
	return n, pre, true
}
