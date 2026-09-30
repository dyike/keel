package ui

import (
	"fmt"
	"sync"
	"sync/atomic"
	"unicode/utf8"

	"github.com/go-gui-org/go-gui/gui"
	"github.com/dyike/keel/capability"
)

// Page owns a fixed component tree and the mutable state of one window.
// Build a fresh Page (and fresh components) for each independent window.
type Page struct {
	title      string
	children   []Component
	index      map[string]*element
	validation error
	closed     atomic.Bool
	mu         sync.Mutex
	host       *gui.Window
	binding    *WindowBinding
	onError    func(error)
	message    string
}

func NewPage(title string, children ...Component) *Page {
	index, err := compile(children)
	return &Page{title: title, children: append([]Component(nil), children...), index: index, validation: err}
}
func (p *Page) Title() string { return p.title }
func (p *Page) Validate() error {
	if p.closed.Load() {
		return capability.ErrClosed
	}
	return p.validation
}
func (p *Page) SetOnError(handler func(error)) { p.mu.Lock(); p.onError = handler; p.mu.Unlock() }

// Attach installs this page on a Go-Gui window before its first frame.
// A Page cannot be attached to two different windows.
func (p *Page) Attach(w *gui.Window) error {
	if w == nil {
		return capability.ErrInvalidArgument
	}
	if err := p.Validate(); err != nil {
		return err
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed.Load() {
		return capability.ErrClosed
	}
	if p.binding != nil || (p.host != nil && p.host != w) {
		return capability.ErrConflict
	}
	p.host = w
	w.SetView(p.View)
	return nil
}

// WindowBinding reserves a Page for asynchronous native window creation.
// Hosts normally use Page.Attach; Keel's window manager reserves before queuing.
type WindowBinding struct{ page *Page }

func (p *Page) ReserveWindow() (*WindowBinding, error) {
	if err := p.Validate(); err != nil {
		return nil, err
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed.Load() {
		return nil, capability.ErrClosed
	}
	if p.binding != nil || p.host != nil {
		return nil, capability.ErrConflict
	}
	b := &WindowBinding{page: p}
	p.binding = b
	return b, nil
}
func (b *WindowBinding) Attach(w *gui.Window) error {
	if w == nil {
		return capability.ErrInvalidArgument
	}
	p := b.page
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed.Load() {
		return capability.ErrClosed
	}
	if p.binding != b || (p.host != nil && p.host != w) {
		return capability.ErrConflict
	}
	p.host = w
	w.SetView(p.View)
	return nil
}

// Refresh schedules a new frame, including when called from a worker goroutine.
// Event handlers refresh automatically. Setters alone do not wake the UI thread.
func (p *Page) Refresh() {
	p.mu.Lock()
	w := p.host
	p.mu.Unlock()
	if w != nil && !p.closed.Load() {
		w.QueueCommand(func(w *gui.Window) { w.InvalidateLayout() })
	}
}
func (p *Page) Close() { p.closed.Store(true); p.mu.Lock(); p.host = nil; p.mu.Unlock() }

func (p *Page) report(err error) {
	p.mu.Lock()
	p.message = err.Error()
	fn := p.onError
	p.mu.Unlock()
	if fn != nil {
		fn(err)
	}
}
func (p *Page) invoke(fn func()) {
	if p.closed.Load() {
		return
	}
	defer func() {
		if recover() != nil {
			p.report(fmt.Errorf("ui: event callback panicked"))
		}
		p.Refresh()
	}()
	p.mu.Lock()
	p.message = ""
	p.mu.Unlock()
	if fn != nil {
		fn()
	}
}
func (p *Page) validateFields() error {
	for _, e := range p.index {
		e.mu.RLock()
		kind, disabled, readOnly, required, value, label, max := e.kind, e.disabled, e.readOnly, e.required, e.value, e.label, e.maxLength
		e.mu.RUnlock()
		if (kind != "input" && kind != "textarea") || disabled || readOnly {
			continue
		}
		if required && value == "" {
			return fmt.Errorf("ui: %s is required", label)
		}
		if max > 0 && utf8.RuneCountInString(value) > max {
			return fmt.Errorf("ui: %s exceeds MaxLength", label)
		}
	}
	return nil
}

// View adapts the page to Go-Gui. Run it on the UI thread or in headless tests.
func (p *Page) View(w *gui.Window) gui.View {
	if p.closed.Load() {
		return gui.Label("Page closed", gui.TextStyle{})
	}
	if p.validation != nil {
		return gui.Label(p.validation.Error(), gui.TextStyle{})
	}
	views := make([]gui.View, 0, len(p.children)+1)
	for _, c := range p.children {
		views = append(views, p.render(c))
	}
	p.mu.Lock()
	message := p.message
	p.mu.Unlock()
	if message != "" {
		views = append(views, gui.Label(message, gui.CurrentTheme().TextStyleError))
	}
	return gui.Column(gui.ContainerCfg{ID: "keel-page", Sizing: gui.FillFill, Padding: gui.PaddingLarge, Spacing: gui.SpacingMedium, SizeBorder: gui.NoBorder, Scrollable: true, Content: views})
}
func (p *Page) render(c Component) gui.View {
	e := c.node()
	e.mu.RLock()
	kind, id, text, value, label, placeholder, inputType, variant := e.kind, e.id, e.text, e.value, e.label, e.placeholder, e.inputType, e.variant
	disabled, readOnly, checked, max := e.disabled, e.readOnly, e.checked, e.maxLength
	click, change, toggle := e.click, e.change, e.toggle
	e.mu.RUnlock()
	switch kind {
	case "heading":
		return gui.Text(gui.TextCfg{Text: text, TextStyle: gui.CurrentTheme().TextStyleTitle})
	case "text":
		return gui.Text(gui.TextCfg{Text: text, Sizing: gui.FillFit})
	case "button":
		v := gui.ButtonPrimary
		if variant == string(Secondary) {
			v = gui.ButtonSecondary
		}
		if variant == string(Danger) {
			v = gui.ButtonDanger
		}
		return gui.Button(gui.ButtonCfg{ID: id, Label: text, Variant: v, Disabled: disabled, OnClick: func(ctx gui.EventCtx) {
			ctx.Consume()
			if e.Disabled() {
				return
			}
			if err := p.validateFields(); err != nil {
				p.report(err)
				p.Refresh()
				return
			}
			p.invoke(click)
		}})
	case "input", "textarea":
		cfg := gui.InputCfg{ID: id, Label: label, Text: value, Placeholder: placeholder, Disabled: disabled, ReadOnly: readOnly, Sizing: gui.FillFit, IsPassword: inputType == string(PasswordInput)}
		if kind == "textarea" {
			cfg.Mode = gui.InputMultiline
			cfg.Height = 140
			cfg.Sizing = gui.FillFixed
			cfg.Scrollable = true
		}

		if inputType == string(EmailInput) {
			cfg.Keyboard = gui.KeyboardEmail
		}
		if inputType == string(NumberInput) {
			cfg.Keyboard = gui.KeyboardDecimal
		}
		cfg.PreTextChange = func(_, proposed string) (string, bool) {
			return proposed, max == 0 || utf8.RuneCountInString(proposed) <= max
		}
		cfg.OnTextChanged = func(v string, ctx gui.EventCtx) {
			if p.closed.Load() {
				return
			}
			e.mu.Lock()
			if e.disabled || e.readOnly {
				e.mu.Unlock()
				return
			}
			// Enforce the limit for undo/redo too, which bypasses PreTextChange.
			if max > 0 && utf8.RuneCountInString(v) > max {
				v = string([]rune(v)[:max])
			}
			e.value = v
			e.mu.Unlock()
			p.invoke(func() {
				if change != nil {
					change(v)
				}
			})
			ctx.Window.InvalidateLayout()
		}
		return gui.Input(cfg)
	case "checkbox":
		return gui.Checkbox(gui.ToggleCfg{ID: id, Label: label, Selected: checked, Disabled: disabled, OnClick: func(ctx gui.EventCtx) {
			ctx.Consume()
			if p.closed.Load() {
				return
			}
			e.mu.Lock()
			if e.disabled {
				e.mu.Unlock()
				return
			}
			e.checked = !e.checked
			v := e.checked
			e.mu.Unlock()
			p.invoke(func() {
				if toggle != nil {
					toggle(v)
				}
			})
		}})
	case "divider":
		return gui.Row(gui.ContainerCfg{Sizing: gui.FillFixed, Height: 1, Color: gui.CurrentTheme().ColorBorder, SizeBorder: gui.NoBorder})
	default:
		children := make([]gui.View, 0, len(e.children))
		for _, child := range e.children {
			children = append(children, p.render(child))
		}
		cfg := gui.ContainerCfg{Content: children, Sizing: gui.FillFit, Spacing: gui.SpacingMedium, SizeBorder: gui.NoBorder}
		if kind == "row" {
			cfg.Sizing = gui.FitFit
			return gui.Row(cfg)
		}
		if kind == "card" {
			cfg.Padding = gui.PaddingMedium
			cfg.SizeBorder = gui.BorderPx(1)
			cfg.ColorBorder = gui.CurrentTheme().ColorBorder
			cfg.Radius = gui.RadiusPx(8)
		}
		return gui.Column(cfg)
	}
}
