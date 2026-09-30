package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"
)

// appProcess is the app under test and the socket connection to it. The app
// is either a child started by launch, or one the user started and attached to;
// only a child is ever killed.
type appProcess struct {
	mu       sync.Mutex
	attached string // socket of an app the user started; empty for a child
	cmd      *exec.Cmd
	dir      string // temp dir holding the socket
	conn     net.Conn
	in       *bufio.Reader
	exited   chan struct{} // closed when the process ends
	exitErr  error         // written before exited closes; read only after
	logs     ringLog
	seen     map[string]bool // window ids reported so far, to flag new windows
}

// launch stops any running app, starts command through the shell and waits
// until the app listens. Compiling with go run can take a while.
func (a *appProcess) launch(command, workdir string, timeout time.Duration, visible bool) error {
	a.stop()
	a.mu.Lock()
	defer a.mu.Unlock()

	// Unix socket paths are limited to about 100 bytes, so use a short temp dir.
	dir, err := os.MkdirTemp("", "keel")
	if err != nil {
		return err
	}
	sock := filepath.Join(dir, "app.sock")
	cmd := exec.Command("sh", "-c", command)
	cmd.Dir = workdir
	cmd.Env = append(os.Environ(), "KEEL_AUTOMATION="+sock)
	if !visible {
		cmd.Env = append(cmd.Env, "KEEL_HEADLESS=1")
	}
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true} // stop kills go run's child too
	a.logs.reset()
	cmd.Stdout = &a.logs
	cmd.Stderr = &a.logs
	if err := cmd.Start(); err != nil {
		os.RemoveAll(dir)
		return err
	}
	a.cmd, a.dir, a.seen = cmd, dir, map[string]bool{}
	a.exited = make(chan struct{})
	go func(exited chan struct{}) {
		a.exitErr = cmd.Wait() // no lock: holders of a.mu wait on exited
		close(exited)
	}(a.exited)

	deadline := time.Now().Add(timeout)
	for {
		select {
		case <-a.exited:
			return fmt.Errorf("app exited before it was ready (%v)\n%s", a.exitErr, a.logs.tail(40))
		default:
		}
		if conn, err := net.Dial("unix", sock); err == nil {
			a.conn, a.in = conn, bufio.NewReaderSize(conn, 1<<20)
			return nil
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("app did not open %s within %v; did it call window.Main?\n%s", sock, timeout, a.logs.tail(40))
		}
		time.Sleep(100 * time.Millisecond)
	}
}

// stop kills a launched app, or only disconnects from an attached one.
func (a *appProcess) stop() {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.conn != nil {
		a.conn.Close()
		a.conn = nil
	}
	a.attached = ""
	if a.cmd != nil {
		select {
		case <-a.exited:
		default:
			syscall.Kill(-a.cmd.Process.Pid, syscall.SIGKILL)
			<-a.exited
		}
		a.cmd = nil
	}
	if a.dir != "" {
		os.RemoveAll(a.dir)
		a.dir = ""
	}
}

var errNotRunning = errors.New("not connected to an app: call launch or attach first")

// socketDir must match ui/window.SocketDir, where apps started with
// KEEL_AUTOMATION=1 listen. (This command does not import Keel.)
func socketDir() string { return filepath.Join(os.TempDir(), "keel") }

// liveSockets lists apps listening in socketDir, removing sockets left by apps
// that died without cleaning up.
func liveSockets() []string {
	paths, _ := filepath.Glob(filepath.Join(socketDir(), "*.sock"))
	var live []string
	for _, p := range paths {
		if c, err := net.DialTimeout("unix", p, time.Second); err == nil {
			c.Close()
			live = append(live, p)
		} else {
			os.Remove(p)
		}
	}
	return live
}

// attach connects to an app the user started. With no socket it picks the
// only running app, or lists them. Connecting while another client is attached
// waits until that client leaves, since the app serves one client at a time.
func (a *appProcess) attach(sock string) (string, error) {
	if sock == "" {
		live := liveSockets()
		switch len(live) {
		case 0:
			return "", fmt.Errorf("no running app found in %s; start one with KEEL_AUTOMATION=1, or pass socket", socketDir())
		case 1:
			sock = live[0]
		default:
			return "", fmt.Errorf("several apps are running; pass socket as one of:\n%s", strings.Join(live, "\n"))
		}
	}
	a.stop()
	conn, err := net.DialTimeout("unix", sock, 5*time.Second)
	if err != nil {
		return "", fmt.Errorf("cannot connect to %s: %v", sock, err)
	}
	a.mu.Lock()
	a.conn, a.in, a.attached, a.seen = conn, bufio.NewReaderSize(conn, 1<<20), sock, map[string]bool{}
	a.mu.Unlock()
	if err := a.callTimeout("windows", nil, nil, 3*time.Second); err != nil {
		a.stop()
		if errors.Is(err, os.ErrDeadlineExceeded) {
			return "", fmt.Errorf("%s did not answer: another client (another agent session?) is attached; the app serves one at a time", sock)
		}
		return "", err
	}
	return sock, nil
}

// call sends one request and decodes the result into out. Requests time out
// after a minute, or after wait_for's own timeout plus that.
func (a *appProcess) call(method string, params map[string]any, out any) error {
	timeout := time.Minute
	if ms, ok := params["timeout_ms"].(int); ok && ms > 0 {
		timeout += time.Duration(ms) * time.Millisecond
	}
	return a.callTimeout(method, params, out, timeout)
}

func (a *appProcess) callTimeout(method string, params map[string]any, out any, timeout time.Duration) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.conn == nil {
		return errNotRunning
	}
	a.conn.SetDeadline(time.Now().Add(timeout))
	defer a.conn.SetDeadline(time.Time{})
	req, _ := json.Marshal(map[string]any{"method": method, "params": params})
	if _, err := a.conn.Write(append(req, '\n')); err != nil {
		return a.lost(err)
	}
	line, err := a.in.ReadBytes('\n')
	if errors.Is(err, os.ErrDeadlineExceeded) {
		// The reply may still arrive and would be read as the next answer, so drop the connection.
		a.conn.Close()
		a.conn, a.attached = nil, ""
		return fmt.Errorf("%s: no reply within %v: %w", method, timeout, err)
	}
	if err != nil {
		return a.lost(err)
	}
	var resp struct {
		Result json.RawMessage `json:"result"`
		Error  string          `json:"error"`
	}
	if err := json.Unmarshal(line, &resp); err != nil {
		return err
	}
	if resp.Error != "" {
		return errors.New(resp.Error)
	}
	if out != nil {
		return json.Unmarshal(resp.Result, out)
	}
	return nil
}

// lost explains a broken connection: usually the app exited or crashed.
func (a *appProcess) lost(err error) error {
	a.conn.Close()
	a.conn = nil
	if a.attached != "" {
		a.attached = ""
		return fmt.Errorf("connection to the attached app closed (%v): it exited, or its last window closed", err)
	}
	select {
	case <-a.exited:
	case <-time.After(2 * time.Second):
		return fmt.Errorf("lost connection to the app: %v\n%s", err, a.logs.tail(40))
	}
	if a.exitErr == nil {
		return fmt.Errorf("app exited normally (its last window closed)\n%s", a.logs.tail(20))
	}
	return fmt.Errorf("app exited: %v\n%s", a.exitErr, a.logs.tail(40))
}

// ringLog keeps the last lines of the app's stdout and stderr.
type ringLog struct {
	mu      sync.Mutex
	lines   []string
	partial string
}

const maxLogLines = 500

func (l *ringLog) Write(p []byte) (int, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	text := l.partial + string(p)
	parts := strings.Split(text, "\n")
	l.partial = parts[len(parts)-1]
	l.lines = append(l.lines, parts[:len(parts)-1]...)
	if over := len(l.lines) - maxLogLines; over > 0 {
		l.lines = l.lines[over:]
	}
	return len(p), nil
}

func (l *ringLog) reset() {
	l.mu.Lock()
	l.lines, l.partial = nil, ""
	l.mu.Unlock()
}

func (l *ringLog) tail(n int) string {
	l.mu.Lock()
	defer l.mu.Unlock()
	lines := l.lines
	if l.partial != "" {
		lines = append(lines[:len(lines):len(lines)], l.partial)
	}
	if len(lines) > n {
		lines = lines[len(lines)-n:]
	}
	if len(lines) == 0 {
		return "(no output)"
	}
	return strings.Join(lines, "\n")
}

var _ io.Writer = (*ringLog)(nil)
