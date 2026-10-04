package main

import (
	"fmt"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
)

// doctor reports what each target needs on this machine.
func (c *cli) doctor(args []string) error {
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
	check("Go 1.26 or newer", err == nil && goAtLeast(v, 1, 26), v)
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
	case "linux":
		fmt.Fprintln(c.out, "  linux        here")
	}
	fmt.Fprintln(c.out, "  The first build downloads the packager,", gogio)
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

// goAtLeast parses "go1.26.1" and compares it with major.minor.
func goAtLeast(v string, major, minor int) bool {
	parts := strings.SplitN(strings.TrimPrefix(v, "go"), ".", 3)
	if len(parts) < 2 {
		return false
	}
	ma, err1 := strconv.Atoi(parts[0])
	digits := parts[1] // "26", or "27rc1" for a prerelease
	if i := strings.IndexFunc(digits, func(r rune) bool { return r < '0' || r > '9' }); i >= 0 {
		digits = digits[:i]
	}
	mi, err2 := strconv.Atoi(digits)
	if err1 != nil || err2 != nil {
		return false
	}
	return ma > major || ma == major && mi >= minor
}
