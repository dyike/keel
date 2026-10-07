//go:build !windows

package main

import (
	"context"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestRunWatchCancelStopsCompilerChildren(t *testing.T) {
	root := t.TempDir()
	bin := filepath.Join(root, "bin")
	project := filepath.Join(root, "project")
	for _, dir := range []string{bin, project} {
		if err := os.Mkdir(dir, 0700); err != nil {
			t.Fatal(err)
		}
	}
	cfg := &Config{Name: "Cancel", AppID: "com.example.cancel", Version: "1.0.0", Binary: "cancel", Main: "."}
	if err := cfg.save(project); err != nil {
		t.Fatal(err)
	}
	childPID := filepath.Join(root, "compiler-child.pid")
	t.Setenv("KEEL_WATCH_CHILD_PID", childPID)
	shim := `#!/bin/sh
if [ "$1" = "list" ]; then exit 0; fi
sleep 60 &
printf '%s' "$!" > "$KEEL_WATCH_CHILD_PID"
wait
`
	if err := os.WriteFile(filepath.Join(bin, "go"), []byte(shim), 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	c := &cli{out: io.Discard, errw: io.Discard}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() { done <- c.watchProject(ctx, project, nil) }()
	var pid string
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		data, _ := os.ReadFile(childPID)
		pid = strings.TrimSpace(string(data))
		if pid != "" {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	if pid == "" {
		t.Fatal("compiler child did not start")
	}
	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("build cancellation hung")
	}
	// A briefly unreaped zombie is already terminated; a running child is a leak.
	output, err := exec.Command("ps", "-p", pid, "-o", "stat=").Output()
	if err == nil && !strings.HasPrefix(strings.TrimSpace(string(output)), "Z") {
		exec.Command("kill", "-KILL", pid).Run()
		t.Fatalf("compiler child survived cancellation: %s", output)
	}
}

func TestRunWatchQuitDuringBuildStopsWatcher(t *testing.T) {
	realGo, err := exec.LookPath("go")
	if err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	project := filepath.Join(root, "project")
	bin := filepath.Join(root, "bin")
	for _, dir := range []string{project, bin} {
		if err := os.Mkdir(dir, 0700); err != nil {
			t.Fatal(err)
		}
	}
	write := func(path, text string) {
		t.Helper()
		if err := os.WriteFile(path, []byte(text), 0700); err != nil {
			t.Fatal(err)
		}
	}
	source := `package main
import("os";"time")
func main(){os.WriteFile(os.Getenv("KEEL_QUIT_START"),[]byte("started"),0600);for{if _,err:=os.Stat(os.Getenv("KEEL_QUIT_FILE"));err==nil{return};time.Sleep(20*time.Millisecond)}}
`
	write(filepath.Join(project, "main.go"), source)
	write(filepath.Join(project, "go.mod"), "module quitter\n\ngo 1.26.1\n")
	cfg := &Config{Name: "Quit", AppID: "com.example.quit", Version: "1.0.0", Binary: "quit", Main: "."}
	if err := cfg.save(project); err != nil {
		t.Fatal(err)
	}
	started := filepath.Join(root, "started")
	quit := filepath.Join(root, "quit")
	slow := filepath.Join(root, "slow")
	building := filepath.Join(root, "building")
	t.Setenv("KEEL_QUIT_START", started)
	t.Setenv("KEEL_QUIT_FILE", quit)
	t.Setenv("KEEL_REAL_GO", realGo)
	t.Setenv("KEEL_SLOW_BUILD", slow)
	t.Setenv("KEEL_BUILD_MARKER", building)
	write(filepath.Join(bin, "go"), `#!/bin/sh
if [ "$1" = "build" ] && [ -f "$KEEL_SLOW_BUILD" ]; then
 printf 'building' > "$KEEL_BUILD_MARKER"
 sleep 60
fi
exec "$KEEL_REAL_GO" "$@"
`)
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	output, err := os.Create(filepath.Join(root, "output.log"))
	if err != nil {
		t.Fatal(err)
	}
	defer output.Close()
	c := &cli{out: output, errw: output}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() { done <- c.watchProject(ctx, project, nil) }()
	finished := false
	defer func() {
		if !finished {
			cancel()
			select {
			case <-done:
			case <-time.After(5 * time.Second):
				t.Error("watcher did not stop")
			}
		}
	}()
	waitFile := func(path string) {
		t.Helper()
		deadline := time.Now().Add(10 * time.Second)
		for time.Now().Before(deadline) {
			if _, err := os.Stat(path); err == nil {
				return
			}
			time.Sleep(20 * time.Millisecond)
		}
		data, _ := os.ReadFile(output.Name())
		t.Fatalf("missing %s: %s", path, data)
	}
	waitFile(started)
	write(slow, "hold compile")
	write(filepath.Join(project, "main.go"), source+"\n// trigger rebuild\n")
	waitFile(building)
	write(quit, "user quit")
	select {
	case err := <-done:
		finished = true
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("user quit left watcher/build alive")
	}
	data, _ := os.ReadFile(output.Name())
	if strings.Count(string(data), "hot reload enabled") != 1 || !strings.Contains(string(data), "application closed; stopping watch") {
		t.Fatal("app restarted after quit", string(data))
	}
}
