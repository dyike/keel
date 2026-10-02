package base

import (
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

// TypeaheadPause is how long a pause in typing starts a new search.
const TypeaheadPause = time.Second

// Typeahead finds an item by the first letters of its label as they are
// typed, like a native list: typing "re" goes to "Red", and pressing the
// same letter again cycles through the items that start with it.
type Typeahead struct {
	typed string
	last  time.Time
}

// Text is the key text a typeahead takes from a key name: a single letter,
// digit or symbol pressed without Ctrl, Cmd or Alt. Gio names letter keys
// in upper case; the search ignores case.
func Text(name string, shortcut bool) (string, bool) {
	r, n := utf8.DecodeRuneInString(name)
	if shortcut || n == 0 || n != len(name) {
		return "", false
	}
	// Gio names keys such as ↑ and ⏎ with one symbol; only letters, digits
	// and ASCII punctuation are typed text.
	if !unicode.IsLetter(r) && !unicode.IsDigit(r) && (r >= utf8.RuneSelf || !unicode.IsPrint(r) || r == ' ') {
		return "", false
	}
	return name, true
}

// Find adds text to the search at time now and returns the enabled item
// whose label starts with it, searching from the item after current and
// wrapping. It returns current and false when nothing matches.
func (t *Typeahead) Find(now time.Time, text string, current int, l List, label func(int) string) (int, bool) {
	if now.Sub(t.last) > TypeaheadPause {
		t.typed = ""
	}
	t.last = now
	t.typed += strings.ToLower(text)
	query := t.typed
	start := current + 1
	if r, _ := utf8.DecodeRuneInString(query); strings.Count(query, string(r)) == utf8.RuneCountInString(query) {
		query = string(r) // the same letter again: the next item starting with it
	} else {
		start = current // a longer prefix: the current item may still match
	}
	for k := range l.Count {
		i := ((max(start, 0)+k)%l.Count + l.Count) % l.Count
		if l.enabled(i) && strings.HasPrefix(strings.ToLower(label(i)), query) {
			return i, true
		}
	}
	return current, false
}

// Reset forgets what was typed.
func (t *Typeahead) Reset() { t.typed = "" }
