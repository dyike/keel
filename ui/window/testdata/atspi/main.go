//go:build linux && !android

// atspi checks the AT-SPI bridge in a real window: as a client on the
// accessibility bus, it finds this application on the registry's desktop,
// reads the window's elements, presses a button, and scrolls a row out of
// view into view and presses it. Run it in a session with at-spi2-core
// (dbus-run-session) and a display.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/window"
	"github.com/godbus/dbus/v5"
)

type ref struct {
	Name string
	Path dbus.ObjectPath
}

const (
	access = "org.a11y.atspi.Accessible"
	props  = "org.freedesktop.DBus.Properties"
)

func fail(format string, args ...any) {
	fmt.Printf("FAIL: "+format+"\n", args...)
	os.Exit(1)
}

func main() {
	saved, picked := 0, -1 // frame lock
	content := el.Root(el.ViewFunc(func(cx *el.Context) el.Element {
		label := "Save"
		if saved > 0 {
			label = fmt.Sprintf("Saved %d", saved)
		}
		list := el.Div().H(el.Dp(100)).ScrollY()
		for i := range 30 {
			list.Child(el.Div().P(4).OnClick(func() { picked = i }).Child(el.Text(fmt.Sprintf("Item %d", i))))
		}
		return el.Div().P(24).Child(
			el.Text("Hello"),
			el.Div().P(8).OnClick(func() { saved++ }).Child(el.Text(label)),
			list,
		)
	}))
	w := window.Open(window.Options{Title: "AT-SPI check", Width: 360, Height: 320, Content: content})
	go func() {
		defer func() { core.Update(func() { w.Close() }) }()
		client := connect()
		c := &checker{conn: client}
		app := c.findApp(filepath.Base(os.Args[0]))
		win := c.poll("window", func() (ref, bool) {
			kids := c.children(app)
			return first(kids), len(kids) > 0
		})
		if role := c.role(win); role != 23 {
			fail("window role %d, want frame (23)", role)
		}
		if name := c.name(win); name != "AT-SPI check" {
			fail("window name %q", name)
		}
		// Elements appear after the first request redraws the window.
		save := c.poll("Save button", func() (ref, bool) { return c.find(win, 43, "Save") })
		if _, ok := c.find(win, 29, "Hello"); !ok {
			fail("no Hello label:\n%s", c.dump(win, ""))
		}
		pane, ok := c.findRole(win, 49)
		if !ok {
			fail("no scroll pane:\n%s", c.dump(win, ""))
		}
		item, ok := c.find(pane, 43, "Item 25")
		if !ok {
			fail("Item 25 is not in the scroll pane:\n%s", c.dump(win, ""))
		}
		if c.showing(item) {
			fail("Item 25 shows before scrolling")
		}
		c.call(save, "org.a11y.atspi.Action.DoAction", nil, int32(0))
		c.poll("Saved 1", func() (ref, bool) { return c.find(win, 43, "Saved 1") })
		c.call(item, "org.a11y.atspi.Component.ScrollTo", nil, uint32(0))
		c.poll("Item 25 showing", func() (ref, bool) { return item, c.showing(item) })
		c.call(item, "org.a11y.atspi.Action.DoAction", nil, int32(0))
		deadline := time.Now().Add(10 * time.Second)
		for {
			got := make(chan int)
			core.Update(func() { got <- picked })
			if <-got == 25 {
				break
			}
			if time.Now().After(deadline) {
				fail("pressing Item 25 did not pick it")
			}
			time.Sleep(50 * time.Millisecond)
		}
		fmt.Println("PASS: AT-SPI registration, roles, names, scroll pane, press, scroll into view")
	}()
	window.Main()
}

func connect() *dbus.Conn {
	session, err := dbus.SessionBus()
	if err != nil {
		fail("session bus: %v", err)
	}
	var addr string
	if err := session.Object("org.a11y.Bus", "/org/a11y/bus").Call("org.a11y.Bus.GetAddress", 0).Store(&addr); err != nil {
		fail("accessibility bus: %v", err)
	}
	conn, err := dbus.Dial(addr)
	if err == nil {
		err = conn.Auth(nil)
	}
	if err == nil {
		err = conn.Hello()
	}
	if err != nil {
		fail("dial %s: %v", addr, err)
	}
	return conn
}

type checker struct{ conn *dbus.Conn }

func (c *checker) call(r ref, method string, out any, args ...any) error {
	call := c.conn.Object(r.Name, r.Path).Call(method, 0, args...)
	if call.Err != nil {
		return call.Err
	}
	if out != nil {
		return call.Store(out)
	}
	return nil
}

func (c *checker) children(r ref) []ref {
	var kids []ref
	if err := c.call(r, access+".GetChildren", &kids); err != nil {
		return nil
	}
	return kids
}

func (c *checker) role(r ref) uint32 {
	var role uint32
	c.call(r, access+".GetRole", &role)
	return role
}

func (c *checker) name(r ref) string {
	var v dbus.Variant
	if c.call(r, props+".Get", &v, access, "Name") != nil {
		return ""
	}
	s, _ := v.Value().(string)
	return s
}

func (c *checker) showing(r ref) bool {
	var s []uint32
	c.call(r, access+".GetState", &s)
	return len(s) == 2 && s[0]&(1<<25) != 0
}

func (c *checker) findApp(name string) ref {
	desktop := ref{"org.a11y.atspi.Registry", "/org/a11y/atspi/accessible/root"}
	return c.poll("application "+name+" on the desktop", func() (ref, bool) {
		for _, app := range c.children(desktop) {
			if c.name(app) == name {
				return app, true
			}
		}
		return ref{}, false
	})
}

func (c *checker) find(under ref, role uint32, name string) (ref, bool) {
	for _, k := range c.children(under) {
		if c.role(k) == role && c.name(k) == name {
			return k, true
		}
		if r, ok := c.find(k, role, name); ok {
			return r, true
		}
	}
	return ref{}, false
}

func (c *checker) findRole(under ref, role uint32) (ref, bool) {
	for _, k := range c.children(under) {
		if c.role(k) == role {
			return k, true
		}
		if r, ok := c.findRole(k, role); ok {
			return r, true
		}
	}
	return ref{}, false
}

func (c *checker) dump(r ref, indent string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s%d %q\n", indent, c.role(r), c.name(r))
	for _, k := range c.children(r) {
		b.WriteString(c.dump(k, indent+"  "))
	}
	return b.String()
}

func (c *checker) poll(what string, fn func() (ref, bool)) ref {
	deadline := time.Now().Add(15 * time.Second)
	for time.Now().Before(deadline) {
		if r, ok := fn(); ok {
			return r
		}
		time.Sleep(100 * time.Millisecond)
	}
	fail("timed out waiting for %s", what)
	return ref{}
}

func first(refs []ref) ref {
	if len(refs) == 0 {
		return ref{}
	}
	return refs[0]
}
