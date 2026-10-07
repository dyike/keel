package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

// Match gogio's SDK/NDK lookup order, including its side-by-side NDK preference.
func androidNDK(root string) string {
	paths, _ := filepath.Glob(filepath.Join(root, "ndk", "*"))
	if latest := androidLatestVersion(paths); latest != "" {
		return latest
	}
	if info, err := os.Stat(filepath.Join(root, "ndk-bundle")); err == nil && info.IsDir() {
		return filepath.Join(root, "ndk-bundle")
	}
	return os.Getenv("ANDROID_NDK_ROOT")
}

func androidLatestVersion(paths []string) string {
	var best string
	var version [3]int
	for _, path := range paths {
		parts := strings.Split(filepath.Base(path), ".")
		var candidate [3]int
		valid := len(parts) <= 3
		for i, part := range parts {
			if i >= 3 {
				break
			}
			n, err := strconv.Atoi(part)
			if err != nil {
				valid = false
			}
			candidate[i] = n
		}
		if !valid {
			continue
		}
		newer := best == ""
		for i := range candidate {
			if candidate[i] != version[i] {
				newer = candidate[i] > version[i]
				break
			}
		}
		if newer {
			best, version = path, candidate
		}
	}
	return best
}

func (c *cli) doctorAndroid(check func(string, bool, string)) {
	root := androidSDK()
	info, err := os.Stat(root)
	check("Android SDK", root != "" && err == nil && info.IsDir(), "set ANDROID_HOME (or ANDROID_SDK_ROOT): "+root)
	platforms, _ := filepath.Glob(filepath.Join(root, "platforms", "android-*", "android.jar"))
	sdk := 0
	for _, path := range platforms {
		n, _ := strconv.Atoi(strings.TrimPrefix(filepath.Base(filepath.Dir(path)), "android-"))
		sdk = max(sdk, n)
	}
	required := 35
	if cfg, err := loadConfig(c.wd()); err == nil {
		required = max(31, cfg.androidTargetSDK())
	}
	check("Android platform SDK", sdk >= required, "install platforms; required API "+strconv.Itoa(required)+", highest installed "+strconv.Itoa(sdk))
	tools, _ := filepath.Glob(filepath.Join(root, "build-tools", "*"))
	buildtools := androidLatestVersion(tools)
	major, _ := strconv.Atoi(strings.Split(filepath.Base(buildtools), ".")[0])
	check("build-tools 35 or newer", major >= 35, "older D8 versions may fail with modern JDKs: "+buildtools)
	for _, tool := range []string{"aapt2", "zipalign", "apksigner", "d8"} {
		suffix := androidExeSuffix()
		if runtime.GOOS == "windows" && (tool == "apksigner" || tool == "d8") {
			suffix = ".bat"
		}
		info, err := os.Stat(filepath.Join(buildtools, tool+suffix))
		check("build-tools "+tool, buildtools != "" && err == nil && !info.IsDir(), buildtools)
	}
	ndk := androidNDK(root)
	host := runtime.GOOS + "-x86_64" // Google's macOS NDK uses this name on Apple Silicon too.
	compilers, _ := filepath.Glob(filepath.Join(ndk, "toolchains", "llvm", "prebuilt", host, "bin", "aarch64-linux-android*-clang"+androidExeSuffix()))
	check("Android NDK", ndk != "" && len(compilers) > 0, "install NDK (side by side), or set ANDROID_NDK_ROOT: "+ndk)
	for _, tool := range []string{"javac", "java", "keytool"} {
		path := tool
		if home := os.Getenv("JAVA_HOME"); home != "" {
			path = filepath.Join(home, "bin", tool+androidExeSuffix())
		}
		_, err := exec.LookPath(path)
		check(tool, err == nil, "install a JDK and set JAVA_HOME or PATH")
	}
	_, err = exec.LookPath(androidTool("adb"))
	check("adb", err == nil, "install platform-tools for keel run -target android")
}
