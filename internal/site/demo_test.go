package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestDemoContentAddressingAndCategoryManifest(t *testing.T) {
	input, out := t.TempDir(), t.TempDir()
	write := func(name, data string) {
		t.Helper()
		p := filepath.Join(input, name)
		if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(data), 0644); err != nil {
			t.Fatal(err)
		}
	}
	write("main.wasm", "full runtime")
	write("wasm.js", "loader")
	write("font.otf", "font")
	write("groups/general/main.wasm", "small runtime")
	write("groups/general/wasm.js", "loader")
	write("groups.json", `{"groups":{"general":{"wasm":"groups/general/main.wasm","script":"groups/general/wasm.js","sections":["button"]}}}`)
	s := &site{out: out}
	build := func() demoConfig {
		t.Helper()
		if err := s.copyDemo(input); err != nil {
			t.Fatal(err)
		}
		b, err := os.ReadFile(filepath.Join(out, "demo/assets.json"))
		if err != nil {
			t.Fatal(err)
		}
		var c demoConfig
		if err = json.Unmarshal(b, &c); err != nil {
			t.Fatal(err)
		}
		return c
	}
	a := build()
	b := build()
	if a.Assets["main.wasm"] != b.Assets["main.wasm"] {
		t.Fatal("unstable content hash")
	}
	if a.Sections["button"] != "general" || a.Sections["general"] != "general" {
		t.Fatal("missing routing", a.Sections)
	}
	asset := a.Assets["groups/general/main.wasm"]
	gz, err := os.ReadFile(filepath.Join(out, "demo", asset.URL+".gz"))
	if err != nil {
		t.Fatal(err)
	}
	z, err := gzip.NewReader(bytes.NewReader(gz))
	if err != nil {
		t.Fatal(err)
	}
	defer z.Close()
	raw, err := io.ReadAll(z)
	if err != nil || string(raw) != "small runtime" {
		t.Fatal("bad gzip", err)
	}
	if asset.Bytes != len(raw) || asset.GzipBytes != len(gz) {
		t.Fatal("wrong size report")
	}
	write("main.wasm", "changed runtime")
	c := build()
	if c.Assets["main.wasm"].URL == a.Assets["main.wasm"].URL {
		t.Fatal("changed bytes reused old URL")
	}
	if c.Assets["font.otf"] != a.Assets["font.otf"] {
		t.Fatal("unrelated asset cache invalidated")
	}
}
func TestDemoRequiresRuntimeAndScript(t *testing.T) {
	input := t.TempDir()
	os.WriteFile(filepath.Join(input, "main.wasm"), []byte("wasm"), 0644)
	if err := (&site{out: t.TempDir()}).copyDemo(input); err == nil {
		t.Fatal("missing loader accepted")
	}
}
func TestDemoLoaderBehavior(t *testing.T) {
	node, err := exec.LookPath("node")
	if err != nil {
		t.Skip("Node is required for the loader's JS tests")
	}
	cmd := exec.Command(node, "--test", "testdata/demo-loader.test.cjs")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("%v\n%s", err, out)
	}
}
