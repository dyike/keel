//go:build (linux && !android) || (darwin && !ios)

package sys

import (
	"bytes"
	"context"
	"fmt"
	"image/color"
	"image/png"
	"reflect"
	"sync"
	"sync/atomic"
	"time"

	"github.com/dyike/keel/native"
	"github.com/godbus/dbus/v5"
	"github.com/godbus/dbus/v5/introspect"
	"github.com/godbus/dbus/v5/prop"
)

const (
	traySNI                           = "org.kde.StatusNotifierItem"
	trayMenuInterface                 = "com.canonical.dbusmenu"
	trayItemPath      dbus.ObjectPath = "/StatusNotifierItem"
	trayMenuPath      dbus.ObjectPath = "/Menu"
)

var dbusTrayID atomic.Uint64
var trayWatchers = []string{"org.kde.StatusNotifierWatcher", "org.freedesktop.StatusNotifierWatcher"}

type trayPixmap struct {
	Width, Height int32
	Pixels        []byte
}
type trayTooltip struct {
	IconName           string
	IconPixmap         []trayPixmap
	Title, Description string
}
type trayLayout struct {
	ID         int32
	Properties map[string]dbus.Variant
	Children   []dbus.Variant
}
type trayItemProperties struct {
	ID         int32
	Properties map[string]dbus.Variant
}
type trayMenuEvent struct {
	ID        int32
	Event     string
	Data      dbus.Variant
	Timestamp uint32
}

type dbusTray struct {
	mu       sync.RWMutex
	conn     *dbus.Conn
	options  TrayOptions
	pixmaps  []trayPixmap
	revision uint32
	ident    string
	closed   bool
	signals  chan *dbus.Signal
}

func trayDBusFailure(err error) error { return fmt.Errorf("%w: tray D-Bus: %v", native.ErrFailed, err) }
func trayMethodError(text string) *dbus.Error {
	return dbus.NewError("org.freedesktop.DBus.Error.InvalidArgs", []any{text})
}

func trayPixmaps(data []byte) ([]trayPixmap, error) {
	art, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, native.ErrInvalidArgument
	}
	b := art.Bounds()
	pixels := make([]byte, b.Dx()*b.Dy()*4)
	i := 0
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			// SNI uses network byte order ARGB32, with straight alpha.
			c := color.NRGBAModel.Convert(art.At(x, y)).(color.NRGBA)
			pixels[i], pixels[i+1], pixels[i+2], pixels[i+3] = c.A, c.R, c.G, c.B
			i += 4
		}
	}
	return []trayPixmap{{int32(b.Dx()), int32(b.Dy()), pixels}}, nil
}

func (t *dbusTray) register() error {
	for _, watcher := range trayWatchers {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		obj := t.conn.Object(watcher, "/StatusNotifierWatcher")
		// A watcher may exist without any host displaying icons.
		var host bool
		err := obj.CallWithContext(ctx, "org.freedesktop.DBus.Properties.Get", 0, watcher, "IsStatusNotifierHostRegistered").Store(&host)
		if err == nil && host {
			err = obj.CallWithContext(ctx, watcher+".RegisterStatusNotifierItem", 0, string(trayItemPath)).Err
		} else if err == nil {
			err = native.ErrUnsupported
		}
		cancel()
		if err == nil {
			return nil
		}
	}
	return fmt.Errorf("%w: no active StatusNotifierHost", native.ErrUnsupported)
}

func (t *dbusTray) watch() {
	for signal := range t.signals {
		hostRegistered := false
		for _, watcher := range trayWatchers {
			if signal.Name == watcher+".StatusNotifierHostRegistered" {
				hostRegistered = true
				break
			}
		}
		if hostRegistered {
			_ = t.register()
			continue
		}
		if signal.Name != "org.freedesktop.DBus.NameOwnerChanged" || len(signal.Body) != 3 {
			continue
		}
		name, _ := signal.Body[0].(string)
		owner, _ := signal.Body[2].(string)
		if owner == "" {
			continue
		}
		for _, watcher := range trayWatchers {
			if name == watcher {
				_ = t.register()
				break
			}
		}
	}
}

func (t *dbusTray) Update(options TrayOptions) error {
	pixmaps, err := trayPixmaps(options.PNG)
	if err != nil {
		return err
	}
	t.mu.Lock()
	if t.closed {
		t.mu.Unlock()
		return native.ErrFailed
	}
	oldOptions, oldPixmaps, oldRevision := t.options, t.pixmaps, t.revision
	t.options, t.pixmaps = options, pixmaps
	t.revision++
	revision := t.revision
	t.mu.Unlock()
	// Properties are read as snapshots; signals explicitly wake the desktop host.
	emit := func() error {
		for _, iface := range []string{traySNI, "org.freedesktop.StatusNotifierItem"} {
			for _, signal := range []string{"NewIcon", "NewToolTip", "NewTitle"} {
				if err := t.conn.Emit(trayItemPath, iface+"."+signal); err != nil {
					return err
				}
			}
		}
		return t.conn.Emit(trayMenuPath, trayMenuInterface+".LayoutUpdated", revision, int32(0))
	}
	if err := emit(); err != nil {
		t.mu.Lock()
		t.options, t.pixmaps, t.revision = oldOptions, oldPixmaps, oldRevision
		t.mu.Unlock()
		return trayDBusFailure(err)
	}
	return nil
}

func (t *dbusTray) Close() error {
	t.mu.Lock()
	if t.closed {
		t.mu.Unlock()
		return nil
	}
	t.closed = true
	t.mu.Unlock()
	// Closing the private connection unregisters the item and all exports/matches.
	return t.conn.Close()
}

func (t *dbusTray) activate() {
	t.mu.RLock()
	event, closed := t.options.Event, t.closed
	t.mu.RUnlock()
	if !closed {
		event(0)
	}
}

func trayFind(items []TrayMenuItem, id uint32) (TrayMenuItem, bool) {
	for _, item := range items {
		if item.ID == id {
			return item, true
		}
		if child, ok := trayFind(item.Children, id); ok {
			return child, true
		}
	}
	return TrayMenuItem{}, false
}

func (t *dbusTray) menuEvent(id int32, event string) bool {
	t.mu.RLock()
	item, found := trayFind(t.options.Menu, uint32(id))
	emit, closed := t.options.Event, t.closed
	t.mu.RUnlock()
	if found && !closed && event == "clicked" && !item.Disabled && !item.Separator && len(item.Children) == 0 {
		emit(item.ID)
	}
	return found
}

func trayMenuProps(item TrayMenuItem) map[string]dbus.Variant {
	values := map[string]dbus.Variant{"label": dbus.MakeVariant(item.Title), "enabled": dbus.MakeVariant(!item.Disabled), "visible": dbus.MakeVariant(true)}
	if item.Separator {
		values["type"] = dbus.MakeVariant("separator")
	}
	if len(item.PNG) > 0 {
		values["icon-data"] = dbus.MakeVariant(append([]byte(nil), item.PNG...))
	}
	if item.Checked {
		values["toggle-type"] = dbus.MakeVariant("checkmark")
		values["toggle-state"] = dbus.MakeVariant(int32(1))
	}
	if len(item.Children) != 0 {
		values["children-display"] = dbus.MakeVariant("submenu")
	}
	return values
}

func trayFilterProps(values map[string]dbus.Variant, names []string) map[string]dbus.Variant {
	if len(names) == 0 {
		return values
	}
	result := make(map[string]dbus.Variant)
	for _, name := range names {
		if value, ok := values[name]; ok {
			result[name] = value
		}
	}
	return result
}

func trayMakeLayout(id int32, items []TrayMenuItem, depth int32, names []string) (trayLayout, bool) {
	result := trayLayout{ID: id, Properties: trayFilterProps(map[string]dbus.Variant{"children-display": dbus.MakeVariant("submenu")}, names), Children: []dbus.Variant{}}
	if id != 0 {
		item, ok := trayFind(items, uint32(id))
		if !ok {
			return result, false
		}
		result.Properties = trayFilterProps(trayMenuProps(item), names)
		items = item.Children
	}
	if depth != 0 {
		for _, item := range items {
			child, _ := trayMakeLayout(int32(item.ID), items, depth-1, names)
			result.Children = append(result.Children, dbus.MakeVariant(child))
		}
	}
	return result, true
}

func (t *dbusTray) properties(menu bool) map[string]dbus.Variant {
	t.mu.RLock()
	defer t.mu.RUnlock()
	if menu {
		return map[string]dbus.Variant{"Version": dbus.MakeVariant(uint32(3)), "TextDirection": dbus.MakeVariant("ltr"), "Status": dbus.MakeVariant("normal"), "IconThemePath": dbus.MakeVariant([]string{})}
	}
	return map[string]dbus.Variant{
		"Category": dbus.MakeVariant("ApplicationStatus"), "Id": dbus.MakeVariant(t.ident), "Title": dbus.MakeVariant(t.options.Tooltip),
		"Status": dbus.MakeVariant("Active"), "WindowId": dbus.MakeVariant(uint32(0)),
		"IconName": dbus.MakeVariant(""), "IconPixmap": dbus.MakeVariant(t.pixmaps),
		"OverlayIconName": dbus.MakeVariant(""), "OverlayIconPixmap": dbus.MakeVariant([]trayPixmap{}),
		"AttentionIconName": dbus.MakeVariant(""), "AttentionIconPixmap": dbus.MakeVariant([]trayPixmap{}), "AttentionMovieName": dbus.MakeVariant(""),
		"ToolTip":    dbus.MakeVariant(trayTooltip{"", []trayPixmap{}, t.options.Tooltip, ""}),
		"ItemIsMenu": dbus.MakeVariant(len(t.options.Menu) != 0), "Menu": dbus.MakeVariant(trayMenuPath),
	}
}

func (t *dbusTray) export() error {
	methods := map[string]any{
		"Activate":          func(x, y int32) *dbus.Error { t.activate(); return nil },
		"SecondaryActivate": func(x, y int32) *dbus.Error { t.activate(); return nil },
		"ContextMenu":       func(x, y int32) *dbus.Error { return nil }, // Host renders the exported DBusMenu.
		"Scroll":            func(delta int32, orientation string) *dbus.Error { return nil },
	}
	// Export both namespaces: existing hosts commonly use org.kde; the published
	// specification uses org.freedesktop.
	for _, iface := range []string{traySNI, "org.freedesktop.StatusNotifierItem"} {
		if err := t.conn.ExportMethodTable(methods, trayItemPath, iface); err != nil {
			return err
		}
	}
	menuMethods := map[string]any{
		"GetLayout": func(parent, depth int32, names []string) (uint32, trayLayout, *dbus.Error) {
			t.mu.RLock()
			defer t.mu.RUnlock()
			layout, ok := trayMakeLayout(parent, t.options.Menu, depth, names)
			if !ok {
				return 0, layout, trayMethodError("unknown menu item")
			}
			return t.revision, layout, nil
		},
		"GetGroupProperties": func(ids []int32, names []string) ([]trayItemProperties, *dbus.Error) {
			t.mu.RLock()
			defer t.mu.RUnlock()
			result := []trayItemProperties{}
			var walk func([]TrayMenuItem)
			walk = func(items []TrayMenuItem) {
				for _, item := range items {
					result = append(result, trayItemProperties{int32(item.ID), trayFilterProps(trayMenuProps(item), names)})
					walk(item.Children)
				}
			}
			if len(ids) == 0 {
				walk(t.options.Menu)
			} else {
				for _, id := range ids {
					if item, ok := trayFind(t.options.Menu, uint32(id)); ok {
						result = append(result, trayItemProperties{id, trayFilterProps(trayMenuProps(item), names)})
					}
				}
			}
			return result, nil
		},
		"GetProperty": func(id int32, name string) (dbus.Variant, *dbus.Error) {
			t.mu.RLock()
			defer t.mu.RUnlock()
			item, ok := trayFind(t.options.Menu, uint32(id))
			if !ok {
				return dbus.Variant{}, trayMethodError("unknown menu item")
			}
			value, ok := trayMenuProps(item)[name]
			if !ok {
				return dbus.Variant{}, trayMethodError("unknown menu property")
			}
			return value, nil
		},
		"Event": func(id int32, event string, data dbus.Variant, timestamp uint32) *dbus.Error {
			if !t.menuEvent(id, event) {
				return trayMethodError("unknown menu item")
			}
			return nil
		},
		"EventGroup": func(events []trayMenuEvent) ([]int32, *dbus.Error) {
			failed := []int32{}
			for _, event := range events {
				if !t.menuEvent(event.ID, event.Event) {
					failed = append(failed, event.ID)
				}
			}
			return failed, nil
		},
		"AboutToShow":      func(id int32) (bool, *dbus.Error) { return false, nil },
		"AboutToShowGroup": func(ids []int32) ([]int32, []int32, *dbus.Error) { return []int32{}, []int32{}, nil },
	}
	if err := t.conn.ExportMethodTable(menuMethods, trayMenuPath, trayMenuInterface); err != nil {
		return err
	}
	for _, menu := range []bool{false, true} {
		path := trayItemPath
		iface := traySNI
		if menu {
			path = trayMenuPath
			iface = trayMenuInterface
		}
		props := func(request string) (map[string]dbus.Variant, *dbus.Error) {
			if request != iface && !(!menu && request == "org.freedesktop.StatusNotifierItem") {
				return nil, dbus.NewError("org.freedesktop.DBus.Error.UnknownInterface", []any{request})
			}
			return t.properties(menu), nil
		}
		table := map[string]any{
			"GetAll": props,
			"Get": func(request, property string) (dbus.Variant, *dbus.Error) {
				all, err := props(request)
				if err != nil {
					return dbus.Variant{}, err
				}
				value, ok := all[property]
				if !ok {
					return dbus.Variant{}, dbus.NewError("org.freedesktop.DBus.Error.UnknownProperty", []any{property})
				}
				return value, nil
			},
			"Set": func(request, property string, value dbus.Variant) *dbus.Error {
				return dbus.NewError("org.freedesktop.DBus.Error.PropertyReadOnly", []any{property})
			},
		}
		if err := t.conn.ExportMethodTable(table, path, "org.freedesktop.DBus.Properties"); err != nil {
			return err
		}
		node := &introspect.Node{Name: string(path), Interfaces: []introspect.Interface{introspect.IntrospectData, prop.IntrospectData, {Name: iface}}}
		for name, value := range t.properties(menu) {
			node.Interfaces[2].Properties = append(node.Interfaces[2].Properties, introspect.Property{Name: name, Type: value.Signature().String(), Access: "read"})
		}
		// Introspection exposes method signatures to hosts that probe capabilities.
		methodTable := methods
		if menu {
			methodTable = menuMethods
		}
		for name, method := range methodTable {
			node.Interfaces[2].Methods = append(node.Interfaces[2].Methods, trayIntrospectMethod(name, method))
		}
		if !menu {
			alias := node.Interfaces[2]
			alias.Name = "org.freedesktop.StatusNotifierItem"
			node.Interfaces = append(node.Interfaces, alias)
		}
		if err := t.conn.Export(introspect.NewIntrospectable(node), path, "org.freedesktop.DBus.Introspectable"); err != nil {
			return err
		}
	}
	return nil
}

func trayIntrospectMethod(name string, method any) introspect.Method {
	typ := reflect.TypeOf(method)
	result := introspect.Method{Name: name}
	for i := 0; i < typ.NumIn(); i++ {
		result.Args = append(result.Args, introspect.Arg{Type: dbus.SignatureOfType(typ.In(i)).String(), Direction: "in"})
	}
	for i := 0; i < typ.NumOut()-1; i++ {
		result.Args = append(result.Args, introspect.Arg{Type: dbus.SignatureOfType(typ.Out(i)).String(), Direction: "out"})
	}
	return result
}
