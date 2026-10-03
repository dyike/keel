package core

import (
	"slices"
	"testing"
)

func TestContextBindingsInheritanceAndOwnership(t *testing.T) {
	const action = "test.context.save"
	t.Cleanup(func() { Bind(action); ClearBindingIn("outer", action); ClearBindingIn("inner", action) })
	Bind(action, "mod+s")
	BindIn("outer", action, "f6")
	if got := BindingsIn(action, "inner", "outer"); !slices.Equal(got, []string{"f6"}) {
		t.Fatal(got)
	}
	chords := []string{"f7", "f8"}
	if err := BindIn("inner", action, chords...); err != nil {
		t.Fatal(err)
	}
	chords[0] = "f9"
	got := BindingsIn(action, "inner", "outer")
	if !slices.Equal(got, []string{"f7", "f8"}) {
		t.Fatal("borrowed input", got)
	}
	got[0] = "f10"
	if err := BindIn("inner", action, "mod+"); err == nil || BindingsIn(action, "inner")[0] != "f7" {
		t.Fatal("invalid binding changed state")
	}
	if err := BindIn("", action, "f6"); err == nil {
		t.Fatal("empty context accepted")
	}
	BindIn("inner", action)
	if len(BindingsIn(action, "inner", "outer")) != 0 {
		t.Fatal("explicit disable fell through")
	}
	ClearBindingIn("inner", action)
	if BindingsIn(action, "inner", "outer")[0] != "f6" {
		t.Fatal("did not restore outer inheritance")
	}
	ClearBindingIn("outer", action)
	if BindingsIn(action, "inner", "outer")[0] != "mod+s" {
		t.Fatal("did not restore global inheritance")
	}
}
