package window

import (
	"encoding/json"
	"fmt"
	"image"
	"os"
	"path/filepath"
	"sync"
)

// SizeState remembers the last normal (not maximized/fullscreen) client size.
// Use a separate state file for each logical window. Load it before Open;
// native windows save it at close/quit, outside the frame lock.
type SizeState struct {
	mu   sync.Mutex
	path string
	size image.Point
}

// LoadSizeState reads an optional JSON state file. A missing file starts empty.
// On other errors it returns an empty usable state together with the error.
func LoadSizeState(path string) (*SizeState, error) {
	s := &SizeState{path: path}
	b, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return s, nil
	}
	if err != nil {
		return s, err
	}
	var size image.Point
	if err := json.Unmarshal(b, &size); err != nil {
		return s, err
	}
	if !validSavedSize(size) {
		return s, fmt.Errorf("window: invalid saved size %v", size)
	}
	s.size = size
	return s, nil
}
func validSavedSize(p image.Point) bool { return p.X > 0 && p.Y > 0 && p.X <= 32768 && p.Y <= 32768 }
func (s *SizeState) snapshot() image.Point {
	if s == nil {
		return image.Point{}
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.size
}
func (s *SizeState) remember(size image.Point) {
	if s == nil || !validSavedSize(size) {
		return
	}
	s.mu.Lock()
	s.size = size
	s.mu.Unlock()
}

// Save writes the most recently observed normal size atomically. It can also
// be called explicitly from background work; do not call it under a frame lock.
func (s *SizeState) Save() error {
	if s == nil {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.path == "" || !validSavedSize(s.size) {
		return nil
	}
	data, err := json.Marshal(s.size)
	if err != nil {
		return err
	}
	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	f, err := os.CreateTemp(dir, ".keel-size-*")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	if _, err = f.Write(data); err != nil {
		f.Close()
		return err
	}
	if err = f.Close(); err != nil {
		return err
	}
	return os.Rename(f.Name(), s.path)
}
func (w *Window) saveSizeState() {
	if err := w.opts.SizeState.Save(); err != nil {
		fmt.Fprintf(os.Stderr, "keel: save window size: %v\n", err)
	}
}

// initialWindowSize selects a comfortable first-run size, bounded by the
// compositor's suggested area. Restored/explicit sizes are only clamped.
func initialWindowSize(requested, available image.Point, automatic bool) image.Point {
	if available.X <= 0 || available.Y <= 0 {
		return requested
	}
	if automatic {
		requested = image.Pt(min(1600, available.X*3/4), min(1000, available.Y*3/4))
	}
	return image.Pt(max(1, min(requested.X, available.X)), max(1, min(requested.Y, available.Y)))
}
