package window

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"reflect"
	"sync"

	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/internal/loop"
)

// A keel run reload restarts the app with the state it had: keel run passes
// KEEL_RUN_STATE, a file the app writes as it exits for a reload and reads
// (then removes) as it starts, and KEEL_RUN_CONTROL, a socket over which keel
// run asks for the reload, so it also works on Windows, which has no SIGINT
// to send to a GUI process. Older keel run versions set neither and restart
// without state.

type reloadState struct {
	Windows []reloadWindow `json:"windows"`
}

// reloadWindow is one window's state. ID is its title and its rank among
// open windows with that title: "Notes#0".
type reloadWindow struct {
	ID     string          `json:"id"`
	Width  int             `json:"width,omitempty"`
	Height int             `json:"height,omitempty"`
	Origin *[2]float64     `json:"origin,omitempty"` // macOS frame origin, in points
	App    json.RawMessage `json:"app,omitempty"`    // Options.KeepState
	UI     json.RawMessage `json:"ui,omitempty"`     // core.ReloadState of Content
}

var reload struct {
	sync.Mutex
	once   sync.Once
	saved  map[string]reloadWindow
	titles map[string]int // windows opened so far per title, for IDs
}

// loadReloadState reads and removes the state of the previous run.
func loadReloadState() {
	reload.once.Do(func() {
		path := os.Getenv("KEEL_RUN_STATE")
		if path == "" {
			return
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return
		}
		// Consumed: a crash before the next reload must not restore it again.
		os.Remove(path)
		var s reloadState
		if json.Unmarshal(b, &s) != nil {
			return
		}
		reload.saved = make(map[string]reloadWindow, len(s.Windows))
		for _, w := range s.Windows {
			reload.saved[w.ID] = w
		}
	})
}

// restoreWindow names w and gives it the state it had before the reload,
// before its native window exists.
func restoreWindow(w *Window) {
	loadReloadState()
	reload.Lock()
	if reload.titles == nil {
		reload.titles = map[string]int{}
	}
	w.reloadID = fmt.Sprintf("%s#%d", w.opts.Title, reload.titles[w.opts.Title])
	reload.titles[w.opts.Title]++
	saved, ok := reload.saved[w.reloadID]
	delete(reload.saved, w.reloadID)
	reload.Unlock()
	if !ok {
		return
	}
	if saved.Width > 0 && saved.Height > 0 {
		w.opts.Width, w.opts.Height = saved.Width, saved.Height
	}
	w.reloadOrigin = saved.Origin
	if w.opts.KeepState != nil && len(saved.App) > 0 {
		if err := decodeAll(saved.App, w.opts.KeepState); err != nil {
			fmt.Fprintf(os.Stderr, "keel: window %q: KeepState not restored: %v\n", w.opts.Title, err)
		}
	}
	if r, ok := w.opts.Content.(core.ReloadState); ok && len(saved.UI) > 0 {
		r.RestoreReloadState(saved.UI)
	}
}

// decodeAll decodes b into the value ptr points to, all or nothing: a field
// whose type changed in the code since the save leaves the defaults whole.
func decodeAll(b []byte, ptr any) error {
	v := reflect.ValueOf(ptr)
	if v.Kind() != reflect.Pointer || v.IsNil() {
		return fmt.Errorf("KeepState is a %T, not a pointer", ptr)
	}
	fresh := reflect.New(v.Elem().Type())
	fresh.Elem().Set(v.Elem())
	if err := json.Unmarshal(b, fresh.Interface()); err != nil {
		return err
	}
	v.Elem().Set(fresh.Elem())
	return nil
}

// saveReloadState writes the open windows' state for the next run. It takes
// the frame lock itself, after asking AppKit for window positions, which
// must not happen under the lock.
func saveReloadState(windows []*Window) {
	path := os.Getenv("KEEL_RUN_STATE")
	if path == "" {
		return
	}
	loop.Lock()
	views := make([]uintptr, len(windows))
	for i, w := range windows {
		if !w.closed {
			views[i] = w.nativeView
		}
	}
	loop.Unlock()
	origins := make([]*[2]float64, len(windows))
	for i, view := range views {
		if view != 0 {
			origins[i] = windowOrigin(view)
		}
	}
	var s reloadState
	loop.Lock()
	for i, w := range windows {
		if w.closed {
			continue
		}
		rw := reloadWindow{ID: w.reloadID, Origin: origins[i]}
		if size := w.lastSize; size.X > 0 && size.Y > 0 {
			rw.Width, rw.Height = size.X, size.Y
		}
		if w.opts.KeepState != nil {
			if b, err := json.Marshal(w.opts.KeepState); err == nil {
				rw.App = b
			} else {
				fmt.Fprintf(os.Stderr, "keel: window %q: KeepState not saved: %v\n", w.opts.Title, err)
			}
		}
		if r, ok := w.opts.Content.(core.ReloadState); ok {
			rw.UI = r.SaveReloadState()
		}
		s.Windows = append(s.Windows, rw)
	}
	loop.Unlock()
	b, err := json.Marshal(s)
	if err != nil {
		return
	}
	// Write then rename, so the next run never reads half a file.
	tmp := path + ".tmp"
	if os.WriteFile(tmp, b, 0o600) == nil {
		os.Rename(tmp, filepath.Clean(path))
	}
}

// listenForReload connects to keel run's control socket. A "reload" line
// saves state, closes the windows as for SIGINT and exits.
func listenForReload(reloadNow func()) {
	addr := os.Getenv("KEEL_RUN_CONTROL")
	if addr == "" {
		return
	}
	conn, err := net.Dial("unix", addr)
	if err != nil {
		return
	}
	fmt.Fprintf(conn, "hello %d\n", os.Getpid())
	go func() {
		defer conn.Close()
		sc := bufio.NewScanner(conn)
		for sc.Scan() {
			if sc.Text() == "reload" {
				reloadNow()
				return
			}
		}
	}()
}
