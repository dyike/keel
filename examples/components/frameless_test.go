package main

import (
	"github.com/dyike/keel/ui/locale"
	"github.com/dyike/keel/ui/window/windowtest"
	"testing"
)

func TestMergedFramelessControls(t *testing.T) {
	previous := locale.Current()
	locale.Apply(locale.English())
	defer locale.Apply(previous)
	opts := framelessOptions()
	if !opts.Frameless {
		t.Fatal("preview must be a real frameless window")
	}
	app := windowtest.New(t, opts)
	app.Want("Keel Frameless Window")
	app.Want(locale.Current().Close)
	app.Click("Search")
	app.Type("keel")
	app.Click("Toggle sidebar")
	app.Want("keel")
	app.Click("Commands")
	app.Want("New order")
	app.Press("escape")
	app.Press("mod+k")
	app.Want("New order")
}
