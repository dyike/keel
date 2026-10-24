package window

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
)

type reloadApp struct {
	State struct{ Page string }
	name  string
}

func (a *reloadApp) Render(cx *el.Context) el.Element {
	return el.Div().Child(el.Input().Name("Name").Bind(&a.name), el.Text("Page "+a.State.Page))
}

// A reload keeps the window size, KeepState and the content's UI state, and
// consumes the file so a crash does not restore it twice.
func TestReloadStateAcrossRuns(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state.json")
	t.Setenv("KEEL_RUN_WATCH", "1")
	t.Setenv("KEEL_RUN_STATE", path)
	resetReload := func() {
		reload.Lock()
		reload.once, reload.saved, reload.titles = sync.Once{}, nil, nil
		reload.Unlock()
	}
	resetReload()
	defer resetReload()

	first := &reloadApp{name: "Ada"}
	first.State.Page = "settings"
	h := OpenHeadless(Options{Title: "Notes", Width: 640, Height: 480, Content: el.Root(first), KeepState: &first.State})
	h.Resize(700, 500)
	saveReloadState([]*Window{h.Window()})
	h.Close()
	b, err := os.ReadFile(path)
	if err != nil || !strings.Contains(string(b), `"id":"Notes#0"`) {
		t.Fatalf("state file: %v %s", err, b)
	}

	resetReload()
	second := &reloadApp{}
	second.State.Page = "home"
	h = OpenHeadless(Options{Title: "Notes", Width: 640, Height: 480, Content: el.Root(second), KeepState: &second.State})
	defer h.Close()
	if second.State.Page != "settings" {
		t.Fatalf("KeepState: page %q", second.State.Page)
	}
	if o := h.Window().opts; o.Width != 700 || o.Height != 500 {
		t.Fatalf("size %dx%d, want 700x500", o.Width, o.Height)
	}
	if e, err := h.Find("Name"); err != nil || e.Value != "Ada" || second.name != "Ada" {
		t.Fatalf("input after reload: %+v %v, bound %q", e, err, second.name)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatal("the state file was not consumed")
	}

	// A KeepState whose type changed leaves the defaults whole, though
	// encoding/json would set the fields that still decode.
	resetReload()
	os.WriteFile(path, mustJSON(t, reloadState{Windows: []reloadWindow{{ID: "Notes#0", App: json.RawMessage(`{"Page":"saved","Count":"x"}`)}}}), 0o600)
	var state struct {
		Page  string
		Count int
	}
	state.Page = "home"
	h3 := OpenHeadless(Options{Title: "Notes", Content: el.Root(&reloadApp{}), KeepState: &state})
	defer h3.Close()
	if state.Page != "home" || state.Count != 0 {
		t.Fatalf("partly decoded KeepState: %+v", state)
	}
}

func mustJSON(t *testing.T, v any) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

type kitReloadApp struct{ name *kit.InputView }

func (a *kitReloadApp) Render(cx *el.Context) el.Element { return el.Div().Child(a.name.Render(cx)) }

// kit components name their elements after their address, which changes
// every run; their state must come back anyway.
func TestReloadKeepsKitInput(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state.json")
	t.Setenv("KEEL_RUN_WATCH", "1")
	t.Setenv("KEEL_RUN_STATE", path)
	reset := func() {
		reload.Lock()
		reload.once, reload.saved, reload.titles = sync.Once{}, nil, nil
		reload.Unlock()
	}
	reset()
	defer reset()
	first := &kitReloadApp{name: kit.Input("Name")}
	h := OpenHeadless(Options{Title: "Kit", Content: el.Root(first)})
	if err := h.Click("Name"); err != nil {
		t.Fatal(err)
	}
	if err := h.Type("Ada"); err != nil {
		t.Fatal(err)
	}
	saveReloadState([]*Window{h.Window()})
	h.Close()

	reset()
	second := &kitReloadApp{name: kit.Input("Name")}
	h = OpenHeadless(Options{Title: "Kit", Content: el.Root(second)})
	defer h.Close()
	if e, err := h.Find("Name"); err != nil || e.Value != "Ada" || second.name.Value() != "Ada" {
		t.Fatalf("kit input after reload: %+v %v, Value %q", e, err, second.name.Value())
	}
}
