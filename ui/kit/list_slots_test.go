package kit

import (
	"testing"

	"github.com/dyike/keel/ui/el"
)

func TestListStateSlots(t *testing.T) {
	retried := 0
	l := List().Searchable(true).
		EmptyContent(text("Nothing here yet")).
		NoMatchesContent(text("No fruit by that name")).
		LoadingContent(text("Fetching fruit")).
		ErrorContent(func(message string, retry func()) el.View {
			return viewFunc(func(cx *el.Context) el.Element {
				return el.Div().Child(el.Text("Oops: "+message), Button("Try again", func() { retried++; retry() }).Render(cx))
			})
		}).
		OnLoadMore(func() {})
	h := renderView(l, 400, 1)
	if !shown(h, "Nothing here yet") || shown(h, "No fruit by that name") {
		t.Fatal("empty list")
	}
	l.SetItems("apple", "pear")
	h.Frame()
	if shown(h, "Nothing here yet") || !shown(h, "apple") {
		t.Fatal("items replace the empty view")
	}
	l.SetQuery("kiwi")
	h.Frame()
	if !shown(h, "No fruit by that name") {
		t.Fatal("no matches")
	}
	l.SetQuery("")
	l.SetLoading(true)
	h.Frame()
	if !shown(h, "Fetching fruit") {
		t.Fatal("loading slot")
	}
	l.SetLoadError("offline")
	h.Frame()
	if !shown(h, "Oops: offline") {
		t.Fatal("error slot")
	}
	click(t, h, "Try again")
	if retried != 1 || !l.loading {
		t.Fatal("retry requests rows again", retried, l.loading)
	}

	// The initial view stands in for the rows until a query is typed.
	s := List("apple", "pear").Searchable(true).InitialContent(text("Type to search fruit"))
	h = renderView(s, 400, 1)
	if !shown(h, "Type to search fruit") || shown(h, "apple") {
		t.Fatal("initial view")
	}
	s.SetQuery("ap")
	h.Frame()
	if shown(h, "Type to search fruit") || !shown(h, "apple") {
		t.Fatal("a query shows results")
	}
}
