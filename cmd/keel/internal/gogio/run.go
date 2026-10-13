// SPDX-License-Identifier: Unlicense OR MIT

package gogio

// Keel: gogio runs inside the keel command instead of go run
// gioui.org/cmd/gogio, so it can find Keel's Gio at its import path. Its
// upstream code keeps package-level state, so Run serializes calls.

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// GioRoot is the import path of Keel's copy of Gio.
const GioRoot = "github.com/dyike/keel/gio"

// Options is the context a packaging run would get from its process.
type Options struct {
	Dir            string    // working directory; default: the current one
	Env            []string  // environment; default: os.Environ()
	Stdout, Stderr io.Writer // default: os.Stdout, os.Stderr
	GioRoot        string    // Gio's import path; default: GioRoot
}

var (
	runMu   sync.Mutex
	flags   *flag.FlagSet
	workDir string
	env     []string
	stdout  io.Writer
	stderr  io.Writer
	gioRoot string
)

func environ() []string { return append([]string(nil), env...) }

// getenv reads the run's environment, last setting winning as in exec.
func getenv(key string) string {
	for i := len(env) - 1; i >= 0; i-- {
		if k, v, ok := strings.Cut(env[i], "="); ok && k == key {
			return v
		}
	}
	return ""
}

// Run packages a Gio program; args are gogio's command line, without the
// program name.
func Run(args []string, o Options) error {
	runMu.Lock()
	defer runMu.Unlock()
	workDir, env, stdout, stderr, gioRoot = o.Dir, o.Env, o.Stdout, o.Stderr, o.GioRoot
	if workDir == "" {
		wd, err := os.Getwd()
		if err != nil {
			return err
		}
		workDir = wd
	}
	absolute, err := filepath.Abs(workDir)
	if err != nil {
		return err
	}
	workDir = absolute
	if env == nil {
		env = os.Environ()
	}
	if stdout == nil {
		stdout = os.Stdout
	}
	if stderr == nil {
		stderr = os.Stderr
	}
	if gioRoot == "" {
		gioRoot = GioRoot
	}
	flags = flag.NewFlagSet("gogio", flag.ContinueOnError)
	flags.SetOutput(stderr)
	flags.Usage = func() { fmt.Fprint(stderr, mainUsage) }
	defineFlags(flags)
	if err := flags.Parse(args); err != nil {
		return err
	}
	// Paths are relative to the run's directory, not the process's.
	for _, p := range []*string{destPath, iconPath, signKey} {
		if *p != "" && !filepath.IsAbs(*p) {
			*p = filepath.Join(workDir, *p)
		}
	}
	if err := flagValidate(); err != nil {
		return fmt.Errorf("gogio: %w", err)
	}
	bi, err := newBuildInfo(flags.Arg(0))
	if err != nil {
		return fmt.Errorf("gogio: %w", err)
	}
	if err := build(bi); err != nil {
		return fmt.Errorf("gogio: %w", err)
	}
	return nil
}

func outputPath(path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(workDir, path)
}
