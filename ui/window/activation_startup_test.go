package window

import "testing"

func TestStartupTime(t *testing.T) {
	for id, want := range map[string]uint32{
		"gnome-shell/keel/1234-0-host_TIME98765": 98765,
		"kde_TIME42_extra":                       42,
		"no-time":                                0,
		"_TIME":                                  0,
		"x_TIME99999999999":                      0, // past 32 bits
	} {
		if got := startupTime(id); got != want {
			t.Errorf("%q: %d, want %d", id, got, want)
		}
	}
	if d := activeWindowData("a_TIME7"); d[0] != 1 || d[1] != 7 {
		t.Fatal("source and timestamp", d)
	}
}
