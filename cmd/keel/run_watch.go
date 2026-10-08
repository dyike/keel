package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

type runFile struct {
	size     int64
	modified int64
}
type runSnapshot map[string]runFile

// Polling also works across local replacements, file additions, atomic editor
// saves and network filesystems, without consuming a descriptor per directory.
func snapshotRunFiles(roots []string) (runSnapshot, error) {
	state := make(runSnapshot)
	for _, root := range roots {
		err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
			if os.IsNotExist(err) {
				return nil
			}
			if err != nil {
				return err
			}
			name := entry.Name()
			if path != root && entry.IsDir() {
				switch name {
				case ".git", "node_modules", "vendor", "dist", ".cache", ".idea", ".vscode", "__pycache__":
					return filepath.SkipDir
				}
			}
			if entry.IsDir() || !entry.Type().IsRegular() {
				return nil
			}
			if strings.HasPrefix(name, ".") || strings.HasSuffix(name, "~") {
				return nil
			}
			switch filepath.Ext(name) {
			case ".log", ".tmp", ".swp", ".swo", ".lock":
				return nil
			}
			info, err := entry.Info()
			if os.IsNotExist(err) {
				return nil
			}
			if err != nil {
				return err
			}
			state[path] = runFile{info.Size(), info.ModTime().UnixNano()}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	return state, nil
}

// Module metadata gives both go.work modules and directory-based replacements;
// downloaded versioned dependencies are immutable and are not watched.
func runWatchRoots(ctx context.Context, dir string) []string {
	roots := []string{dir}
	workspace := os.Getenv("GOWORK")
	if workspace == "" {
		for parent := dir; ; parent = filepath.Dir(parent) {
			candidate := filepath.Join(parent, "go.work")
			if _, err := os.Stat(candidate); err == nil {
				workspace = candidate
				break
			}
			if filepath.Dir(parent) == parent {
				break
			}
		}
	}
	if workspace != "" && workspace != "off" {
		roots = append(roots, workspace)
	}
	cmd := exec.CommandContext(ctx, "go", "list", "-m", "-json", "all")
	cmd.Dir = dir
	data, err := cmd.Output()
	if err != nil {
		return roots
	}
	decoder := json.NewDecoder(strings.NewReader(string(data)))
	seen := map[string]bool{filepath.Clean(dir): true}
	for {
		var module struct {
			Dir, Version string
			Main         bool
			Replace      *struct{ Dir, Version string }
		}
		if decoder.Decode(&module) != nil {
			break
		}
		root := ""
		if module.Main {
			root = module.Dir
		}
		if module.Replace != nil && module.Replace.Version == "" {
			root = module.Replace.Dir
		}
		if root != "" && !seen[filepath.Clean(root)] {
			roots = append(roots, root)
			seen[filepath.Clean(root)] = true
		}
	}
	return roots
}

type runOutput struct {
	mu     *sync.Mutex
	writer io.Writer
}

func (w runOutput) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.writer.Write(p)
}

type runBuild struct {
	binary, icon string
	cleanup      func()
	err          error
}

func (c *cli) compileRun(ctx context.Context, dir, tmp string, generation int) runBuild {
	result := runBuild{cleanup: func() {}}
	cfg, err := loadConfig(dir)
	if err != nil {
		result.err = err
		return result
	}
	result.binary = filepath.Join(tmp, fmt.Sprintf("%s-%d", cfg.Binary, generation))
	if runtime.GOOS == "windows" {
		result.binary += ".exe"
	}
	cmd := exec.CommandContext(ctx, "go", "build", "-o", result.binary, "-ldflags", appIDFlag(cfg), "./"+filepath.ToSlash(filepath.Clean(cfg.Main)))
	cmd.Dir = dir
	cmd.Stdout = c.out
	cmd.Stderr = c.errw
	configureRunProcess(cmd)
	cmd.Cancel = func() error { return killRunProcess(cmd) }
	cmd.WaitDelay = 2 * time.Second
	result.err = cmd.Run()
	if result.err == nil {
		result.icon, result.cleanup, result.err = prepareRunIcon(dir, cfg, runtime.GOOS)
		if result.err == nil {
			var cleanupBundle func()
			result.binary, cleanupBundle, result.err = prepareRunBundle(ctx, cfg, result.binary, result.icon)
			cleanupIcon := result.cleanup
			result.cleanup = func() { cleanupBundle(); cleanupIcon() }
		}
	}
	return result
}

func (c *cli) watchProject(ctx context.Context, dir string, args []string) error {
	dir, err := filepath.Abs(dir)
	if err != nil {
		return err
	}
	tmp, err := os.MkdirTemp("", "keel-run-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)
	var outputMu sync.Mutex
	runner := *c
	runner.out = runOutput{&outputMu, c.out}
	runner.errw = runOutput{&outputMu, c.errw}
	rootsCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	roots := runWatchRoots(rootsCtx, dir)
	cancel()
	snapshot, err := snapshotRunFiles(roots)
	if err != nil {
		return err
	}
	fmt.Fprintln(runner.out, "keel: watching project and local dependencies; Ctrl+C to stop")
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()
	builds := make(chan runBuild, 1)
	var buildCancel context.CancelFunc
	var app *exec.Cmd
	var appDone chan error
	appCleanup := func() {}
	// Stop and reap children before removing their binary/icon files.
	defer func() {
		if buildCancel != nil {
			buildCancel()
			result := <-builds
			result.cleanup()
		}
		if app != nil {
			stopRunProcess(app, appDone)
		}
		appCleanup()
	}()
	dirty := true
	changed := time.Time{}
	generation := 0
	for {
		if dirty && buildCancel == nil && time.Since(changed) >= 200*time.Millisecond {
			dirty = false
			generation++
			buildCtx, cancel := context.WithCancel(ctx)
			buildCancel = cancel
			fmt.Fprintln(runner.out, "keel: building...")
			go func(n int) { builds <- runner.compileRun(buildCtx, dir, tmp, n) }(generation)
		}
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			next, err := snapshotRunFiles(roots)
			if err != nil {
				fmt.Fprintln(runner.errw, "keel: watch:", err)
				continue
			}
			if !maps.Equal(snapshot, next) {
				snapshot = next
				dirty = true
				changed = time.Now()
				if buildCancel != nil {
					buildCancel()
				}
			}
		case result := <-builds:
			buildCancel()
			buildCancel = nil
			if dirty {
				result.cleanup()
				os.Remove(result.binary)
				continue
			}
			if result.err != nil {
				result.cleanup()
				os.Remove(result.binary)
				fmt.Fprintln(runner.errw, "keel: build failed; waiting for changes:", result.err)
				continue
			}
			if app != nil {
				// A quit can race with compilation finishing. Give an already
				// exited application priority rather than reopening its window.
				select {
				case exitErr := <-appDone:
					app = nil
					appDone = nil
					if exitErr == nil {
						result.cleanup()
						os.Remove(result.binary)
						fmt.Fprintln(runner.out, "keel: application closed; stopping watch")
						return nil
					}
				default:
					stopRunProcess(app, appDone)
					app = nil
					appDone = nil
				}
			}
			appCleanup()
			binary := result.binary
			cleanupIcon := result.cleanup
			appCleanup = func() { cleanupIcon(); os.Remove(binary) }
			app = exec.Command(binary, args...)
			app.Dir = dir
			app.Stdin = os.Stdin
			app.Stdout = runner.out
			app.Stderr = runner.errw
			app.Env = append(os.Environ(), "KEEL_RUN_ICON="+result.icon, "KEEL_RUN_WATCH=1")
			configureRunProcess(app)
			if err := app.Start(); err != nil {
				app = nil
				return err
			}
			appDone = make(chan error, 1)
			current, done := app, appDone
			go func() { done <- current.Wait() }()
			fmt.Fprintf(runner.out, "keel: running (hot reload enabled; pid %d)\n", app.Process.Pid)
			// A changed go.mod/go.work may add or remove local replacements.
			rootsCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
			nextRoots := runWatchRoots(rootsCtx, dir)
			cancel()
			if !sameRunRoots(roots, nextRoots) {
				roots = nextRoots
				snapshot, err = snapshotRunFiles(roots)
				if err != nil {
					return err
				}
			}
		case err := <-appDone:
			app = nil
			appDone = nil
			if err == nil {
				fmt.Fprintln(runner.out, "keel: application closed; stopping watch")
				return nil
			}
			fmt.Fprintln(runner.errw, "keel: application exited; waiting for changes:", err)
		}
	}
}

func sameRunRoots(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	set := map[string]bool{}
	for _, root := range a {
		set[root] = true
	}
	for _, root := range b {
		if !set[root] {
			return false
		}
	}
	return true
}

func stopRunProcess(cmd *exec.Cmd, done <-chan error) {
	interruptRunProcess(cmd)
	timer := time.NewTimer(1500 * time.Millisecond)
	defer timer.Stop()
	select {
	case <-done:
	case <-timer.C:
		killRunProcess(cmd)
		<-done
	}
}
