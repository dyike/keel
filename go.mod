module github.com/dyike/keel

go 1.26.1

// v0.0.1 and v0.0.2 were published before Keel had a license.
// Use a licensed release; see LICENSE in the chosen version.
retract [v0.0.1, v0.0.2]

require (
	gioui.org v0.10.3
	github.com/alecthomas/chroma/v2 v2.27.0
	github.com/ebitengine/purego v0.11.1
	github.com/go-text/typesetting v0.3.5
	github.com/godbus/dbus/v5 v5.2.2
	github.com/jezek/xgb v1.3.1
	github.com/modelcontextprotocol/go-sdk v1.8.0
	github.com/srwiley/oksvg v0.0.0-20221011165216-be6e8873101c
	github.com/srwiley/rasterx v0.0.0-20210519020934-456a8d69b780
	github.com/tc-hib/winres v0.3.1
	github.com/yuin/goldmark v1.8.6
	golang.org/x/exp/shiny v0.0.0-20250408133849-7e4ce0ab07d0
	golang.org/x/image v0.46.0
	golang.org/x/net v0.58.0
	golang.org/x/sys v0.48.0
)

require (
	gioui.org/shader v1.0.9 // indirect
	github.com/dlclark/regexp2/v2 v2.2.1 // indirect
	github.com/google/jsonschema-go v0.4.3 // indirect
	github.com/nfnt/resize v0.0.0-20180221191011-83c6a9932646 // indirect
	github.com/segmentio/asm v1.1.3 // indirect
	github.com/segmentio/encoding v0.5.4 // indirect
	github.com/yosida95/uritemplate/v3 v3.0.2 // indirect
	golang.org/x/oauth2 v0.35.0 // indirect
	golang.org/x/sync v0.23.0 // indirect
	golang.org/x/text v0.42.0 // indirect
	golang.org/x/time v0.15.0 // indirect
)

// Keel draws with its own copies of Gio and go-text, in third_party: fixes
// and speedups that cannot wait for upstream releases (third_party/README.md
// lists them). A replace only applies to the module that writes it: apps
// add the same lines to use the copies too (docs/getting-started.md).
replace (
	gioui.org => ./third_party/gio
	github.com/go-text/typesetting => ./third_party/typesetting
)
