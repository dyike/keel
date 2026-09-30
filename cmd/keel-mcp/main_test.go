package main

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// TestEndToEnd runs keel-mcp as an agent would and drives the multiwindow
// example through it: read, type, click, open a second window, press keys,
// screenshot, close. The app renders off-screen, so no display is touched.
func TestEndToEnd(t *testing.T) {
	if testing.Short() {
		t.Skip("builds and runs the example app")
	}
	root := repoRoot()

	call, _ := startServer(t, nil)

	expect(t, call("launch", map[string]any{"command": "go run ./examples/multiwindow", "dir": root}),
		`Window w1 "Keel 主窗口"`, `textbox "你的名字" value=""`, `button "打招呼"`)
	call("type", map[string]any{"text": "小明", "ref": "e3"})
	expect(t, call("click", map[string]any{"text": "打招呼"}), `text "你好，小明！"`)
	expect(t, call("press_key", map[string]any{"key": "mod+,"}), `w2 "设置" (new)`)
	expect(t, call("snapshot", map[string]any{"window": "w2"}), `checkbox "启用提示" checked`, `value="设置窗口 #1"`)
	expect(t, call("click", map[string]any{"window": "w2", "text": "启用提示"}), `checkbox "启用提示" unchecked`)
	expect(t, call("click", map[string]any{"window": "w2", "text": "保存"}), "已保存：设置窗口 #1，提示=false")
	expect(t, call("screenshot", map[string]any{"window": "w2"}), "image:image/png")
	// The real window must go, not just its shadow: the list is updated only
	// when the on-screen window is destroyed.
	if got := call("close_window", map[string]any{"window": "w2"}); strings.Contains(got, "设置") {
		t.Fatalf("settings window still open after close_window:\n%s", got)
	}
	// Reopening works, and the app sees the old window as closed.
	expect(t, call("press_key", map[string]any{"key": "mod+,"}), `"设置" (new)`)
	expect(t, call("close_window", nil), "Closed.")
	expect(t, call("close_window", nil), "last window")
}

// TestAttach starts the app the way a user would, then attaches, detaches and
// attaches again: state survives, and a second client is told the app is busy.
func TestAttach(t *testing.T) {
	if testing.Short() {
		t.Skip("builds and runs the example app")
	}
	// A short private TMPDIR: socket paths are limited to ~100 bytes, and the
	// app and keel-mcp must agree on it to find each other.
	tmp, err := os.MkdirTemp("/tmp", "k")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmp)
	env := append(os.Environ(), "TMPDIR="+tmp)

	bin := filepath.Join(tmp, "multiwindow")
	build := exec.Command("go", "build", "-o", bin, "./examples/multiwindow")
	build.Dir = repoRoot()
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build: %v\n%s", err, out)
	}
	appCmd := exec.Command(bin)
	appCmd.Env = append(env, "KEEL_AUTOMATION=1", "KEEL_HEADLESS=1")
	if err := appCmd.Start(); err != nil {
		t.Fatal(err)
	}
	exited := make(chan error, 1)
	go func() { exited <- appCmd.Wait() }()
	defer appCmd.Process.Kill()

	call, try := startServer(t, env)
	var got string
	for deadline := time.Now().Add(30 * time.Second); ; time.Sleep(200 * time.Millisecond) {
		if got, err = try("attach", nil); err == nil {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("attach never succeeded: %v", err)
		}
	}
	expect(t, got, "Attached to "+filepath.Join(tmp, "keel", "multiwindow-"), `textbox "你的名字" value=""`)
	call("type", map[string]any{"text": "小明", "ref": "e3"})
	expect(t, call("stop", nil), "Disconnected; the app keeps running")
	select {
	case err := <-exited:
		t.Fatalf("app exited after detach: %v", err)
	case <-time.After(300 * time.Millisecond):
	}

	expect(t, call("attach", nil), `value="小明"`) // state survived the detach
	_, other := startServer(t, env)
	if _, err := other("attach", nil); err == nil || !strings.Contains(err.Error(), "another client") {
		t.Fatalf("second client: got %v, want a busy error", err)
	}
	expect(t, call("close_window", nil), "last window")
	select {
	case <-exited:
	case <-time.After(5 * time.Second):
		t.Fatal("app did not exit after its last window closed")
	}
}

func repoRoot() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(file), "..", "..")
}

// startServer runs keel-mcp as an MCP client would. call fails the test on a
// tool error; try returns it.
func startServer(t *testing.T, env []string) (call func(string, map[string]any) string, try func(string, map[string]any) (string, error)) {
	t.Helper()
	ctx := context.Background()
	cmd := exec.Command("go", "run", ".")
	cmd.Dir = filepath.Join(repoRoot(), "cmd", "keel-mcp")
	cmd.Env = env
	session, err := mcp.NewClient(&mcp.Implementation{Name: "test"}, nil).Connect(ctx, &mcp.CommandTransport{Command: cmd}, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { session.Close() })
	try = func(name string, args map[string]any) (string, error) {
		res, err := session.CallTool(ctx, &mcp.CallToolParams{Name: name, Arguments: args})
		if err != nil {
			return "", err
		}
		var out []string
		for _, c := range res.Content {
			switch c := c.(type) {
			case *mcp.TextContent:
				out = append(out, c.Text)
			case *mcp.ImageContent:
				out = append(out, "image:"+c.MIMEType)
			}
		}
		s := strings.Join(out, "\n")
		if res.IsError {
			return s, errors.New(s)
		}
		return s, nil
	}
	call = func(name string, args map[string]any) string {
		t.Helper()
		s, err := try(name, args)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		t.Logf("%s → %s", name, s)
		return s
	}
	return call, try
}

func expect(t *testing.T, got string, wants ...string) {
	t.Helper()
	for _, w := range wants {
		if !strings.Contains(got, w) {
			t.Fatalf("missing %q in:\n%s", w, got)
		}
	}
}

// TestVisible drives an app whose windows are on screen: agent actions go
// through each window's shadow while the real window keeps running. Set
// KEEL_DESKTOP=1 to run it; it opens windows for a few seconds.
func TestVisible(t *testing.T) {
	if os.Getenv("KEEL_DESKTOP") != "1" {
		t.Skip("set KEEL_DESKTOP=1 to open real windows")
	}
	call, _ := startServer(t, nil)
	expect(t, call("launch", map[string]any{"command": "go run ./examples/multiwindow", "dir": repoRoot(), "visible": true}),
		`Window w1 "Keel 主窗口" 640×420`)
	call("type", map[string]any{"text": "小明", "ref": "e3"})
	expect(t, call("click", map[string]any{"text": "打招呼"}), `text "你好，小明！"`)
	expect(t, call("click", map[string]any{"text": "打开设置窗口"}), `w2 "设置" (new)`)
	expect(t, call("click", map[string]any{"window": "w2", "text": "保存"}), "已保存：设置窗口 #1，提示=true")
	expect(t, call("screenshot", map[string]any{"window": "w2"}), "image:image/png")
	// The real window must go, not just its shadow: the list only drops a
	// window when the on-screen one is destroyed.
	if got := call("close_window", map[string]any{"window": "w2"}); strings.Contains(got, "设置") {
		t.Fatalf("settings window still open after close_window:\n%s", got)
	}
	expect(t, call("press_key", map[string]any{"key": "mod+,"}), `"设置" (new)`)
	expect(t, call("close_window", nil), "Closed.")
	expect(t, call("close_window", nil), "last window")
}
