// Command keel-mcp is an MCP server that lets an agent run end-to-end tests
// against a Keel app: launch it, read what is on screen, click, type, press
// keys, scroll and take screenshots.
//
// It starts the app with KEEL_AUTOMATION set, so the app renders off-screen and
// answers requests on a unix socket (see ui/window/automation_server.go). It
// imports no Keel package; the two sides share only that JSON protocol.
//
//	claude mcp add keel -- go run github.com/dyike/keel/cmd/keel-mcp
package main

import (
	"context"
	"log"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func main() {
	app := &appProcess{}
	defer app.stop()

	s := mcp.NewServer(&mcp.Implementation{Name: "keel", Version: "0.1.0"}, &mcp.ServerOptions{
		Instructions: instructions,
	})
	registerTools(s, app)
	if err := s.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Print(err)
	}
}

const instructions = `Drives a Keel desktop app for end-to-end tests.
Start with launch, or attach to an app the user started with
KEEL_AUTOMATION=1. Every action returns the window's elements; each has a ref
(e1, e2, ...) valid until the next action. Prefer refs or visible text over
coordinates. Coordinates are in dp; screenshots use 1 px = 1 dp. launch runs
the app off-screen unless visible is set; an attached app shows its windows if
the user started it without KEEL_HEADLESS=1, and the user sees every action.`
