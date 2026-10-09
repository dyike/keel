package window

// Automation protocol: one JSON object per line over the KEEL_AUTOMATION unix
// socket. Clients connect one at a time; when one disconnects the app keeps
// running and accepts the next, so an agent can attach, detach and attach again. Request {"method": "...", "params": {...}}; response {"result": ...}
// or {"error": "..."}. Requests are handled one at a time on the main goroutine.
//
// Every method takes an optional "window" (an id such as "w1"; default: the
// active window). Actions return the window's snapshot after they render.
//
//	windows                                   list open windows
//	snapshot                                  elements with refs, roles, names, bounds
//	click      {ref | text | x,y}             left click at the element's center
//	type       {text, ref?, clear?}           focus ref (if given), optionally select all, insert text
//	press      {key}                          key chord: "enter", "tab", "mod+a", "esc"
//	scroll     {dy, ref? | x,y?}              wheel scroll; positive dy scrolls down (dp)
//	wait_for   {text, timeout_ms?}            render until some element contains text
//	screenshot                                PNG, base64, 1 px = 1 dp
//	close                                     close the window as its close button would

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/dyike/keel/third_party/gio/f32"
	"github.com/dyike/keel/third_party/gio/io/system"

	"github.com/dyike/keel/ui/internal/loop"
)

type request struct {
	Method string `json:"method"`
	Params params `json:"params"`
}

type params struct {
	Window    string   `json:"window"`
	Ref       string   `json:"ref"`
	Text      string   `json:"text"`
	X         *float32 `json:"x"`
	Y         *float32 `json:"y"`
	DY        float32  `json:"dy"`
	Key       string   `json:"key"`
	Clear     bool     `json:"clear"`
	TimeoutMS int      `json:"timeout_ms"`
}

type response struct {
	Result any    `json:"result,omitempty"`
	Error  string `json:"error,omitempty"`
}

// WindowInfo describes an open window.
type WindowInfo struct {
	ID     string `json:"id"`
	Title  string `json:"title"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
	Active bool   `json:"active"`
}

// Snapshot is a window and its elements.
type Snapshot struct {
	Window   WindowInfo `json:"window"`
	Elements []Element  `json:"elements"`
}

func serveAutomation() {
	if err := os.MkdirAll(filepath.Dir(auto.addr), 0o700); err != nil {
		fmt.Fprintln(os.Stderr, "keel automation:", err)
		os.Exit(1)
	}
	os.Remove(auto.addr)
	ln, err := net.Listen("unix", auto.addr)
	if err != nil {
		fmt.Fprintln(os.Stderr, "keel automation:", err)
		os.Exit(1)
	}
	defer os.Remove(auto.addr)
	// Remove the socket on Ctrl-C too, so attach does not find a dead app.
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sig
		os.Remove(auto.addr)
		os.Exit(130)
	}()
	fmt.Fprintln(os.Stderr, "keel automation: listening on", auto.addr)
	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Fprintln(os.Stderr, "keel automation:", err)
			os.Exit(1)
		}
		serveConn(conn)
	}
}

func serveConn(conn net.Conn) {
	defer conn.Close()
	in := bufio.NewScanner(conn)
	in.Buffer(make([]byte, 1<<20), 1<<20)
	out := json.NewEncoder(conn)
	for in.Scan() {
		var req request
		var resp response
		if err := json.Unmarshal(in.Bytes(), &req); err != nil {
			resp.Error = "bad request: " + err.Error()
		} else if res, err := handle(req); err != nil {
			resp.Error = err.Error()
		} else {
			resp.Result = res
		}
		if err := out.Encode(resp); err != nil {
			return
		}
		auto.mu.Lock()
		empty := len(auto.windows) == 0
		auto.mu.Unlock()
		if empty && offScreen() { // like a real app: the last window closed, so exit
			os.Remove(auto.addr)
			os.Exit(0)
		}
		// On screen, Window.destroy exits a moment later instead: exiting
		// while AppKit is still tearing the window down can crash.
	}
}

func handle(req request) (any, error) {
	p := req.Params
	if req.Method == "windows" {
		return windowInfos(), nil
	}
	w, err := target(p.Window)
	if err != nil {
		return nil, err
	}
	switch req.Method {
	case "snapshot":
		return w.snapshotResult(), nil
	case "click":
		pt, err := w.point(p)
		if err != nil {
			return nil, err
		}
		w.click(pt)
	case "type":
		if p.Ref != "" {
			e, err := w.find(p.Ref, "")
			if err != nil {
				return nil, err
			}
			w.click(e.center())
		}
		if p.Clear {
			if err := w.press("mod+a"); err != nil {
				return nil, err
			}
		}
		if err := w.typeText(p.Text); err != nil {
			return nil, err
		}
	case "press":
		if err := w.press(p.Key); err != nil {
			return nil, err
		}
	case "scroll":
		size := w.virt.getSize()
		pt := f32.Pt(float32(size.X)/2, float32(size.Y)/2)
		if p.Ref != "" || p.Text != "" || p.X != nil {
			if pt, err = w.point(p); err != nil {
				return nil, err
			}
		}
		w.scroll(pt, p.DY)
	case "wait_for":
		if p.Text == "" {
			return nil, errors.New("wait_for needs text")
		}
		timeout := time.Duration(p.TimeoutMS) * time.Millisecond
		if timeout <= 0 {
			timeout = 5 * time.Second
		}
		for deadline := time.Now().Add(timeout); ; time.Sleep(50 * time.Millisecond) {
			for _, e := range w.snapshot() {
				if strings.Contains(e.Name, p.Text) || strings.Contains(e.Value, p.Text) {
					return w.snapshotResult(), nil
				}
			}
			if time.Now().After(deadline) {
				return nil, fmt.Errorf("timed out after %v waiting for %q", timeout, p.Text)
			}
		}
	case "screenshot":
		return w.screenshot()
	case "close":
		if w.win == nil { // headless: only the twin exists
			queueAction(w, system.ActionClose)
			runActions()
		} else if err := closeReal(w); err != nil {
			return nil, err
		}
		return windowInfos(), nil
	default:
		return nil, fmt.Errorf("unknown method %q", req.Method)
	}
	auto.mu.Lock()
	if !w.closed {
		auto.active = w
	}
	auto.mu.Unlock()
	if w.closed { // the action closed its own window
		return windowInfos(), nil
	}
	return w.snapshotResult(), nil
}

func (w *Window) snapshotResult() Snapshot {
	els := w.snapshot()
	return Snapshot{Window: w.info(), Elements: els}
}

// point resolves where an action lands: an element's center, or explicit x,y.
func (w *Window) point(p params) (f32.Point, error) {
	if p.X != nil && p.Y != nil {
		return f32.Pt(*p.X, *p.Y), nil
	}
	if p.Ref == "" && p.Text == "" {
		return f32.Point{}, errors.New("pass ref, text, or x and y")
	}
	e, err := w.find(p.Ref, p.Text)
	if err != nil {
		return f32.Point{}, err
	}
	return e.center(), nil
}

func target(id string) (*Window, error) {
	auto.mu.Lock()
	defer auto.mu.Unlock()
	if id == "" {
		if auto.active == nil {
			return nil, errors.New("no open window")
		}
		return auto.active, nil
	}
	for _, w := range auto.windows {
		if w.virt.id == id {
			return w, nil
		}
	}
	return nil, fmt.Errorf("no open window %q", id)
}

func windowInfos() []WindowInfo {
	auto.mu.Lock()
	ws := append([]*Window(nil), auto.windows...)
	auto.mu.Unlock()
	out := make([]WindowInfo, 0, len(ws))
	for _, w := range ws {
		out = append(out, w.info())
	}
	return out
}

func (w *Window) info() WindowInfo {
	auto.mu.Lock()
	active := auto.active == w
	auto.mu.Unlock()
	size := w.virt.getSize()
	return WindowInfo{ID: w.virt.id, Title: w.opts.Title, Width: size.X, Height: size.Y, Active: active}
}

// closeReal closes an on-screen window and waits until it is destroyed, so the
// reply reflects the screen. Closing only the twin would leave the window up.
func closeReal(w *Window) error {
	w.perform(system.ActionClose)
	for deadline := time.Now().Add(5 * time.Second); ; time.Sleep(20 * time.Millisecond) {
		loop.Lock()
		closed := w.closed
		loop.Unlock()
		if closed {
			return nil
		}
		if time.Now().After(deadline) {
			return errors.New("the window did not close within 5s")
		}
	}
}
