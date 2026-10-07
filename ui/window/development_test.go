package window

import "testing"

func TestDevelopmentShutdownRunsCloseCallbacksOnce(t *testing.T) {
	calls := 0
	open := newWindow(Options{OnClose: func() { calls++ }})
	alreadyClosed := newWindow(Options{OnClose: func() { t.Fatal("called already closed window") }})
	alreadyClosed.closed = true
	closeDevelopmentWindows([]*Window{open, alreadyClosed})
	closeDevelopmentWindows([]*Window{open, alreadyClosed})
	if calls != 1 || !open.closed || open.opts.OnClose != nil {
		t.Fatal("close callback not run exactly once", calls)
	}
}
