// Command keel creates, runs and packages Keel desktop apps.
//
//	keel new myapp         create a project
//	keel run               run it
//	keel build             package it for this platform, into dist/
//	keel build -target windows
//	keel icon              write each platform's icons, to check them
//	keel doctor            check the toolchain for each target
//
// Install with go install github.com/dyike/keel/cmd/keel@latest.
package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

func main() {
	c := &cli{out: os.Stdout, errw: os.Stderr}
	os.Exit(c.main(os.Args[1:]))
}

// cli carries what commands write to and how they run tools, so tests can
// capture both.
type cli struct {
	out, errw io.Writer
	dir       string // working directory; "" is the process's
	dryRun    bool   // print external commands instead of running them
}

const usage = `keel creates, runs and packages Keel desktop apps.

Usage:
  keel new <dir> [flags]     create a project in dir
  keel run [-- args]         run the project in this directory
  keel build [flags]         package the project into dist/
  keel icon [-o dir]         write each platform's icons, to check them
  keel doctor                check the toolchain for each target
  keel version               print the version

Run "keel <command> -h" for a command's flags.
`

func (c *cli) main(args []string) int {
	if len(args) == 0 {
		fmt.Fprint(c.errw, usage)
		return 2
	}
	var err error
	switch args[0] {
	case "new":
		err = c.newProject(args[1:])
	case "run":
		err = c.runProject(args[1:])
	case "build":
		err = c.build(args[1:])
	case "icon":
		err = c.iconCommand(args[1:])
	case "doctor":
		err = c.doctor(args[1:])
	case "version":
		fmt.Fprintln(c.out, "keel", keelVersion())
	case "help", "-h", "-help", "--help":
		fmt.Fprint(c.out, usage)
	default:
		fmt.Fprintf(c.errw, "keel: unknown command %q\n\n%s", args[0], usage)
		return 2
	}
	if err != nil {
		fmt.Fprintln(c.errw, "keel:", err)
		return 1
	}
	return 0
}

// command runs (or with -n prints) an external tool in dir, its output
// passed through.
func (c *cli) command(dir string, env []string, name string, args ...string) error {
	if c.dryRun {
		fmt.Fprintln(c.out, strings.Join(append([]string{name}, args...), " "))
		return nil
	}
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, c.out, c.errw
	if len(env) > 0 {
		cmd.Env = append(os.Environ(), env...)
	}
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s %s: %w", name, strings.Join(args, " "), err)
	}
	return nil
}

func (c *cli) wd() string {
	if c.dir != "" {
		return c.dir
	}
	wd, _ := os.Getwd()
	return wd
}
