// Package ui composes desktop controls in Go and renders them with Go-Gui.
package ui

import (
	"sync"
	"sync/atomic"
)

// Component can be composed from Keel's primitives; reusable Go functions can return it.
type Component interface{ node() *element }
type element struct {
	mu                                                            sync.RWMutex
	id, kind, text, value, label, placeholder, inputType, variant string
	disabled, readOnly, required, checked                         bool
	maxLength                                                     int
	children                                                      []Component
	click                                                         func()
	change                                                        func(string)
	toggle                                                        func(bool)
}

func (e *element) node() *element { return e }

var nextID atomic.Uint64

func newElement(kind string) *element {
	return &element{id: "keel:" + base36(nextID.Add(1)), kind: kind}
}
func base36(n uint64) string {
	const chars = "0123456789abcdefghijklmnopqrstuvwxyz"
	if n == 0 {
		return "0"
	}
	s := []byte{}
	for n > 0 {
		s = append([]byte{chars[n%36]}, s...)
		n /= 36
	}
	return string(s)
}

// ID is stable for this component and can be used for accessibility and diagnostics.
func (e *element) ID() string         { return e.id }
func (e *element) SetDisabled(v bool) { e.mu.Lock(); e.disabled = v; e.mu.Unlock() }
func (e *element) Disabled() bool     { e.mu.RLock(); defer e.mu.RUnlock(); return e.disabled }

type TextComponent struct{ *element }

func Text(text string) *TextComponent {
	e := newElement("text")
	e.text = text
	return &TextComponent{e}
}
func Heading(text string) *TextComponent {
	e := newElement("heading")
	e.text = text
	return &TextComponent{e}
}
func (t *TextComponent) SetText(text string) { t.mu.Lock(); t.text = text; t.mu.Unlock() }
func (t *TextComponent) Text() string        { t.mu.RLock(); defer t.mu.RUnlock(); return t.text }

type ButtonVariant string

const (
	Primary   ButtonVariant = "primary"
	Secondary ButtonVariant = "secondary"
	Danger    ButtonVariant = "danger"
)

type ButtonOptions struct {
	Text     string
	Variant  ButtonVariant
	Disabled bool
	OnClick  func()
}
type ButtonComponent struct{ *element }

func Button(text string, onClick func()) *ButtonComponent {
	return NewButton(ButtonOptions{Text: text, OnClick: onClick})
}
func NewButton(o ButtonOptions) *ButtonComponent {
	e := newElement("button")
	e.text = o.Text
	e.variant = string(o.Variant)
	if e.variant == "" {
		e.variant = string(Primary)
	}
	e.disabled = o.Disabled
	e.click = o.OnClick
	return &ButtonComponent{e}
}
func (b *ButtonComponent) SetText(text string) { b.mu.Lock(); b.text = text; b.mu.Unlock() }

type InputType string

const (
	TextInput     InputType = "text"
	PasswordInput InputType = "password"
	EmailInput    InputType = "email"
	NumberInput   InputType = "number"
	SearchInput   InputType = "search"
)

type InputOptions struct {
	Label, Placeholder, Value    string
	Type                         InputType
	Disabled, ReadOnly, Required bool
	MaxLength                    int
	// OnChange runs on the UI thread after each accepted text change.
	OnChange func(string)
}
type InputComponent struct{ *element }

func Input(o InputOptions) *InputComponent    { return newInput("input", o) }
func TextArea(o InputOptions) *InputComponent { return newInput("textarea", o) }
func newInput(kind string, o InputOptions) *InputComponent {
	e := newElement(kind)
	e.label = o.Label
	e.placeholder = o.Placeholder
	e.value = o.Value
	e.inputType = string(o.Type)
	if e.inputType == "" {
		e.inputType = "text"
	}
	e.disabled = o.Disabled
	e.readOnly = o.ReadOnly
	e.required = o.Required
	e.maxLength = o.MaxLength
	e.change = o.OnChange
	return &InputComponent{e}
}
func (i *InputComponent) Value() string     { i.mu.RLock(); defer i.mu.RUnlock(); return i.value }
func (i *InputComponent) SetValue(v string) { i.mu.Lock(); i.value = v; i.mu.Unlock() }

type CheckboxOptions struct {
	Label             string
	Checked, Disabled bool
	OnChange          func(bool)
}
type CheckboxComponent struct{ *element }

func Checkbox(o CheckboxOptions) *CheckboxComponent {
	e := newElement("checkbox")
	e.label = o.Label
	e.checked = o.Checked
	e.disabled = o.Disabled
	e.toggle = o.OnChange
	return &CheckboxComponent{e}
}
func (c *CheckboxComponent) Value() bool     { c.mu.RLock(); defer c.mu.RUnlock(); return c.checked }
func (c *CheckboxComponent) SetValue(v bool) { c.mu.Lock(); c.checked = v; c.mu.Unlock() }

func container(kind string, children ...Component) Component {
	e := newElement(kind)
	e.children = append([]Component(nil), children...)
	return e
}
func Column(children ...Component) Component { return container("column", children...) }
func Row(children ...Component) Component    { return container("row", children...) }
func Card(children ...Component) Component   { return container("card", children...) }
func Divider() Component                     { return newElement("divider") }
