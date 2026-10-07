package window

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"

	"gioui.org/io/key"
	"github.com/dyike/keel/ui/core"
)

// MenuAction forwards a standard editing command to the focused native view.
// OnSelect, when supplied, overrides the action. Go editors consume actions
// through core.NextEditAction; AppKit also forwards to the focused native view.
type MenuAction string

const (
	MenuCopy      MenuAction = "copy"
	MenuCut       MenuAction = "cut"
	MenuPaste     MenuAction = "paste"
	MenuSelectAll MenuAction = "select-all"
	MenuUndo      MenuAction = "undo"
	MenuRedo      MenuAction = "redo"
)

// MenuRole assigns a macOS system menu slot without prescribing its contents.
type MenuRole string

const (
	MenuApplication MenuRole = "application"
	MenuEdit        MenuRole = "edit"
	MenuWindow      MenuRole = "window"
	MenuHelp        MenuRole = "help"
	MenuServices    MenuRole = "services"
)

// MenuItem describes a top-level menu, submenu, action or separator. IDs are
// optional, but explicit IDs permit UpdateItem and Invoke. Items are copied;
// callers may reuse slices. Callbacks run serially with other UI callbacks.
type MenuItem struct {
	ID, Title, Shortcut          string
	Disabled, Checked, Separator bool
	Role                         MenuRole
	Action                       MenuAction
	Children                     []MenuItem
	OnSelect                     func()
}

// MenuBar owns a portable application menu model. Install uses AppKit on macOS,
// Win32 menus on Windows, and Keel-rendered window menus on Linux X11/Wayland.
// Items can also drive an application's own Go menu renderer.
type MenuBar struct {
	mu    sync.Mutex
	items []MenuItem
}

var applicationMenu struct {
	sync.Mutex
	current    *MenuBar
	generation uint64
}

func NewMenuBar(items ...MenuItem) (*MenuBar, error) {
	normalized, err := prepareMenuItems(items)
	if err != nil {
		return nil, err
	}
	return &MenuBar{items: normalized}, nil
}

// NativeApplicationMenu reports whether this build has a native menu backend.
func NativeApplicationMenu() bool { return !offScreen() && platformNativeApplicationMenu() }

// SetApplicationMenu installs a menu without retaining a controller. Use
// NewMenuBar and Install when the menu will be updated dynamically.
func SetApplicationMenu(items ...MenuItem) error {
	bar, err := NewMenuBar(items...)
	if err != nil {
		return err
	}
	return bar.Install()
}

func (m *MenuBar) Install() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.installLocked()
}

func (m *MenuBar) installLocked() error {
	applicationMenu.Lock()
	defer applicationMenu.Unlock()
	applicationMenu.generation++
	data, err := json.Marshal(menuWire{applicationMenu.generation, wireMenuItems(m.items)})
	if err != nil {
		return err
	}
	applicationMenu.current = m
	if !offScreen() {
		platformInstallApplicationMenu(data)
	}
	// Refresh custom renderers and portable shortcut filters after background edits.
	go core.Update(func() {})
	return nil
}

// Items returns a deep copy, including Go callbacks for custom renderers.
func (m *MenuBar) Items() []MenuItem {
	m.mu.Lock()
	defer m.mu.Unlock()
	return cloneMenuItems(m.items)
}

// SetItems replaces the model atomically. Validation errors leave it unchanged.
func (m *MenuBar) SetItems(items ...MenuItem) error {
	normalized, err := prepareMenuItems(items)
	if err != nil {
		return err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.items = normalized
	if m.installed() {
		return m.installLocked()
	}
	return nil
}

// UpdateItem edits a copy of one item. The editor must not call methods on m.
// Titles, shortcuts, checked/disabled states, children and callbacks are mutable.
func (m *MenuBar) UpdateItem(id string, edit func(*MenuItem)) error {
	if edit == nil {
		return errors.New("window: nil menu item editor")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	items := cloneMenuItems(m.items)
	item, _ := findMenuItem(items, id, false)
	if item == nil {
		return fmt.Errorf("window: menu item %q not found", id)
	}
	edit(item)
	normalized, err := prepareMenuItems(items)
	if err != nil {
		return err
	}
	m.items = normalized
	if m.installed() {
		return m.installLocked()
	}
	return nil
}

func (m *MenuBar) installed() bool {
	applicationMenu.Lock()
	defer applicationMenu.Unlock()
	return applicationMenu.current == m
}

// Invoke activates an enabled leaf item from UI code, e.g. a custom menu view.
// It returns false for missing/disabled items or unavailable native edit actions.
func (m *MenuBar) Invoke(id string) bool {
	m.mu.Lock()
	item, disabled := findMenuItem(m.items, id, false)
	if item == nil || disabled || item.Separator || len(item.Children) > 0 {
		m.mu.Unlock()
		return false
	}
	callback, action := item.OnSelect, item.Action
	m.mu.Unlock()
	if callback != nil {
		callback()
		return true
	}
	if action != "" {
		if requestMenuEdit(action) {
			return true
		}
		return !offScreen() && platformMenuEdit(action)
	}
	return false
}

func dispatchNativeMenu(generation uint64, id string) {
	// Never enter Gio invalidation from AppKit's main-thread callback: it can
	// synchronously flush events while the invalidation lock is already held.
	go core.Update(func() { invokeNativeMenu(generation, id) })
}

func invokeNativeMenu(generation uint64, id string) {
	applicationMenu.Lock()
	bar := applicationMenu.current
	current := applicationMenu.generation == generation
	applicationMenu.Unlock()
	if bar != nil && current {
		bar.Invoke(id)
	}
}

func findMenuItem(items []MenuItem, id string, parentDisabled bool) (*MenuItem, bool) {
	for i := range items {
		item := &items[i]
		disabled := parentDisabled || item.Disabled
		if item.ID == id {
			return item, disabled
		}
		if found, blocked := findMenuItem(item.Children, id, disabled); found != nil {
			return found, blocked
		}
	}
	return nil, false
}

func cloneMenuItems(items []MenuItem) []MenuItem {
	out := append([]MenuItem(nil), items...)
	for i := range out {
		out[i].Children = cloneMenuItems(out[i].Children)
	}
	return out
}

func prepareMenuItems(items []MenuItem) ([]MenuItem, error) {
	out := cloneMenuItems(items)
	seen := map[string]bool{}
	var validate func([]MenuItem, string) error
	validate = func(items []MenuItem, path string) error {
		for i := range items {
			item := &items[i]
			where := fmt.Sprintf("%s/%d", path, i)
			if item.ID == "" {
				item.ID = "@keel" + where
			}
			if strings.ContainsRune(item.ID, '\x00') || strings.ContainsRune(item.Title, '\x00') {
				return fmt.Errorf("window: menu text contains NUL")
			}
			if seen[item.ID] {
				return fmt.Errorf("window: duplicate menu ID %q", item.ID)
			}
			seen[item.ID] = true
			if item.Separator {
				if item.OnSelect != nil || item.Action != "" || item.Shortcut != "" || len(item.Children) > 0 {
					return fmt.Errorf("window: separator %q has an action", item.ID)
				}
				continue
			}
			if strings.TrimSpace(item.Title) == "" {
				return fmt.Errorf("window: menu item %q has no title", item.ID)
			}
			switch item.Action {
			case "", MenuCopy, MenuCut, MenuPaste, MenuSelectAll, MenuUndo, MenuRedo:
			default:
				return fmt.Errorf("window: unknown menu action %q", item.Action)
			}
			switch item.Role {
			case "", MenuApplication, MenuEdit, MenuWindow, MenuHelp, MenuServices:
			default:
				return fmt.Errorf("window: unknown menu role %q", item.Role)
			}
			if item.Shortcut != "" {
				if _, _, err := core.ParseShortcut(item.Shortcut); err != nil {
					return err
				}
			}
			if len(item.Children) > 0 && (item.OnSelect != nil || item.Action != "") {
				return fmt.Errorf("window: submenu %q has a leaf action", item.ID)
			}
			if err := validate(item.Children, where); err != nil {
				return err
			}
		}
		return nil
	}
	return out, validate(out, "")
}

type menuWire struct {
	Generation uint64         `json:"generation"`
	Items      []menuWireItem `json:"items"`
}
type menuWireItem struct {
	ID        string         `json:"id"`
	Title     string         `json:"title"`
	Key       string         `json:"key"`
	Modifiers uint32         `json:"modifiers"`
	Disabled  bool           `json:"disabled"`
	Checked   bool           `json:"checked"`
	Separator bool           `json:"separator"`
	Role      MenuRole       `json:"role"`
	Action    MenuAction     `json:"action"`
	Children  []menuWireItem `json:"children,omitempty"`
}

func wireMenuItems(items []MenuItem) []menuWireItem {
	out := make([]menuWireItem, len(items))
	for i, item := range items {
		wire := menuWireItem{ID: item.ID, Title: item.Title, Disabled: item.Disabled, Checked: item.Checked, Separator: item.Separator, Role: item.Role, Action: item.Action, Children: wireMenuItems(item.Children)}
		if item.OnSelect != nil {
			wire.Action = ""
		}
		if item.Shortcut != "" {
			name, mods, _ := core.ParseShortcut(item.Shortcut)
			wire.Key = strings.ToLower(string(name))
			if mods&key.ModCommand != 0 {
				wire.Modifiers |= 1
			}
			if mods&key.ModCtrl != 0 {
				wire.Modifiers |= 2
			}
			if mods&key.ModAlt != 0 {
				wire.Modifiers |= 4
			}
			if mods&key.ModShift != 0 {
				wire.Modifiers |= 8
			}
		}
		out[i] = wire
	}
	return out
}

func applicationMenuShortcuts() []shortcut {
	applicationMenu.Lock()
	bar := applicationMenu.current
	applicationMenu.Unlock()
	if bar == nil || (!offScreen() && platformNativeMenuShortcuts()) {
		return nil
	}
	return menuShortcuts(bar)
}

func menuShortcuts(bar *MenuBar) []shortcut {
	var shortcuts []shortcut
	var collect func([]MenuItem, bool)
	collect = func(items []MenuItem, blocked bool) {
		for _, item := range items {
			disabled := blocked || item.Disabled
			if !disabled && item.Shortcut != "" && item.OnSelect != nil {
				binding, _ := parseShortcut(item.Shortcut)
				id := item.ID
				binding.fn = func() { bar.Invoke(id) }
				shortcuts = append(shortcuts, binding)
			}
			collect(item.Children, disabled)
		}
	}
	collect(bar.Items(), false)
	return shortcuts
}
