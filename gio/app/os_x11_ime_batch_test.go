//go:build linux && !android && !nox11

package app

import (
	"github.com/dyike/keel/gio/io/key"
	"golang.org/x/sys/unix"
	"sync"
	"testing"
)

func TestX11IMEBurstKeepsEveryUpdateAndCommit(t *testing.T) {
	var pipe [2]int
	if err := unix.Pipe2(pipe[:], unix.O_NONBLOCK|unix.O_CLOEXEC); err != nil {
		t.Fatal(err)
	}
	defer unix.Close(pipe[0])
	defer unix.Close(pipe[1])
	c, recorded := imeBatchCallbacks()
	w := &x11Window{w: c}
	w.notify.read, w.notify.write = pipe[0], pipe[1]
	var wg sync.WaitGroup
	for range 100 {
		wg.Add(1)
		go func() { defer wg.Done(); (*ibusEventHandler)(w).Event("ni", key.Range{Start: 2, End: 2}, true, true) }()
	}
	wg.Wait()
	(*ibusEventHandler)(w).Event("你好", key.Range{Start: -1, End: -1}, true, false)
	w.runFuncs.drain()
	if recorded.edits != 101 || c.EditorState().Snippet.Text != "A🙂你好Z" {
		t.Fatalf("IME burst lost edits or its final commit: edits=%d state=%+v", recorded.edits, c.EditorState())
	}
}
