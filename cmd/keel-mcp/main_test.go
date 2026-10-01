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

// TestOrders walks the order desk example through every widget the way an
// agent would: search, dropdown filter, sort, row selection with keys, detail
// and confirm dialogs, a form, tabs and progress bars.
func TestOrders(t *testing.T) {
	if testing.Short() {
		t.Skip("builds and runs the example app")
	}
	call, _ := startServer(t, nil)
	got := call("launch", map[string]any{"command": "go run ./examples/orders", "dir": repoRoot()})
	expect(t, got, `tab "订单列表" selected`, `table "" value="36 行"`, `columnheader "金额"`, `row "SO-1001 | 华东物流 | 待付款 | 300.00"`)

	expect(t, call("type", map[string]any{"text": "华东", "ref": "e5"}), "显示 6 条")
	expect(t, call("click", map[string]any{"text": "全部状态"}), `option "已发货"`)
	expect(t, call("click", map[string]any{"text": "已发货"}), `select "请选择" value="已发货"`, "显示 3 条")
	call("type", map[string]any{"text": "", "ref": "e5", "clear": true})
	call("press_key", map[string]any{"key": "backspace"})
	call("click", map[string]any{"text": "已发货"})
	expect(t, call("click", map[string]any{"text": "全部状态"}), "显示 36 条")

	call("click", map[string]any{"text": "金额"})
	got = call("click", map[string]any{"text": "金额"}) // descending
	expect(t, got, `row "SO-1022 | 成都餐饮 | 已完成 | 3177.00"`)
	call("click", map[string]any{"text": "SO-1022"})
	expect(t, call("press_key", map[string]any{"key": "down"}), `row "SO-1021 | 深圳电子 | 待付款 | 3040.00" selected`)
	expect(t, call("press_key", map[string]any{"key": "enter"}), `dialog "SO-1021"`, "深圳电子 · 待付款")
	call("click", map[string]any{"text": "确定"})

	expect(t, call("click", map[string]any{"text": "删除所选"}), `alertdialog "删除订单"`, `button "删除"`)
	if got := call("press_key", map[string]any{"key": "esc"}); strings.Contains(got, "dialog") {
		t.Fatalf("Esc left the dialog open:\n%s", got)
	}
	call("click", map[string]any{"text": "删除所选"})
	expect(t, call("click", map[string]any{"text": "删除"}), "共 35 条")

	got = call("press_key", map[string]any{"key": "mod+n"})
	expect(t, got, `tab "新建订单" selected`, `textbox "客户"`, `radio "转账" checked`, `switch "加急处理" unchecked`)
	expect(t, call("click", map[string]any{"text": "保存订单"}), "请填写客户名称。")
	call("type", map[string]any{"text": "西安机械", "ref": refOf(got, `textbox "客户"`)})
	call("type", map[string]any{"text": "4200", "ref": refOf(got, `textbox "金额"`)})
	call("click", map[string]any{"text": "待付款"})
	call("click", map[string]any{"text": "已付款"})
	call("click", map[string]any{"text": "现金"})
	expect(t, call("click", map[string]any{"text": "加急处理"}), `switch "加急处理" checked`, `radio "现金" checked`)
	expect(t, call("click", map[string]any{"text": "保存订单"}), `tab "订单列表" selected`, `row "SO-1037 | 西安机械 | 已付款 | 4200.00" selected`)

	expect(t, call("click", map[string]any{"text": "统计"}), "订单 36 笔", `progressbar "待付款 8 笔" value="22%"`, `progressbar "已付款 10 笔" value="28%"`)
}

// refOf finds the ref of the first line containing s in a snapshot.
func refOf(snapshot, s string) string {
	for _, line := range strings.Split(snapshot, "\n") {
		if strings.Contains(line, s) {
			return strings.Fields(line)[0]
		}
	}
	return ""
}

// TestChat streams Markdown answers like an AI assistant: the agent sends a
// prompt, waits for the streamed answer, reads its code block and table, and
// stops a second answer halfway.
func TestChat(t *testing.T) {
	if testing.Short() {
		t.Skip("builds and runs the example app")
	}
	call, _ := startServer(t, nil)
	got := call("launch", map[string]any{"command": "go run ./examples/chat -delay 2ms", "dir": repoRoot()})
	expect(t, got, `textbox "消息"`, `button "发送"`)
	call("type", map[string]any{"text": "用 Go 写一个并发下载器", "ref": refOf(got, `textbox "消息"`)})
	expect(t, call("press_key", map[string]any{"key": "enter"}), `"正在回答…"`, `button "停止"`)
	got = call("wait_for", map[string]any{"text": "下载器完成", "timeout_ms": 30000})
	expect(t, got, `table "" value="3 行"`, `row "8 | 5.6s | 推荐"`, `"在线"`, `button "发送"`)
	// Following the stream scrolled the chat to the end; scroll up to the code.
	call("scroll", map[string]any{"dy": -400})
	expect(t, call("snapshot", nil), `code "go"`, "func Download")

	call("type", map[string]any{"text": "对比一下缓存方案，列个表"})
	call("press_key", map[string]any{"key": "enter"})
	got = call("click", map[string]any{"text": "停止"})
	got = call("wait_for", map[string]any{"text": "已停止", "timeout_ms": 10000})
	expect(t, got, `button "发送"`)
}
