package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime/debug"
	"strconv"
	"strings"
)

// configFile describes a project; keel new writes it and keel build reads it.
const configFile = "keel.json"

// Config is keel.json.
type Config struct {
	// Name is shown to people: the window title, the .app and menu entry.
	Name string `json:"name"`
	// AppID is the reverse-DNS identifier: the macOS bundle ID, the Linux
	// app_id and .desktop name. Permissions and notifications are keyed by
	// it, so keep it once released.
	AppID string `json:"appid"`
	// Version is major.minor.patch; Build counts builds of one version.
	Version string `json:"version"`
	Build   int    `json:"build,omitempty"`
	// Binary is the executable's file name.
	Binary string `json:"binary"`
	// Icon is the artwork: a square PNG, 1024px, full bleed (no rounded
	// corners, no margin). keel build cuts it into each platform's shape.
	Icon string `json:"icon"`
	// IconMask is "platform" (the default), cutting the artwork into each
	// platform's plate, or "none", keeping the artwork's own outline and
	// transparency and only fitting it into the plate's area.
	IconMask string `json:"icon_mask,omitempty"`
	// Icons replaces the generated icon of a platform ("darwin", "windows",
	// "linux", "ios", "android") with a finished square PNG; iOS flattens transparency.
	Icons map[string]string `json:"icons,omitempty"`
	// Main is the main package, relative to keel.json.
	Main string `json:"main"`
	// IOS configures the mobile build; omitted projects target iOS 18.0.
	IOS     *IOSConfig     `json:"ios,omitempty"`
	Android *AndroidConfig `json:"android,omitempty"`
}

// AndroidConfig specifies API levels used by the Android packager.
type AndroidConfig struct {
	MinimumSDK int `json:"minimum_sdk,omitempty"`
	TargetSDK  int `json:"target_sdk,omitempty"`
}

func (c *Config) androidMinimumSDK() int {
	if c.Android != nil && c.Android.MinimumSDK != 0 {
		return c.Android.MinimumSDK
	}
	return 23
}
func (c *Config) androidTargetSDK() int {
	if c.Android != nil && c.Android.TargetSDK != 0 {
		return c.Android.TargetSDK
	}
	return 35
}

type IOSConfig struct {
	MinimumVersion string `json:"minimum_version,omitempty"`
}

func (c *Config) iosMinimumVersion() string {
	if c.IOS != nil && c.IOS.MinimumVersion != "" {
		return c.IOS.MinimumVersion
	}
	return "18.0"
}

var (
	appIDPattern      = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9-]*(\.[A-Za-z][A-Za-z0-9-]*)+$`)
	versionPattern    = regexp.MustCompile(`^\d+\.\d+\.\d+$`)
	iosVersionPattern = regexp.MustCompile(`^\d+\.\d+(\.\d+)?$`)
)

func (c *Config) validate() error {
	var errs []error
	if c.Name == "" {
		errs = append(errs, errors.New("name is empty"))
	}
	if !appIDPattern.MatchString(c.AppID) {
		errs = append(errs, fmt.Errorf("appid %q is not reverse DNS, like com.example.app", c.AppID))
	}
	if !versionPattern.MatchString(c.Version) {
		errs = append(errs, fmt.Errorf("version %q is not major.minor.patch", c.Version))
	}
	if c.Binary == "" || c.Binary == "." || c.Binary == ".." || strings.ContainsAny(c.Binary, `/\ `) {
		errs = append(errs, fmt.Errorf("binary %q is not a plain file name", c.Binary))
	}
	if c.Build < 0 {
		errs = append(errs, errors.New("build is negative"))
	}
	if c.IconMask != "" && c.IconMask != "platform" && c.IconMask != "none" {
		errs = append(errs, fmt.Errorf("icon_mask %q is not \"platform\" or \"none\"", c.IconMask))
	}
	for platform := range c.Icons {
		if platform != "darwin" && platform != "windows" && platform != "linux" && platform != "ios" && platform != "android" {
			errs = append(errs, fmt.Errorf("icons: unknown platform %q (darwin, windows, linux, ios, android)", platform))
		}
	}
	minimum := c.iosMinimumVersion()
	if !iosVersionPattern.MatchString(minimum) {
		errs = append(errs, fmt.Errorf("ios.minimum_version %q: use a version such as 18.0", minimum))
	} else if major, _ := strconv.Atoi(strings.Split(minimum, ".")[0]); major < 13 {
		errs = append(errs, errors.New("ios.minimum_version must be at least 13.0 for UIScene"))
	}
	if c.androidMinimumSDK() < 23 {
		errs = append(errs, errors.New("android.minimum_sdk must be at least 23"))
	}
	if c.androidTargetSDK() < max(31, c.androidMinimumSDK()) {
		errs = append(errs, errors.New("android.target_sdk must be at least 31 and not lower than minimum_sdk"))
	}
	return errors.Join(errs...)
}

// fourPart is the version in gogio's major.minor.patch.build form.
func (c *Config) fourPart() string { return fmt.Sprintf("%s.%d", c.Version, max(c.Build, 1)) }

// versionWords is the version as Windows' four 16-bit numbers.
func (c *Config) versionWords() [4]uint16 {
	var w [4]uint16
	for i, part := range strings.SplitN(c.fourPart(), ".", 4) {
		n, _ := strconv.Atoi(part)
		w[i] = uint16(min(max(n, 0), 65535))
	}
	return w
}

func loadConfig(dir string) (*Config, error) {
	data, err := os.ReadFile(filepath.Join(dir, configFile))
	if errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("no %s here; create a project with keel new", configFile)
	}
	if err != nil {
		return nil, err
	}
	c := &Config{Icon: "appicon.png", Main: "."}
	if err := json.Unmarshal(data, c); err != nil {
		return nil, fmt.Errorf("%s: %w", configFile, err)
	}
	if err := c.validate(); err != nil {
		return nil, fmt.Errorf("%s: %w", configFile, err)
	}
	return c, nil
}

func (c *Config) save(dir string) error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, configFile), append(data, '\n'), 0o644)
}

// keelVersion is this tool's release, which new projects require; a
// development build asks for the latest.
func keelVersion() string {
	if bi, ok := debug.ReadBuildInfo(); ok && bi.Main.Path == "github.com/dyike/keel" {
		if v := bi.Main.Version; v != "" && v != "(devel)" && !strings.Contains(v, "-0.") {
			return v
		}
	}
	return "latest"
}
