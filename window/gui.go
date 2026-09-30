package window

import (
	"fmt"
	"sync"

	"github.com/go-gui-org/go-gui/gui"
	"github.com/go-gui-org/go-gui/gui/appinfo"
	"github.com/dyike/keel/capability"
	"github.com/dyike/keel/ui"
)

// Options configures one native desktop window. UI and View are mutually exclusive.
// GUI supplies advanced Go-Gui settings; nonzero common fields override it.
type Options struct {
	Title                          string
	Width, Height                  int
	Hidden, Frameless, Transparent bool
	UI                             *ui.Page
	View                           func(*gui.Window) gui.View
	GUI                            gui.WindowCfg
}

type Manager struct {
	mu              sync.Mutex
	app             *gui.App
	info            appinfo.Info
	initial         []*gui.Window
	pages           map[*ui.Page]bool
	running, closed bool
	pending         int
}

func NewManager(app *gui.App) (*Manager, error) {
	if app == nil {
		return nil, capability.ErrInvalidArgument
	}
	return &Manager{app: app, pages: make(map[*ui.Page]bool)}, nil
}

// SetAppInfo configures the identity for windows created subsequently. Set before Run.
func (m *Manager) SetAppInfo(info appinfo.Info) { m.mu.Lock(); m.info = info; m.mu.Unlock() }
func config(o Options) (gui.WindowCfg, error) {
	if o.Width < 0 || o.Height < 0 || (o.UI != nil && o.View != nil) {
		return gui.WindowCfg{}, capability.ErrInvalidArgument
	}
	if o.UI != nil {
		if err := o.UI.Validate(); err != nil {
			return gui.WindowCfg{}, err
		}
	}
	cfg := o.GUI
	if o.Title != "" {
		cfg.Title = o.Title
	}
	if cfg.Title == "" && o.UI != nil {
		cfg.Title = o.UI.Title()
	}
	if o.Width != 0 {
		cfg.Width = o.Width
	}
	if o.Height != 0 {
		cfg.Height = o.Height
	}
	if cfg.Width == 0 {
		cfg.Width = 640
	}
	if cfg.Height == 0 {
		cfg.Height = 420
	}
	if cfg.Width < 0 || cfg.Height < 0 {
		return gui.WindowCfg{}, capability.ErrInvalidArgument
	}
	if o.Frameless {
		cfg.Decorations = gui.DecorationNone
	}
	if o.Transparent {
		cfg.Transparent = true
	}
	return cfg, nil
}

// New creates a window before Run, or queues creation on the UI thread at runtime.
// Runtime creation is asynchronous: use OnReady to access Host and native controls.
func (m *Manager) New(o Options) (*Window, error) {
	cfg, err := config(o)
	if err != nil {
		return nil, err
	}
	m.mu.Lock()
	if m.closed {
		m.mu.Unlock()
		return nil, capability.ErrClosed
	}
	if o.UI != nil && m.pages[o.UI] {
		m.mu.Unlock()
		return nil, fmt.Errorf("%w: create a fresh Page for each window", capability.ErrConflict)
	}
	// Go-Gui has a 16-request pending buffer. Refuse overflow instead of silently dropping.
	if m.running && m.pending >= 16 {
		m.mu.Unlock()
		return nil, capability.ErrNotReady
	}
	var binding *ui.WindowBinding
	if o.UI != nil {
		binding, err = o.UI.ReserveWindow()
		if err != nil {
			m.mu.Unlock()
			return nil, err
		}
		m.pages[o.UI] = true
	}
	if cfg.AppInfo == (appinfo.Info{}) {
		cfg.AppInfo = m.info
	}
	d := &guiDriver{}
	w, _ := New(d)
	init := cfg.OnInit
	configureView := func(host *gui.Window) error {
		if o.UI != nil {
			if err := binding.Attach(host); err != nil {
				return err
			}
		}
		if o.View != nil {
			host.SetView(o.View)
		}
		if o.UI != nil {
			go func() {
				<-host.Ctx().Done()
				o.UI.Close()
				m.mu.Lock()
				delete(m.pages, o.UI)
				m.mu.Unlock()
			}()
		}
		return nil
	}
	running := m.running
	cfg.OnInit = func(host *gui.Window) {
		if running {
			viewErr := configureView(host)
			m.mu.Lock()
			m.pending--
			m.mu.Unlock()
			if viewErr != nil {
				d.bind(host)
				host.Close()
				return
			}
		}
		d.bind(host)
		if init != nil {
			init(host)
		}
		if o.Hidden {
			host.Hide()
		}
		d.markReady()
	}
	if running {
		m.pending++
		m.mu.Unlock()
		m.app.OpenWindow(cfg)
	} else {
		host := gui.NewWindow(cfg)
		if err := configureView(host); err != nil {
			m.mu.Unlock()
			host.WindowCleanup()
			return nil, err
		}
		d.bind(host)
		m.initial = append(m.initial, host)
		m.mu.Unlock()
	}
	return w, nil
}

// BeginRun freezes the startup window list. A manager can run only once.
func (m *Manager) BeginRun() ([]*gui.Window, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.closed {
		return nil, capability.ErrClosed
	}
	if m.running {
		return nil, capability.ErrConflict
	}
	if len(m.initial) == 0 {
		return nil, fmt.Errorf("%w: create a window before Run", capability.ErrNotReady)
	}
	m.running = true
	return append([]*gui.Window(nil), m.initial...), nil
}

// EndRun releases startup windows that never acquired a native backend (for
// example when cgo is disabled). Call on the UI thread after the event loop ends.
func (m *Manager) EndRun() {
	m.mu.Lock()
	initial := append([]*gui.Window(nil), m.initial...)
	m.mu.Unlock()
	for _, w := range initial {
		if w.NativePlatformBackend() == nil {
			w.WindowCleanup()
		}
	}
}

// Shutdown rejects new windows and closes page state. Call on the UI thread after Run.
func (m *Manager) Shutdown() {
	m.mu.Lock()
	m.closed = true
	pages := make([]*ui.Page, 0, len(m.pages))
	for p := range m.pages {
		pages = append(pages, p)
	}
	initial := append([]*gui.Window(nil), m.initial...)
	running := m.running
	m.mu.Unlock()
	for _, p := range pages {
		p.Close()
	}
	if !running {
		for _, host := range initial {
			host.WindowCleanup()
		}
	}
}

// Dispatch queues a Go callback on the UI thread. It can be called from workers
// or global shortcut callbacks, and uses a surviving window to wake the loop.
// A nil return means queued; a request is dropped if the app closes before execution.
func (m *Manager) Dispatch(fn func()) error {
	if fn == nil {
		return capability.ErrInvalidArgument
	}
	m.mu.Lock()
	if m.closed {
		m.mu.Unlock()
		return capability.ErrClosed
	}
	if !m.running {
		m.mu.Unlock()
		return capability.ErrNotReady
	}
	initial := append([]*gui.Window(nil), m.initial...)
	m.mu.Unlock()
	hosts := append(m.app.Windows(), initial...)
	for _, host := range hosts {
		if host.Ctx().Err() != nil {
			continue
		}
		host.QueueCommand(func(*gui.Window) {
			m.mu.Lock()
			closed := m.closed
			m.mu.Unlock()
			if !closed {
				fn()
			}
		})
		return nil
	}
	return capability.ErrNotReady
}

// Quit schedules cleanup and closes all native windows on their UI thread.
func (m *Manager) Quit(cleanup func()) error {
	return m.Dispatch(func() {
		if cleanup != nil {
			cleanup()
		}
		for _, w := range m.app.Windows() {
			w.Close()
		}
	})
}

type guiDriver struct {
	mu        sync.Mutex
	host      *gui.Window
	ready     bool
	callbacks []func()
}

func (d *guiDriver) bind(w *gui.Window) { d.mu.Lock(); d.host = w; d.mu.Unlock() }
func (d *guiDriver) get() *gui.Window   { d.mu.Lock(); defer d.mu.Unlock(); return d.host }
func (d *guiDriver) markReady() {
	d.mu.Lock()
	if d.ready {
		d.mu.Unlock()
		return
	}
	d.ready = true
	callbacks := d.callbacks
	d.callbacks = nil
	d.mu.Unlock()
	for _, f := range callbacks {
		f()
	}
}
func (d *guiDriver) onReady(f func()) {
	d.mu.Lock()
	if !d.ready {
		d.callbacks = append(d.callbacks, f)
		d.mu.Unlock()
		return
	}
	host := d.host
	d.mu.Unlock()
	host.QueueCommand(func(*gui.Window) { f() })
}
func (d *guiDriver) enqueue(fn func(*gui.Window)) error {
	d.mu.Lock()
	host, ready := d.host, d.ready
	d.mu.Unlock()
	if !ready {
		return capability.ErrNotReady
	}
	if host.Ctx().Err() != nil {
		return capability.ErrClosed
	}
	host.QueueCommand(fn)
	return nil
}
func (d *guiDriver) Show() error { return d.enqueue(func(w *gui.Window) { w.Show() }) }
func (d *guiDriver) Hide() error { return d.enqueue(func(w *gui.Window) { w.Hide() }) }
func (d *guiDriver) Toggle() error {
	return d.enqueue(func(w *gui.Window) {
		if w.IsVisible() {
			w.Hide()
		} else {
			w.Show()
		}
	})
}
func (d *guiDriver) Close() error {
	host := d.get()
	if host == nil {
		return capability.ErrNotReady
	}
	host.Close()
	return nil
}

// Go-Gui v0.82 exposes neither native handles nor these window setters.
func (d *guiDriver) SetAlwaysOnTop(bool) error    { return capability.ErrUnsupported }
func (d *guiDriver) SetClickThrough(bool) error   { return capability.ErrUnsupported }
func (d *guiDriver) SetVibrancy(v Vibrancy) error { return setGUIVibrancy(d, v) }

// Host exposes advanced Go-Gui APIs. It is nil for pending runtime windows.
// Use raw host setters only on the UI thread; use QueueCommand from workers.
func (w *Window) Host() *gui.Window {
	if d, ok := w.driver.(*guiDriver); ok {
		return d.get()
	}
	return nil
}
func (w *Window) OnReady(fn func()) error {
	if fn == nil {
		return capability.ErrInvalidArgument
	}
	if d, ok := w.driver.(*guiDriver); ok {
		if host := d.get(); host != nil && host.Ctx().Err() != nil {
			return capability.ErrClosed
		}
		d.onReady(fn)
		return nil
	}
	return capability.ErrUnsupported
}
func (w *Window) Close() error {
	if d, ok := w.driver.(interface{ Close() error }); ok {
		return d.Close()
	}
	return capability.ErrUnsupported
}
