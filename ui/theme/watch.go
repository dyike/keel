package theme

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/dyike/keel/ui/internal/loop"
)

var current = "light"

// Use applies the registered palette name and remembers it, so CurrentName
// reports it and a ThemeWatcher reapplies it when its file changes. Call it
// like Apply: before opening windows or under the frame lock.
func Use(name string) error {
	p, ok := Named(name)
	if !ok {
		return fmt.Errorf("theme: unknown theme %q", name)
	}
	current = name
	Apply(p)
	return nil
}

// CurrentName is the theme last chosen with Use; "light" by default. Apply
// with an unnamed palette does not change it.
func CurrentName() string { return current }

// ThemeWatcher keeps the themes in a directory registered while their files
// change, e.g. so a designer sees a theme file update the running app.
type ThemeWatcher struct {
	dir      string
	onChange func(names []string, err error)
	stop     chan struct{}
	once     sync.Once

	mu    sync.Mutex
	seen  map[string]time.Time // file modification times
	names map[string]string    // theme name by file
}

// WatchThemes registers every *.json theme file in dir now, then checks the
// directory every interval (default one second) for added, changed or
// removed files. Changes are registered on the next frame, under the frame
// lock; if the theme in use (CurrentName) changed, it is applied again.
// onChange, if set, runs then too with the changed theme names, or with
// the error of a file that failed to parse; the other files still load.
//
// It polls instead of using file system events, so it needs nothing but the
// standard library and works the same on every platform.
func WatchThemes(dir string, interval time.Duration, onChange func(names []string, err error)) (*ThemeWatcher, error) {
	if _, err := os.Stat(dir); err != nil {
		return nil, err
	}
	if interval <= 0 {
		interval = time.Second
	}
	w := &ThemeWatcher{dir: dir, onChange: onChange, stop: make(chan struct{}), seen: map[string]time.Time{}, names: map[string]string{}}
	changes, err := w.scan()
	// The first load happens now, so the themes exist before any window.
	w.apply(changes, err, false)
	go func() {
		t := time.NewTicker(interval)
		defer t.Stop()
		for {
			select {
			case <-w.stop:
				return
			case <-t.C:
				if changes, err := w.scan(); len(changes) > 0 || err != nil {
					loop.Post(func() { w.apply(changes, err, true) })
				}
			}
		}
	}()
	return w, nil
}

// Stop ends the watching; registered themes stay.
func (w *ThemeWatcher) Stop() { w.once.Do(func() { close(w.stop) }) }

// themeChange is one theme to register: a parsed palette, or a removed one.
type themeChange struct {
	name    string
	palette Palette
}

// scan reads the files that changed since the last scan.
func (w *ThemeWatcher) scan() ([]themeChange, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	entries, err := os.ReadDir(w.dir)
	if err != nil {
		return nil, err
	}
	var changes []themeChange
	var firstErr error
	present := map[string]bool{}
	for _, e := range entries {
		if e.IsDir() || !strings.EqualFold(filepath.Ext(e.Name()), ".json") {
			continue
		}
		file := filepath.Join(w.dir, e.Name())
		present[file] = true
		info, err := e.Info()
		if err != nil {
			continue
		}
		if t, ok := w.seen[file]; ok && t.Equal(info.ModTime()) {
			continue
		}
		w.seen[file] = info.ModTime()
		data, err := os.ReadFile(file)
		if err == nil {
			var name string
			var p Palette
			if name, p, err = ParseTheme(data); err == nil {
				w.names[file] = name
				changes = append(changes, themeChange{name, p})
				continue
			}
		}
		if firstErr == nil {
			firstErr = fmt.Errorf("theme file %s: %w", e.Name(), err)
		}
	}
	for file := range w.seen {
		if !present[file] {
			delete(w.seen, file)
			delete(w.names, file)
		}
	}
	return changes, firstErr
}

// apply registers changes; under the frame lock when frame is set.
func (w *ThemeWatcher) apply(changes []themeChange, err error, frame bool) {
	var names []string
	for _, c := range changes {
		Register(c.name, c.palette)
		names = append(names, c.name)
		if frame && c.name == current {
			Apply(c.palette)
		}
	}
	slices.Sort(names)
	if w.onChange != nil && (len(names) > 0 || err != nil) {
		w.onChange(names, err)
	}
}
