module gioui.org

go 1.25.0

require (
	gioui.org/shader v1.0.9
	github.com/go-text/typesetting v0.3.5
	golang.org/x/exp/shiny v0.0.0-20250408133849-7e4ce0ab07d0
	golang.org/x/image v0.26.0
	golang.org/x/sys v0.39.0
	golang.org/x/text v0.32.0
)

require golang.org/x/net v0.48.0

require (
	github.com/ebitengine/purego v0.11.1
	github.com/godbus/dbus/v5 v5.2.2
)

// Keel: build against Keel's copy of go-text next to this one.
replace github.com/go-text/typesetting => ../typesetting
