package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"testing"
	"time"
)

// keel run asks a connected app, by process ID, to save its state and exit.
func TestReloadControl(t *testing.T) {
	ctl := newReloadControl(t.TempDir())
	if ctl == nil {
		t.Skip("no unix sockets here")
	}
	defer ctl.close()
	if env := strings.Join(ctl.env("/tmp/state.json"), " "); !strings.Contains(env, "KEEL_RUN_STATE=/tmp/state.json") || !strings.Contains(env, "KEEL_RUN_CONTROL="+ctl.path) {
		t.Fatalf("env %q", env)
	}
	if ctl.ask(42) {
		t.Fatal("asked an app that never connected")
	}
	conn, err := net.Dial("unix", ctl.path)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	fmt.Fprintf(conn, "hello 42\n")
	deadline := time.Now().Add(5 * time.Second)
	for !ctl.ask(42) {
		if time.Now().After(deadline) {
			t.Fatal("the app's hello was not registered")
		}
		time.Sleep(10 * time.Millisecond)
	}
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	line, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil || line != "reload\n" {
		t.Fatalf("app received %q, %v", line, err)
	}
	var nilCtl *reloadControl
	if nilCtl.ask(1) || len(nilCtl.env("s")) != 1 {
		t.Fatal("a nil control must fall back to signals")
	}
}
