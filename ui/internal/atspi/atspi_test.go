//go:build !js

package atspi

import (
	"bufio"
	"fmt"
	"image"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/godbus/dbus/v5"
)

// startBus runs a private dbus-daemon; tests skip without one.
func startBus(t *testing.T) string {
	t.Helper()
	daemon, err := exec.LookPath("dbus-daemon")
	if err != nil {
		t.Skip("dbus-daemon not found")
	}
	dir := t.TempDir()
	conf := filepath.Join(dir, "bus.conf")
	err = os.WriteFile(conf, []byte(`<!DOCTYPE busconfig PUBLIC "-//freedesktop//DTD D-Bus Bus Configuration 1.0//EN"
 "http://www.freedesktop.org/standards/dbus/1.0/busconfig.dtd">
<busconfig>
  <type>session</type>
  <listen>unix:dir=`+dir+`</listen>
  <auth>ANONYMOUS</auth>
  <allow_anonymous/>
  <policy context="default">
    <allow send_destination="*" eavesdrop="true"/>
    <allow eavesdrop="true"/>
    <allow own="*"/>
  </policy>
</busconfig>`), 0o600)
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(daemon, "--config-file="+conf, "--print-address", "--nofork")
	out, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { cmd.Process.Kill(); cmd.Wait() })
	addr, err := bufio.NewReader(out).ReadString('\n')
	if err != nil {
		t.Fatal(err)
	}
	return strings.TrimSpace(addr)
}

func dial(t *testing.T, addr string) *dbus.Conn {
	t.Helper()
	conn, err := dbus.Dial(addr)
	if err != nil {
		t.Fatal(err)
	}
	if err := conn.Auth([]dbus.Auth{dbus.AuthAnonymous()}); err != nil {
		t.Fatal(err)
	}
	if err := conn.Hello(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { conn.Close() })
	return conn
}

// registry stands in for at-spi2-registryd: it accepts the application.
type registry struct {
	desktop ref
	mu      sync.Mutex
	plugs   []ref
}

func (r *registry) Embed(plug ref) (ref, *dbus.Error) {
	r.mu.Lock()
	r.plugs = append(r.plugs, plug)
	r.mu.Unlock()
	return r.desktop, nil
}

type actions struct {
	mu                  sync.Mutex
	pressed, scrolledTo []string
}

func (a *actions) Press(key string) bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.pressed = append(a.pressed, key)
	return true
}

func (a *actions) ScrollTo(key string) bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.scrolledTo = append(a.scrolledTo, key)
	return true
}

func (a *actions) ScrollPage(string, int, int) bool { return true }

func ptr(b bool) *bool { return &b }

// A client finds the application on the desktop, walks a window's
// elements, reads their roles, names, states and extents, presses one and
// scrolls another into view, and hears about changes.
func TestBridge(t *testing.T) {
	addr := startBus(t)
	client := dial(t, addr)
	reg := &registry{desktop: ref{client.Names()[0], rootPath}}
	if err := client.Export(reg, rootPath, "org.a11y.atspi.Socket"); err != nil {
		t.Fatal(err)
	}
	if reply, err := client.RequestName("org.a11y.atspi.Registry", 0); err != nil || reply != dbus.RequestNameReplyPrimaryOwner {
		t.Fatal(reply, err)
	}
	app := dial(t, addr)
	acts := &actions{}
	activated := make(chan struct{}, 1)
	b, err := New(app, "demo", func(id int) Actions {
		if id == 1 {
			return acts
		}
		return nil
	}, func() { activated <- struct{}{} })
	if err != nil {
		t.Fatal(err)
	}
	if err := b.Register(); err != nil {
		t.Fatal(err)
	}
	if len(reg.plugs) != 1 || reg.plugs[0] != (ref{app.Names()[0], rootPath}) {
		t.Fatalf("registry got %v", reg.plugs)
	}
	if err := client.AddMatchSignal(dbus.WithMatchInterface(ifaceObject)); err != nil {
		t.Fatal(err)
	}
	signals := make(chan *dbus.Signal, 64)
	client.Signal(signals)

	view := image.Rect(0, 0, 400, 300)
	area := image.Rect(0, 40, 400, 300)
	nodes := []Node{
		{Key: "1", Role: "text", Name: "Orders", Bounds: image.Rect(10, 10, 200, 30), Viewport: view, Parent: -1},
		{Key: "2", Role: "checkbox", Name: "Paid", Checked: ptr(false), Bounds: image.Rect(10, 40, 200, 60), Viewport: view, Parent: -1},
		{Key: "3", Role: "scrollarea", Bounds: area, Viewport: view, Parent: -1},
		{Key: "4", Role: "button", Name: "Order 1", Bounds: image.Rect(0, 60, 400, 90), Viewport: area, Parent: 2},
		{Key: "5", Role: "button", Name: "Order 99", Bounds: image.Rect(0, 3000, 400, 3030), Viewport: area, Parent: 2, Disabled: true},
	}
	b.AddWindow(1, "Orders")
	b.Update(1, "Orders", nodes)

	call := func(path dbus.ObjectPath, method string, out any, args ...any) {
		t.Helper()
		if err := client.Object(app.Names()[0], path).Call(method, 0, args...).Store(out); err != nil {
			t.Fatalf("%s %s: %v", path, method, err)
		}
	}
	var apps []ref
	call(rootPath, ifaceAccess+".GetChildren", &apps)
	if len(apps) != 1 {
		t.Fatalf("application children %v", apps)
	}
	select {
	case <-activated:
	case <-time.After(time.Second):
		t.Fatal("first request did not activate the bridge")
	}
	win := apps[0].Path
	var role uint32
	call(win, ifaceAccess+".GetRole", &role)
	var title dbus.Variant
	call(win, ifaceProps+".Get", &title, ifaceAccess, "Name")
	if role != roleFrame || title.Value() != "Orders" {
		t.Fatalf("window role %d name %v", role, title)
	}
	var top []ref
	call(win, ifaceAccess+".GetChildren", &top)
	if len(top) != 3 {
		t.Fatalf("window children %v", top)
	}
	var rows []ref
	call(top[2].Path, ifaceAccess+".GetChildren", &rows)
	if len(rows) != 2 {
		t.Fatalf("scroll pane children %v", rows)
	}
	call(top[2].Path, ifaceAccess+".GetRole", &role)
	if role != roleScrollPane {
		t.Fatalf("scroll area role %d", role)
	}
	var parent dbus.Variant
	call(rows[1].Path, ifaceProps+".Get", &parent, ifaceAccess, "Parent")
	if p, ok := parent.Value().([]any); !ok || p[1] != top[2].Path {
		t.Fatalf("row parent %v", parent)
	}
	has := func(s []uint32, state uint) bool { return s[state/32]&(1<<(state%32)) != 0 }
	var states []uint32
	call(rows[1].Path, ifaceAccess+".GetState", &states)
	if has(states, stateShowing) || has(states, stateEnabled) || !has(states, stateVisible) {
		t.Fatalf("hidden disabled row states %x", states)
	}
	call(rows[0].Path, ifaceAccess+".GetState", &states)
	if !has(states, stateShowing) || !has(states, stateEnabled) || !has(states, stateFocusable) {
		t.Fatalf("row states %x", states)
	}
	call(top[1].Path, ifaceAccess+".GetState", &states)
	if !has(states, stateCheckable) || has(states, stateChecked) {
		t.Fatalf("checkbox states %x", states)
	}
	var ext extents
	call(rows[1].Path, ifaceComp+".GetExtents", &ext, uint32(coordWindow))
	if ext != (extents{0, 3000, 400, 30}) {
		t.Fatalf("row extents %v", ext)
	}
	var at ref
	call(win, ifaceComp+".GetAccessibleAtPoint", &at, int32(20), int32(70), uint32(coordWindow))
	if at != rows[0] {
		t.Fatalf("at (20,70): %v, want %v", at, rows[0])
	}
	var ifaces []string
	call(top[0].Path, ifaceAccess+".GetInterfaces", &ifaces)
	if strings.Contains(strings.Join(ifaces, " "), ifaceAction) {
		t.Fatalf("text has an action: %v", ifaces)
	}
	var ok bool
	call(rows[0].Path, ifaceAction+".DoAction", &ok, int32(0))
	call(rows[1].Path, ifaceComp+".ScrollTo", &ok, uint32(0))
	acts.mu.Lock()
	if strings.Join(acts.pressed, ",") != "4" || strings.Join(acts.scrolledTo, ",") != "5" {
		t.Fatalf("pressed %v, scrolled to %v", acts.pressed, acts.scrolledTo)
	}
	acts.mu.Unlock()

	// Check the box, drop a row: clients hear both.
	for len(signals) > 0 {
		<-signals
	}
	nodes[1].Checked = ptr(true)
	b.Update(1, "Orders", nodes[:4])
	want := map[string]bool{"StateChanged checked 1": false, "ChildrenChanged remove 1": false}
	deadline := time.After(2 * time.Second)
	for {
		done := true
		for _, seen := range want {
			done = done && seen
		}
		if done {
			break
		}
		select {
		case s := <-signals:
			member := s.Name[strings.LastIndex(s.Name, ".")+1:]
			key := fmt.Sprintf("%s %s %d", member, s.Body[0], s.Body[1])
			if _, ok := want[key]; ok {
				want[key] = true
			}
		case <-deadline:
			t.Fatalf("signals missing: %v", want)
		}
	}
}
