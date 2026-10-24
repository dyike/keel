//go:build !js

package atspi

import (
	"image"

	"github.com/godbus/dbus/v5"
)

// D-Bus interfaces, exported over the whole /org/a11y/atspi/accessible
// subtree: each method finds its object from the message's path.

var errUnknown = dbus.NewError("org.freedesktop.DBus.Error.UnknownObject", []any{"no such accessible object"})

var errFailed = dbus.NewError("org.freedesktop.DBus.Error.Failed", []any{"internal error"})

// with runs fn on the object a call names, holding b.mu. A fault in the
// bridge fails the call instead of the application.
func (b *Bridge) with(msg dbus.Message, fn func(o object)) (err *dbus.Error) {
	path, _ := msg.Headers[dbus.FieldPath].Value().(dbus.ObjectPath)
	b.mu.Lock()
	defer b.mu.Unlock()
	defer func() {
		if r := recover(); r != nil {
			err = errFailed
		}
	}()
	b.asked()
	o, ok := b.lookup(path)
	if !ok {
		return errUnknown
	}
	fn(o)
	return nil
}

type accessibleIface struct{ b *Bridge }

func (a accessibleIface) GetChildAtIndex(msg dbus.Message, i int32) (r ref, err *dbus.Error) {
	r = ref{"", "/org/a11y/atspi/null"}
	err = a.b.with(msg, func(o object) {
		if kids := a.b.children(o); i >= 0 && int(i) < len(kids) {
			r = a.b.refOf(kids[i])
		}
	})
	return r, err
}

func (a accessibleIface) GetChildren(msg dbus.Message) (refs []ref, err *dbus.Error) {
	refs = []ref{}
	err = a.b.with(msg, func(o object) {
		for _, c := range a.b.children(o) {
			refs = append(refs, a.b.refOf(c))
		}
	})
	return refs, err
}

func (a accessibleIface) GetIndexInParent(msg dbus.Message) (i int32, err *dbus.Error) {
	i = -1
	err = a.b.with(msg, func(o object) {
		p, ok := a.b.parentObject(o)
		if !ok {
			return
		}
		for j, c := range a.b.children(p) {
			if a.b.refOf(c) == a.b.refOf(o) {
				i = int32(j)
			}
		}
	})
	return i, err
}

type relation struct {
	Type    uint32
	Targets []ref
}

func (a accessibleIface) GetRelationSet(msg dbus.Message) ([]relation, *dbus.Error) {
	return []relation{}, a.b.with(msg, func(object) {})
}

func (a accessibleIface) GetRole(msg dbus.Message) (r uint32, err *dbus.Error) {
	err = a.b.with(msg, func(o object) { r = a.b.role(o) })
	return r, err
}

func (a accessibleIface) GetRoleName(msg dbus.Message) (s string, err *dbus.Error) {
	err = a.b.with(msg, func(o object) { s = roleNames[a.b.role(o)] })
	return s, err
}

func (a accessibleIface) GetLocalizedRoleName(msg dbus.Message) (string, *dbus.Error) {
	return a.GetRoleName(msg)
}

func (a accessibleIface) GetState(msg dbus.Message) (s []uint32, err *dbus.Error) {
	err = a.b.with(msg, func(o object) {
		st := a.b.states(o)
		s = []uint32{st[0], st[1]}
	})
	return s, err
}

func (a accessibleIface) GetAttributes(msg dbus.Message) (map[string]string, *dbus.Error) {
	attrs := map[string]string{}
	err := a.b.with(msg, func(o object) {
		if o.app {
			attrs["toolkit"] = "Keel"
		}
	})
	return attrs, err
}

func (a accessibleIface) GetApplication(msg dbus.Message) (r ref, err *dbus.Error) {
	err = a.b.with(msg, func(object) { r = ref{a.b.name, rootPath} })
	return r, err
}

func (a accessibleIface) GetInterfaces(msg dbus.Message) (s []string, err *dbus.Error) {
	err = a.b.with(msg, func(o object) { s = a.b.interfaces(o) })
	return s, err
}

type applicationIface struct{ b *Bridge }

func (a applicationIface) GetLocale(msg dbus.Message, lctype uint32) (string, *dbus.Error) {
	return "", a.b.with(msg, func(object) {})
}

func (a applicationIface) GetApplicationBusAddress(msg dbus.Message) (string, *dbus.Error) {
	return "", a.b.with(msg, func(object) {})
}

// Coordinate types.
const (
	coordScreen = 0
	coordWindow = 1
	coordParent = 2
)

type extents struct{ X, Y, Width, Height int32 }

type componentIface struct{ b *Bridge }

// rect returns an object's extents in the given coordinates. Window
// positions on the screen are not known (Wayland does not tell), so screen
// coordinates are window coordinates.
func (b *Bridge) rect(o object, coord uint32) image.Rectangle {
	r := b.extents(o)
	if coord == coordParent && !o.app && o.node >= 0 {
		if p, ok := b.parentObject(o); ok {
			r = r.Sub(b.extents(p).Min)
		}
	}
	return r
}

func (c componentIface) Contains(msg dbus.Message, x, y int32, coord uint32) (in bool, err *dbus.Error) {
	err = c.b.with(msg, func(o object) { in = image.Pt(int(x), int(y)).In(c.b.rect(o, coord)) })
	return in, err
}

func (c componentIface) GetAccessibleAtPoint(msg dbus.Message, x, y int32, coord uint32) (r ref, err *dbus.Error) {
	r = ref{"", "/org/a11y/atspi/null"}
	err = c.b.with(msg, func(o object) {
		if o.app {
			return
		}
		p := image.Pt(int(x), int(y))
		if coord == coordParent {
			p = p.Add(c.b.extents(o).Min)
		}
		// The innermost element showing at p: children follow their parent.
		for i := len(o.w.nodes) - 1; i >= 0; i-- {
			n := o.w.nodes[i]
			if !p.In(n.Bounds.Intersect(n.Viewport)) || !c.b.inside(o, i) {
				continue
			}
			r = c.b.refOf(object{w: o.w, node: i})
			return
		}
	})
	return r, err
}

// inside reports whether node i is below o.
func (b *Bridge) inside(o object, i int) bool {
	if o.node < 0 {
		return true
	}
	for p := o.w.nodes[i].Parent; p >= 0; p = o.w.nodes[p].Parent {
		if p == o.node {
			return true
		}
	}
	return false
}

func (c componentIface) GetExtents(msg dbus.Message, coord uint32) (e extents, err *dbus.Error) {
	err = c.b.with(msg, func(o object) {
		r := c.b.rect(o, coord)
		e = extents{int32(r.Min.X), int32(r.Min.Y), int32(r.Dx()), int32(r.Dy())}
	})
	return e, err
}

func (c componentIface) GetPosition(msg dbus.Message, coord uint32) (x, y int32, err *dbus.Error) {
	err = c.b.with(msg, func(o object) {
		r := c.b.rect(o, coord)
		x, y = int32(r.Min.X), int32(r.Min.Y)
	})
	return x, y, err
}

func (c componentIface) GetSize(msg dbus.Message) (w, h int32, err *dbus.Error) {
	err = c.b.with(msg, func(o object) {
		r := c.b.extents(o)
		w, h = int32(r.Dx()), int32(r.Dy())
	})
	return w, h, err
}

// Layers.
const (
	layerWidget = 3
	layerWindow = 7
)

func (c componentIface) GetLayer(msg dbus.Message) (l uint32, err *dbus.Error) {
	err = c.b.with(msg, func(o object) {
		l = layerWidget
		if o.node < 0 {
			l = layerWindow
		}
	})
	return l, err
}

func (c componentIface) GetMDIZOrder(msg dbus.Message) (int16, *dbus.Error) {
	return 0, c.b.with(msg, func(object) {})
}

func (c componentIface) GrabFocus(msg dbus.Message) (bool, *dbus.Error) {
	return false, c.b.with(msg, func(object) {})
}

func (c componentIface) GetAlpha(msg dbus.Message) (float64, *dbus.Error) {
	return 1, c.b.with(msg, func(object) {})
}

// act runs a window action outside b.mu.
func (b *Bridge) act(msg dbus.Message, do func(a Actions, key string) bool) (bool, *dbus.Error) {
	var id int
	var key string
	err := b.with(msg, func(o object) {
		if !o.app && o.node >= 0 {
			id, key = o.w.id, o.w.nodes[o.node].Key
		}
	})
	if err != nil || key == "" || b.actions == nil {
		return false, err
	}
	a := b.actions(id)
	if a == nil {
		return false, nil
	}
	return do(a, key), nil
}

func (c componentIface) ScrollTo(msg dbus.Message, how uint32) (bool, *dbus.Error) {
	return c.b.act(msg, func(a Actions, key string) bool { return a.ScrollTo(key) })
}

func (c componentIface) ScrollToPoint(msg dbus.Message, coord uint32, x, y int32) (bool, *dbus.Error) {
	return c.ScrollTo(msg, 0)
}

type actionIface struct{ b *Bridge }

type actionInfo struct{ Name, Description, KeyBinding string }

func (a actionIface) GetDescription(msg dbus.Message, i int32) (string, *dbus.Error) {
	return "", a.b.with(msg, func(object) {})
}

func (a actionIface) GetName(msg dbus.Message, i int32) (string, *dbus.Error) {
	return "click", a.b.with(msg, func(object) {})
}

func (a actionIface) GetLocalizedName(msg dbus.Message, i int32) (string, *dbus.Error) {
	return a.GetName(msg, i)
}

func (a actionIface) GetKeyBinding(msg dbus.Message, i int32) (string, *dbus.Error) {
	return "", a.b.with(msg, func(object) {})
}

func (a actionIface) GetActions(msg dbus.Message) ([]actionInfo, *dbus.Error) {
	return []actionInfo{{Name: "click"}}, a.b.with(msg, func(object) {})
}

func (a actionIface) DoAction(msg dbus.Message, i int32) (bool, *dbus.Error) {
	if i != 0 {
		return false, nil
	}
	return a.b.act(msg, func(a Actions, key string) bool { return a.Press(key) })
}

type propsIface struct{ b *Bridge }

func (p propsIface) GetAll(msg dbus.Message, iface string) (props map[string]dbus.Variant, err *dbus.Error) {
	props = map[string]dbus.Variant{}
	err = p.b.with(msg, func(o object) {
		b := p.b
		switch iface {
		case ifaceAccess:
			props["Name"] = dbus.MakeVariant(b.label(o))
			props["Description"] = dbus.MakeVariant(b.description(o))
			props["Parent"] = dbus.MakeVariant(b.parentRef(o))
			props["ChildCount"] = dbus.MakeVariant(int32(len(b.children(o))))
			props["Locale"] = dbus.MakeVariant("")
			props["AccessibleId"] = dbus.MakeVariant(string(b.refOf(o).Path))
			props["HelpText"] = dbus.MakeVariant("")
		case ifaceApp:
			props["ToolkitName"] = dbus.MakeVariant("Keel")
			props["Version"] = dbus.MakeVariant("1")
			props["AtspiVersion"] = dbus.MakeVariant("2.1")
			props["Id"] = dbus.MakeVariant(int32(0))
		case ifaceAction:
			n := int32(0)
			if !o.app && o.node >= 0 && pressable(o.w.nodes[o.node].Role) {
				n = 1
			}
			props["NActions"] = dbus.MakeVariant(n)
		}
	})
	return props, err
}

func (p propsIface) Get(msg dbus.Message, iface, name string) (dbus.Variant, *dbus.Error) {
	props, err := p.GetAll(msg, iface)
	if err != nil {
		return dbus.Variant{}, err
	}
	v, ok := props[name]
	if !ok {
		return dbus.Variant{}, dbus.NewError("org.freedesktop.DBus.Error.UnknownProperty", []any{name})
	}
	return v, nil
}

// Set accepts the registry's Application.Id and ignores the rest.
func (p propsIface) Set(msg dbus.Message, iface, name string, value dbus.Variant) *dbus.Error {
	return p.b.with(msg, func(object) {})
}
