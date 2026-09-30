package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Wire types, mirroring ui/window's automation protocol.
type windowInfo struct {
	ID     string `json:"id"`
	Title  string `json:"title"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
	Active bool   `json:"active"`
}

type element struct {
	Ref      string `json:"ref"`
	Role     string `json:"role"`
	Name     string `json:"name"`
	Value    string `json:"value"`
	Checked  *bool  `json:"checked"`
	Selected *bool  `json:"selected"`
	Disabled bool   `json:"disabled"`
	X        int    `json:"x"`
	Y        int    `json:"y"`
	Width    int    `json:"width"`
	Height   int    `json:"height"`
}

type snapshot struct {
	Window   windowInfo `json:"window"`
	Elements []element  `json:"elements"`
}

// Tool inputs. The jsonschema tags become the descriptions the agent sees.
type (
	launchIn struct {
		Command        string `json:"command" jsonschema:"shell command that starts the app, e.g. 'go run ./examples/hello' or a built binary"`
		Dir            string `json:"dir,omitempty" jsonschema:"working directory for the command; default is keel-mcp's own working directory"`
		TimeoutSeconds int    `json:"timeout_seconds,omitempty" jsonschema:"how long to wait for the app to start, including compiling; default 120"`
		Visible        bool   `json:"visible,omitempty" jsonschema:"show the app's windows on screen so the user can watch; default renders off-screen"`
	}
	attachIn struct {
		Socket string `json:"socket,omitempty" jsonschema:"socket path the app printed at startup; omit to find running apps"`
	}
	windowIn struct {
		Window string `json:"window,omitempty" jsonschema:"window id such as w2; default is the active window"`
	}
	clickIn struct {
		Window string   `json:"window,omitempty" jsonschema:"window id; default is the active window"`
		Ref    string   `json:"ref,omitempty" jsonschema:"element ref from the latest snapshot, e.g. e4"`
		Text   string   `json:"text,omitempty" jsonschema:"visible text or value of the element; exact matches and controls win"`
		X      *float64 `json:"x,omitempty" jsonschema:"x in dp, only when there is no ref or text"`
		Y      *float64 `json:"y,omitempty" jsonschema:"y in dp, only when there is no ref or text"`
	}
	typeIn struct {
		Window string `json:"window,omitempty" jsonschema:"window id; default is the active window"`
		Text   string `json:"text" jsonschema:"text to insert at the caret"`
		Ref    string `json:"ref,omitempty" jsonschema:"textbox to click first; omit to type into the focused one"`
		Clear  bool   `json:"clear,omitempty" jsonschema:"select all existing text first so the new text replaces it"`
	}
	pressIn struct {
		Window string `json:"window,omitempty" jsonschema:"window id; default is the active window"`
		Key    string `json:"key" jsonschema:"key chord: enter, esc, tab, shift+tab, space, backspace, up, a, mod+a (mod is Cmd on macOS), ctrl+shift+s"`
	}
	scrollIn struct {
		Window string   `json:"window,omitempty" jsonschema:"window id; default is the active window"`
		DY     float64  `json:"dy" jsonschema:"distance in dp; positive scrolls down, negative up"`
		Ref    string   `json:"ref,omitempty" jsonschema:"element to scroll over; default is the window center"`
		X      *float64 `json:"x,omitempty" jsonschema:"x in dp to scroll at"`
		Y      *float64 `json:"y,omitempty" jsonschema:"y in dp to scroll at"`
	}
	waitIn struct {
		Window    string `json:"window,omitempty" jsonschema:"window id; default is the active window"`
		Text      string `json:"text" jsonschema:"text to wait for in any element's name or value"`
		TimeoutMS int    `json:"timeout_ms,omitempty" jsonschema:"default 5000"`
	}
	logsIn struct {
		Lines int `json:"lines,omitempty" jsonschema:"how many trailing lines; default 100"`
	}
)

func registerTools(s *mcp.Server, app *appProcess) {
	// act runs an automation request that returns a snapshot and formats it.
	act := func(method string, params map[string]any) (*mcp.CallToolResult, any, error) {
		var snap snapshot
		if err := app.call(method, params, &snap); err != nil {
			return nil, nil, err
		}
		return text(app.describe(snap)), nil, nil
	}

	mcp.AddTool(s, &mcp.Tool{Name: "launch", Description: "Start the app under test and return its first window: off-screen by default, or on screen with visible. Stops any app started earlier."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in launchIn) (*mcp.CallToolResult, any, error) {
			dir := in.Dir
			if dir == "" {
				dir, _ = os.Getwd()
			}
			timeout := time.Duration(in.TimeoutSeconds) * time.Second
			if timeout <= 0 {
				timeout = 120 * time.Second
			}
			if err := app.launch(in.Command, dir, timeout, in.Visible); err != nil {
				return nil, nil, err
			}
			return act("snapshot", nil)
		})

	mcp.AddTool(s, &mcp.Tool{Name: "attach", Description: "Connect to an app the user already started with KEEL_AUTOMATION=1 (or a socket path), in whatever state it is in. With no socket, picks the only running app or lists them. stop only disconnects; it never ends an attached app."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in attachIn) (*mcp.CallToolResult, any, error) {
			sock, err := app.attach(in.Socket)
			if err != nil {
				return nil, nil, err
			}
			var ws []windowInfo
			if err := app.call("windows", nil, &ws); err != nil {
				return nil, nil, err
			}
			res, _, err := act("snapshot", nil)
			if err != nil {
				return nil, nil, err
			}
			head := "Attached to " + sock + "\n"
			if len(ws) > 1 {
				head += app.windowList(ws, "") // snapshot shows the active one
			}
			res.Content[0].(*mcp.TextContent).Text = head + res.Content[0].(*mcp.TextContent).Text
			return res, nil, nil
		})

	mcp.AddTool(s, &mcp.Tool{Name: "snapshot", Description: "List the elements of a window: ref, role, name, value, state and bounds."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in windowIn) (*mcp.CallToolResult, any, error) {
			return act("snapshot", map[string]any{"window": in.Window})
		})

	mcp.AddTool(s, &mcp.Tool{Name: "click", Description: "Left-click an element (by ref or text) or a point. Returns the window's elements afterwards."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in clickIn) (*mcp.CallToolResult, any, error) {
			return act("click", map[string]any{"window": in.Window, "ref": in.Ref, "text": in.Text, "x": in.X, "y": in.Y})
		})

	mcp.AddTool(s, &mcp.Tool{Name: "type", Description: "Type text into a textbox. Pass ref to focus it first and clear to replace its content."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in typeIn) (*mcp.CallToolResult, any, error) {
			return act("type", map[string]any{"window": in.Window, "ref": in.Ref, "text": in.Text, "clear": in.Clear})
		})

	mcp.AddTool(s, &mcp.Tool{Name: "press_key", Description: "Press and release a key chord in a window, e.g. enter, tab, esc, mod+s. Window shortcuts fire; Tab moves focus."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in pressIn) (*mcp.CallToolResult, any, error) {
			return act("press", map[string]any{"window": in.Window, "key": in.Key})
		})

	mcp.AddTool(s, &mcp.Tool{Name: "scroll", Description: "Scroll with the mouse wheel over an element or point (default: window center)."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in scrollIn) (*mcp.CallToolResult, any, error) {
			return act("scroll", map[string]any{"window": in.Window, "dy": in.DY, "ref": in.Ref, "x": in.X, "y": in.Y})
		})

	mcp.AddTool(s, &mcp.Tool{Name: "wait_for", Description: "Wait until an element's name or value contains text, e.g. the result of background work."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in waitIn) (*mcp.CallToolResult, any, error) {
			return act("wait_for", map[string]any{"window": in.Window, "text": in.Text, "timeout_ms": in.TimeoutMS})
		})

	mcp.AddTool(s, &mcp.Tool{Name: "screenshot", Description: "PNG of a window as it is rendered. 1 px = 1 dp, the same coordinates click and scroll use."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in windowIn) (*mcp.CallToolResult, any, error) {
			var b64 string
			if err := app.call("screenshot", map[string]any{"window": in.Window}, &b64); err != nil {
				return nil, nil, err
			}
			png, err := base64.StdEncoding.DecodeString(b64)
			if err != nil {
				return nil, nil, err
			}
			return &mcp.CallToolResult{Content: []mcp.Content{&mcp.ImageContent{Data: png, MIMEType: "image/png"}}}, nil, nil
		})

	mcp.AddTool(s, &mcp.Tool{Name: "windows", Description: "List open windows."},
		func(ctx context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
			var ws []windowInfo
			if err := app.call("windows", nil, &ws); err != nil {
				return nil, nil, err
			}
			return text(app.windowList(ws, "")), nil, nil
		})

	mcp.AddTool(s, &mcp.Tool{Name: "close_window", Description: "Close a window as its close button would. Closing the last window ends the app."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in windowIn) (*mcp.CallToolResult, any, error) {
			var ws []windowInfo
			if err := app.call("close", map[string]any{"window": in.Window}, &ws); err != nil {
				return nil, nil, err
			}
			if len(ws) == 0 {
				return text("Closed the last window; the app is exiting."), nil, nil
			}
			return text("Closed.\n" + app.windowList(ws, "")), nil, nil
		})

	mcp.AddTool(s, &mcp.Tool{Name: "logs", Description: "The app's recent stdout and stderr: panics, log output, compile errors. Only for apps started with launch."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in logsIn) (*mcp.CallToolResult, any, error) {
			if app.isAttached() {
				return text("This app was started by the user, so its output is in the terminal that started it."), nil, nil
			}
			n := in.Lines
			if n <= 0 {
				n = 100
			}
			return text(app.logs.tail(n)), nil, nil
		})

	mcp.AddTool(s, &mcp.Tool{Name: "stop", Description: "End the app started with launch, or disconnect from an attached app (which keeps running)."},
		func(ctx context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
			if app.isAttached() {
				app.stop()
				return text("Disconnected; the app keeps running."), nil, nil
			}
			app.stop()
			return text("Stopped."), nil, nil
		})
}

func text(s string) *mcp.CallToolResult {
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: s}}}
}

// describe renders a snapshot as compact text for the agent:
//
//	Window w1 "Keel" 640×480
//	e4 textbox "你的名字" value="" @42,148 546×40
func (a *appProcess) describe(s snapshot) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Window %s %q %d×%d\n", s.Window.ID, s.Window.Title, s.Window.Width, s.Window.Height)
	for _, e := range s.Elements {
		fmt.Fprintf(&b, "%s %s %q", e.Ref, e.Role, e.Name)
		if e.Role == "textbox" || e.Value != "" {
			fmt.Fprintf(&b, " value=%q", e.Value)
		}
		if e.Selected != nil && *e.Selected {
			b.WriteString(" selected")
		}
		if e.Checked != nil {
			if *e.Checked {
				b.WriteString(" checked")
			} else {
				b.WriteString(" unchecked")
			}
		}
		if e.Disabled {
			b.WriteString(" disabled")
		}
		fmt.Fprintf(&b, " @%d,%d %d×%d\n", e.X, e.Y, e.Width, e.Height)
	}
	var ws []windowInfo
	if err := a.call("windows", nil, &ws); err == nil && len(ws) > 1 {
		b.WriteString(a.windowList(ws, s.Window.ID))
	} else {
		a.markSeen(ws)
	}
	return b.String()
}

// windowList lists windows, flagging ones not reported before.
func (a *appProcess) windowList(ws []windowInfo, current string) string {
	a.mu.Lock()
	seen := a.seen
	a.mu.Unlock()
	var b strings.Builder
	b.WriteString("Open windows:")
	for _, w := range ws {
		fmt.Fprintf(&b, " %s %q", w.ID, w.Title)
		switch {
		case w.ID == current:
			b.WriteString(" (shown above)")
		case !seen[w.ID]:
			b.WriteString(" (new)")
		}
		b.WriteString(";")
	}
	a.markSeen(ws)
	return strings.TrimSuffix(b.String(), ";") + "\n"
}

func (a *appProcess) markSeen(ws []windowInfo) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.seen == nil {
		a.seen = map[string]bool{}
	}
	for _, w := range ws {
		a.seen[w.ID] = true
	}
}

func (a *appProcess) isAttached() bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.attached != ""
}
