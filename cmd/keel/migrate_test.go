package main

import "testing"

func TestMigrateSource(t *testing.T) {
	src := `package app

import (
	"fmt"

	"gioui.org/layout"
	op "gioui.org/op"
	"gioui.org/shader"
	gotext "github.com/go-text/typesetting/font"
	// "gioui.org/widget" stays: it is a comment
)

var _ = "gioui.org/app" // strings stay too
`
	want := `package app

import (
	"fmt"

	"github.com/dyike/keel/third_party/gio/layout"
	op "github.com/dyike/keel/third_party/gio/op"
	"gioui.org/shader"
	gotext "github.com/dyike/keel/third_party/typesetting/font"
	// "gioui.org/widget" stays: it is a comment
)

var _ = "gioui.org/app" // strings stay too
`
	out, err := migrateSource("app.go", []byte(src))
	if err != nil {
		t.Fatal(err)
	}
	if string(out) != want {
		t.Fatalf("got\n%s", out)
	}
	if out, err := migrateSource("other.go", []byte("package app\n\nimport \"fmt\"\n")); err != nil || out != nil {
		t.Fatalf("unchanged file rewritten: %q %v", out, err)
	}
}
