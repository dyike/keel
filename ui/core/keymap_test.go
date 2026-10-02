package core

import (
	"slices"
	"testing"
)

func TestKeymapBindAndLoad(t *testing.T) {
	t.Cleanup(func() { Bind("test.save"); Bind("test.quit") })
	if err := Bind("test.save", "mod+s", "ctrl+x"); err != nil {
		t.Fatal(err)
	}
	if got := Bindings("test.save"); !slices.Equal(got, []string{"mod+s", "ctrl+x"}) {
		t.Fatalf("bindings %v", got)
	}
	if Bind("test.save", "mod+") == nil || len(Bindings("test.save")) != 2 {
		t.Fatal("an invalid chord replaced the binding")
	}
	if err := LoadKeymap([]byte(`{"test.save": [], "test.quit": ["mod+q"]}`)); err != nil {
		t.Fatal(err)
	}
	if Bindings("test.save") != nil || !slices.Equal(Bindings("test.quit"), []string{"mod+q"}) {
		t.Fatalf("after load: save %v quit %v", Bindings("test.save"), Bindings("test.quit"))
	}
	if LoadKeymap([]byte(`{"test.quit": ["mod+w"], "test.save": ["shift+"]}`)) == nil || !slices.Equal(Bindings("test.quit"), []string{"mod+q"}) {
		t.Fatal("a keymap with an invalid entry was partly applied")
	}
	if _, ok := Keymap()["test.quit"]; !ok {
		t.Fatal("Keymap is missing a binding")
	}
}
