package window

import "testing"

func TestApplicationCleanupOrdersCallbacksOnce(t *testing.T) {
	// These windows need no platform resources, but are registered like real Open.
	var got []string
	a := newWindow(Options{OnClose: func() { got = append(got, "window") }})
	b := newWindow(Options{OnClose: func() { got = append(got, "window") }})
	registerApplicationWindow(a)
	registerApplicationWindow(b)
	defer unregisterApplicationWindow(a)
	defer unregisterApplicationWindow(b)
	application.Lock()
	oldQuit := application.onQuit
	application.onQuit = func() { got = append(got, "app") }
	application.Unlock()
	defer func() { application.Lock(); application.onQuit = oldQuit; application.Unlock() }()
	runApplicationCleanup()
	runApplicationCleanup()
	if len(got) != 3 || got[0] != "window" || got[1] != "window" || got[2] != "app" {
		t.Fatalf("cleanup order = %v", got)
	}
	if !a.closed || !b.closed {
		t.Fatal("exit did not mark windows closed")
	}
}
