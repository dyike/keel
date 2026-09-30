package hotkey

import (
	"errors"
	"testing"

	"github.com/dyike/keel/native"
	"github.com/dyike/keel/native/internal/sys"
)

func TestParse(t *testing.T) {
	key, mods, err := parse("Cmd+Shift+K")
	if err != nil || key != "k" || mods != sys.ModCmd|sys.ModShift {
		t.Fatalf("got %q %b %v", key, mods, err)
	}
	for _, bad := range []string{"k", "cmd+", "cmd+a+b", ""} {
		if _, _, err := parse(bad); !errors.Is(err, native.ErrInvalidArgument) {
			t.Errorf("%q: got %v", bad, err)
		}
	}
}
