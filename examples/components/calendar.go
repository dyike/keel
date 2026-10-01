package main

import (
	"time"

	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
)

func init() {
	registerSection("calendar", "inputs", func() core.Widget {
		one := kit.Calendar().DisableDates(func(t time.Time) bool { return t.Weekday() == time.Sunday })
		span := kit.Calendar().Range()
		return el.Embed(el.ViewFunc(func(cx *el.Context) el.Element {
			return el.Div().P(24).Row().Gap(32).Items(el.Start).Child(one.Render(cx), span.Render(cx))
		}))
	})
}
