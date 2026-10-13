// Package tray provides desktop status icons and native popup menus without Gio
// or cgo. Callbacks run serially on a Go goroutine; write UI state with core.Update.
package tray

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	"strings"
	"sync"

	"github.com/dyike/keel/native"
	"github.com/dyike/keel/native/internal/sys"
)

// MenuItem is an action, submenu, or separator. Snapshots are copied. IDs are
// optional; when supplied they must be unique throughout the menu.
type MenuItem struct {
	ID, Title                    string
	Disabled, Checked, Separator bool
	Children                     []MenuItem
	OnSelect                     func()
}

type Options struct {
	Icon       image.Image
	Template   bool
	Tooltip    string
	Menu       []MenuItem
	OnActivate func()
}

// Tray owns a status icon. Methods are safe from concurrent goroutines. Callback
// execution is independent of the OS event thread and serialized per Tray.
type Tray struct {
	mu           sync.Mutex
	backend      sys.TrayBackend
	options      sys.TrayOptions
	actions      map[uint32]func()
	next         uint32
	closed       bool
	eventsMu     sync.Mutex
	events       []uint32
	eventsClosed bool
	wake         chan struct{}
}

// New creates a tray icon. macOS creation is synchronous on the main thread;
// call before window.Main, or from a goroutine outside the UI frame lock.
func New(options Options) (*Tray, error) { return newTray(options, sys.TrayOpen) }

func newTray(options Options, open func(sys.TrayOptions) (sys.TrayBackend, error)) (*Tray, error) {
	data, err := encodeIcon(options.Icon)
	if err != nil {
		return nil, err
	}
	if strings.ContainsRune(options.Tooltip, 0) {
		return nil, native.ErrInvalidArgument
	}
	t := &Tray{wake: make(chan struct{}, 1)}
	items, actions, err := t.prepareMenu(options.Menu)
	if err != nil {
		return nil, err
	}
	actions[0] = options.OnActivate
	t.actions = actions
	t.options = sys.TrayOptions{PNG: data, Template: options.Template, Tooltip: options.Tooltip, Menu: items, Event: t.event}
	t.backend, err = open(t.options)
	if err != nil {
		return nil, err
	}
	go t.run()
	return t, nil
}

func encodeIcon(icon image.Image) ([]byte, error) {
	if icon == nil {
		return nil, fmt.Errorf("%w: tray icon is required", native.ErrInvalidArgument)
	}
	size := icon.Bounds().Size()
	if size.X <= 0 || size.Y <= 0 || size.X > 1024 || size.Y > 1024 {
		return nil, fmt.Errorf("%w: tray icon must be 1..1024 pixels per axis", native.ErrInvalidArgument)
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, icon); err != nil {
		return nil, fmt.Errorf("%w: %v", native.ErrInvalidArgument, err)
	}
	return buf.Bytes(), nil
}

func (t *Tray) prepareMenu(items []MenuItem) ([]sys.TrayMenuItem, map[uint32]func(), error) {
	actions := make(map[uint32]func())
	ids := make(map[string]bool)
	count := 0
	var walk func([]MenuItem, int) ([]sys.TrayMenuItem, error)
	walk = func(items []MenuItem, depth int) ([]sys.TrayMenuItem, error) {
		if depth > 16 {
			return nil, fmt.Errorf("%w: tray menu too deep", native.ErrInvalidArgument)
		}
		result := make([]sys.TrayMenuItem, 0, len(items))
		for _, item := range items {
			count++
			if count > 1024 || strings.ContainsRune(item.Title, 0) || (item.ID != "" && ids[item.ID]) {
				return nil, native.ErrInvalidArgument
			}
			if item.ID != "" {
				ids[item.ID] = true
			}
			if item.Separator {
				if item.Title != "" || item.Checked || len(item.Children) != 0 || item.OnSelect != nil {
					return nil, native.ErrInvalidArgument
				}
			} else if strings.TrimSpace(item.Title) == "" || (len(item.Children) != 0 && item.OnSelect != nil) {
				return nil, native.ErrInvalidArgument
			}
			// DBusMenu uses signed 32-bit IDs; Windows version 4 icon IDs use 16 bits,
			// but menu command IDs are full UINTs.
			if t.next == 0x7fffffff {
				return nil, native.ErrFailed
			}
			t.next++
			id := t.next
			children, err := walk(item.Children, depth+1)
			if err != nil {
				return nil, err
			}
			result = append(result, sys.TrayMenuItem{ID: id, Title: item.Title, Disabled: item.Disabled, Checked: item.Checked, Separator: item.Separator, Children: children})
			if item.Disabled {
				var disable func([]sys.TrayMenuItem)
				disable = func(children []sys.TrayMenuItem) {
					for _, child := range children {
						delete(actions, child.ID)
						disable(child.Children)
					}
				}
				disable(children)
			}
			if !item.Disabled && !item.Separator && len(item.Children) == 0 {
				actions[id] = item.OnSelect
			}
		}
		return result, nil
	}
	menu, err := walk(items, 0)
	return menu, actions, err
}

// SetIcon replaces the icon, keeping the Template setting supplied to New.
func (t *Tray) SetIcon(icon image.Image) error {
	data, err := encodeIcon(icon)
	if err != nil {
		return err
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.closed {
		return native.ErrFailed
	}
	next := t.options
	next.PNG = data
	if err := t.backend.Update(next); err != nil {
		return err
	}
	t.options = next
	return nil
}

func (t *Tray) SetTooltip(text string) error {
	if strings.ContainsRune(text, 0) {
		return native.ErrInvalidArgument
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.closed {
		return native.ErrFailed
	}
	next := t.options
	next.Tooltip = text
	if err := t.backend.Update(next); err != nil {
		return err
	}
	t.options = next
	return nil
}

// SetMenu atomically replaces the menu. Validation errors preserve the old menu.
func (t *Tray) SetMenu(items []MenuItem) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.closed {
		return native.ErrFailed
	}
	menu, actions, err := t.prepareMenu(items)
	if err != nil {
		return err
	}
	next := t.options
	next.Menu = menu
	if err := t.backend.Update(next); err != nil {
		return err
	}
	actions[0] = t.actions[0]
	t.options, t.actions = next, actions
	return nil
}

// Close removes the icon and discards queued callbacks. It is idempotent and may
// be called from a callback. A callback already running is allowed to finish.
func (t *Tray) Close() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.closed {
		return nil
	}
	err := t.backend.Close()
	t.closed = true
	t.eventsMu.Lock()
	t.eventsClosed = true
	t.events = nil
	t.eventsMu.Unlock()
	t.actions = nil
	select {
	case t.wake <- struct{}{}:
	default:
	}
	return err
}

func (t *Tray) event(id uint32) {
	// This lock only protects ingress; no native calls or user callbacks take it.
	t.eventsMu.Lock()
	if !t.eventsClosed {
		t.events = append(t.events, id)
		select {
		case t.wake <- struct{}{}:
		default:
		}
	}
	t.eventsMu.Unlock()
}

func (t *Tray) run() {
	for range t.wake {
		t.eventsMu.Lock()
		ids := t.events
		t.events = nil
		closed := t.eventsClosed
		t.eventsMu.Unlock()
		if closed {
			return
		}
		for _, id := range ids {
			t.mu.Lock()
			fn := t.actions[id]
			closed := t.closed
			t.mu.Unlock()
			if closed {
				return
			}
			if fn != nil {
				fn()
			}
		}
	}
}

// Main runs a standalone tray app's platform loop. Do not call it together with
// window.Main. On macOS lock the main OS thread before New and call Main there.
func Main() { sys.TrayMain() }
