package sys

import (
	"sync"
	"unicode/utf16"
)

// Windows shows one balloon notification per tray icon, so the balloon
// center keeps the single notification on screen: posting replaces it,
// removing its ID or the balloon closing clears it, and a click runs its
// handler once. The shell itself is behind balloonShell so the rules are
// testable off Windows.

type balloonShell interface {
	show(title, body string) error // add or update the icon and its balloon
	hide() error                   // remove the icon and its balloon
}

type balloonEvent uint8

const (
	balloonClicked balloonEvent = iota
	balloonClosed               // timed out or dismissed
)

type balloonCenter struct {
	sync.Mutex
	shell   balloonShell
	current string
	onClick func()
	seq     uint64 // counts posts, so a late close of an old balloon is ignored
}

func (c *balloonCenter) post(id, title, body string, onClick func()) error {
	c.Lock()
	defer c.Unlock()
	if err := c.shell.show(balloonText(title, 63), balloonText(body, 255)); err != nil {
		return err
	}
	c.current, c.onClick = id, onClick
	c.seq++
	return nil
}

func (c *balloonCenter) remove(id string) error {
	c.Lock()
	defer c.Unlock()
	if c.current != id {
		return nil // already replaced or closed: removing is harmless
	}
	c.current, c.onClick = "", nil
	return c.shell.hide()
}

// event handles the shell's report about the balloon shown at post seq.
// It returns the click handler to run, outside the lock.
func (c *balloonCenter) event(e balloonEvent, seq uint64) func() {
	c.Lock()
	defer c.Unlock()
	if c.current == "" || seq != c.seq {
		return nil
	}
	fn := c.onClick
	c.current, c.onClick = "", nil
	c.shell.hide()
	if e == balloonClicked {
		return fn
	}
	return nil
}

func (c *balloonCenter) sequence() uint64 {
	c.Lock()
	defer c.Unlock()
	return c.seq
}

// balloonText fits s into a fixed UTF-16 field of max units, ending with an
// ellipsis when cut, never splitting a surrogate pair.
func balloonText(s string, max int) string {
	u := utf16.Encode([]rune(s))
	if len(u) <= max {
		return s
	}
	cut := max - 1
	if cut > 0 && utf16.IsSurrogate(rune(u[cut-1])) && u[cut-1] < 0xdc00 {
		cut--
	}
	return string(utf16.Decode(u[:cut])) + "…"
}
