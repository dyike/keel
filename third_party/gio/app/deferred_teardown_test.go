package app

import "testing"

func TestDeferredTeardownPreservesHandlesThroughViewCleanup(t *testing.T) {
	var d deferredTeardown
	releases := 0
	live, notified, cleaned, destroyed := true, false, false, false
	d.begin(func() { notified = true }, func() {
		releases++
		if !cleaned {
			t.Fatal("native connection destroyed before view consumer cleaned up")
		}
		live, destroyed = false, true
	})
	if !notified || !live || destroyed {
		t.Fatal("invalid-view notification must precede native teardown")
	}
	// Event returns the queued invalid view. Its consumer can still release
	// pointer/seat proxies here without accessing a disconnected display.
	d.complete() // queued notification does not acknowledge cleanup
	if !live {
		t.Fatal("destroyed before invalid view was returned")
	}
	d.viewReturned()
	cleaned = true
	d.begin(func() { t.Fatal("duplicate notification") }, func() { t.Fatal("replaced teardown") })
	d.complete() // the consumer calls Event again
	if live || !destroyed {
		t.Fatal("native handles still live after acknowledgment")
	}
	d.complete() // must not release native resources twice
	if releases != 1 {
		t.Fatalf("released %d times", releases)
	}
}
