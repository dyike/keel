package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestRunWatchSnapshot(t *testing.T) {
	root := t.TempDir()
	write := func(name, text string) {
		t.Helper()
		path := filepath.Join(root, name)
		if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(text), 0600); err != nil {
			t.Fatal(err)
		}
	}
	write("main.go", "first")
	write("assets/embed.bin", "asset")
	before, err := snapshotRunFiles([]string{root})
	if err != nil {
		t.Fatal(err)
	}
	for _, ignored := range []string{".git/index", "dist/output.go", "node_modules/script.js", "server.log", "lock.lock", "file.go~", ".main.go.swp"} {
		write(ignored, "ignored")
	}
	after, err := snapshotRunFiles([]string{root})
	if err != nil {
		t.Fatal(err)
	}
	if len(after) != len(before) {
		t.Fatal("generated/editor files were watched", after)
	}
	write("main.go", "updated source")
	write("new.go", "new source")
	os.Remove(filepath.Join(root, "assets/embed.bin"))
	after, err = snapshotRunFiles([]string{root})
	if err != nil {
		t.Fatal(err)
	}
	if after[filepath.Join(root, "main.go")] == before[filepath.Join(root, "main.go")] {
		t.Fatal("edit not detected")
	}
	if _, ok := after[filepath.Join(root, "new.go")]; !ok {
		t.Fatal("new file not detected")
	}
	if _, ok := after[filepath.Join(root, "assets/embed.bin")]; ok {
		t.Fatal("deleted file retained")
	}
}

func TestRunWatchRebuildRecoveryAndLocalReplacement(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "app")
	dep := filepath.Join(root, "dep")
	for _, folder := range []string{dir, dep} {
		if err := os.Mkdir(folder, 0700); err != nil {
			t.Fatal(err)
		}
	}
	write := func(path, text string) {
		t.Helper()
		if err := os.WriteFile(path, []byte(text), 0600); err != nil {
			t.Fatal(err)
		}
	}
	write(filepath.Join(dir, "go.mod"), "module watchtest\n\ngo 1.26.1\nrequire watchdep v0.0.0\nreplace watchdep => ../dep\n")
	write(filepath.Join(dep, "go.mod"), "module watchdep\n\ngo 1.26.1\n")
	depSource := filepath.Join(dep, "value.go")
	write(depSource, "package watchdep\nconst Value = \"first\"\n")
	write(filepath.Join(dir, "caption.txt"), "asset-one")
	source := `package main
import (
 "os"
 "os/signal"
 "encoding/json"
 _ "embed"
 "watchdep"
)
//go:embed caption.txt
var caption string
func event(kind string){exe,_:=os.Executable();f,_:=os.OpenFile(os.Getenv("KEEL_WATCH_TEST_EVENTS"),os.O_CREATE|os.O_WRONLY|os.O_APPEND,0600);defer f.Close();json.NewEncoder(f).Encode(map[string]any{"kind":kind,"pid":os.Getpid(),"exe":exe,"value":watchdep.Value,"caption":caption,"args":os.Args[1:]})}
func main(){event("start");signalCh:=make(chan os.Signal,1);signal.Notify(signalCh,os.Interrupt);<-signalCh;event("stop")}
`
	write(filepath.Join(dir, "main.go"), source)
	cfg := &Config{Name: "Watch", AppID: "com.example.watch", Version: "1.0.0", Binary: "watch", Icon: "appicon.png", Main: "."}
	if err := cfg.save(dir); err != nil {
		t.Fatal(err)
	}
	events := filepath.Join(root, "events.log")
	t.Setenv("KEEL_WATCH_TEST_EVENTS", events)
	output, err := os.Create(filepath.Join(root, "output.log"))
	if err != nil {
		t.Fatal(err)
	}
	defer output.Close()
	c := &cli{dir: dir, out: output, errw: output}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- c.watchProject(ctx, dir, []string{"--label", "two words"}) }()
	stopped := false
	defer func() {
		if !stopped {
			cancel()
			select {
			case <-done:
			case <-time.After(10 * time.Second):
				t.Error("watch did not stop")
			}
		}
	}()
	type event struct {
		Kind           string
		PID            int
		Value, Caption string
		Exe            string
		Args           []string
	}
	readEvents := func() []event {
		data, _ := os.ReadFile(events)
		var result []event
		for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
			var e event
			if json.Unmarshal([]byte(line), &e) == nil {
				result = append(result, e)
			}
		}
		return result
	}
	waitFor := func(predicate func([]event) bool) []event {
		t.Helper()
		deadline := time.Now().Add(30 * time.Second)
		for time.Now().Before(deadline) {
			es := readEvents()
			if predicate(es) {
				return es
			}
			select {
			case err := <-done:
				stopped = true
				t.Fatalf("watch exited early: %v", err)
			default:
			}
			time.Sleep(50 * time.Millisecond)
		}
		log, _ := os.ReadFile(output.Name())
		t.Fatalf("timed out: %s", log)
		return nil
	}
	starts := func(es []event) []event {
		var out []event
		for _, e := range es {
			if e.Kind == "start" {
				out = append(out, e)
			}
		}
		return out
	}
	first := starts(waitFor(func(es []event) bool { return len(starts(es)) == 1 }))[0]
	if fmt.Sprint(first.Args) != "[--label two words]" {
		t.Fatal(first.Args)
	}
	// A failed edit keeps the current executable alive and watches for recovery.
	write(depSource, "package watchdep\nconst Value =\n")
	deadline := time.Now().Add(10 * time.Second)
	for {
		data, _ := os.ReadFile(output.Name())
		if strings.Contains(string(data), "build failed; waiting for changes") {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("invalid edit not compiled", string(data))
		}
		time.Sleep(50 * time.Millisecond)
	}
	if es := readEvents(); len(es) != 1 {
		t.Fatal("failed compile stopped the old app", es)
	}
	write(depSource, "package watchdep\nconst Value = \"second\"\n")
	second := starts(waitFor(func(es []event) bool { return len(starts(es)) == 2 }))[1]
	if second.Value != "second" || second.PID == first.PID {
		t.Fatal("replacement change did not restart", second)
	}
	write(filepath.Join(dir, "caption.txt"), "asset-two")
	third := starts(waitFor(func(es []event) bool { return len(starts(es)) == 3 }))[2]
	if third.Caption != "asset-two" {
		t.Fatal("embedded asset not rebuilt", third)
	}
	cancel()
	select {
	case err := <-done:
		stopped = true
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("cancel did not stop watcher")
	}
	// On Unix, interruption is graceful; Windows terminates the process tree.
	es := readEvents()
	if es[len(es)-1].PID != third.PID {
		t.Fatal("wrong final application", es)
	}
	if runtime.GOOS != "windows" {
		stops := 0
		for _, e := range es {
			if e.Kind == "stop" {
				stops++
			}
		}
		if stops != 3 || es[len(es)-1].Kind != "stop" {
			t.Fatal("applications were not stopped and reaped", es)
		}
	}
	for _, e := range starts(es) {
		if _, err := os.Stat(e.Exe); !os.IsNotExist(err) {
			t.Fatal("temporary executable was not removed", e.Exe, err)
		}
	}

}

func TestRunWatchRootsIncludeWorkspaceModules(t *testing.T) {
	root := t.TempDir()
	for _, name := range []string{"first", "second"} {
		dir := filepath.Join(root, name)
		if err := os.Mkdir(dir, 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module "+name+"\n\ngo 1.26.1\n"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	workspace := filepath.Join(root, "go.work")
	if err := os.WriteFile(workspace, []byte("go 1.26.1\nuse (\n ./first\n ./second\n)\n"), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GOWORK", workspace)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	roots := runWatchRoots(ctx, filepath.Join(root, "first"))
	if !sameRunRoots(roots, []string{filepath.Join(root, "first"), filepath.Join(root, "second"), workspace}) {
		t.Fatal("workspace roots", roots)
	}
	before, err := snapshotRunFiles(roots)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(workspace, []byte("go 1.26.1\nuse ./first\n"), 0600); err != nil {
		t.Fatal(err)
	}
	after, err := snapshotRunFiles(roots)
	if err != nil {
		t.Fatal(err)
	}
	if before[workspace] == after[workspace] {
		t.Fatal("workspace edit not detected")
	}
}
