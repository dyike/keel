package main

import (
	"testing"

	"github.com/dyike/keel/ui/window/windowtest"
)

func TestTemplateToStudio(t *testing.T) {
	s := newStudio()
	app := windowtest.New(t, options(s))
	app.Click("Browse templates")
	app.Click("Use Signal")
	app.Want("Signal / preview")
	app.Want("Palette · Tangerine")
	app.Click("Iris")
	app.Want("Palette · Iris")
	app.Click("Next frame")
	app.Want("002 / 120")
	app.Click("Previous frame")
	app.Click("Previous frame")
	app.Want("001 / 120")
	app.Click("Keel Motion home")
	app.Want("Make videos")
	app.Want("Signal / preview")
}

func TestFilmTimelineLength(t *testing.T) {
	s := newStudio()
	app := windowtest.New(t, options(s))
	app.Click("Open studio")
	app.Click(filmName)
	app.Want("001 / 810")
	s.timeline.SetValue(809)
	if got := s.timeline.Value(); got != 809 {
		t.Fatalf("film timeline stops at frame %g", got)
	}
	s.seek(809)
	app.Click("Next frame")
	app.Want("810 / 810")
	app.Click("Orbit")
	app.Want("001 / 120")
}
