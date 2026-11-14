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

func TestSharedDemoOnlyPublishesPairedAssetsAndFullFallback(t *testing.T) {
	input, out := t.TempDir(), t.TempDir()
	for name, data := range map[string]string{
		"shared-build.json": `{"primary":"primary.abc.wasm","loader":"shared.abc.js","startup":"start.abc.js","fullWasm":"full.abc.wasm","fullScript":"full.abc.js","sections":{"input":"inputs.abc.wasm"}}`,
		"primary.abc.wasm":  "primary", "shared.abc.js": "shared loader", "start.abc.js": "start",
		"full.abc.wasm": "full runtime", "full.abc.js": "full start", "inputs.abc.wasm": "input child",
		"font.otf": "font", "baseline-opt.wasm": "must not ship", "functions.txt": "must not ship",
	} {
		if err := os.WriteFile(filepath.Join(input, name), []byte(data), 0644); err != nil {
			t.Fatal(err)
		}
	}
	if err := (&site{out: out}).copyDemo(input); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(out, "demo/assets.json"))
	if err != nil {
		t.Fatal(err)
	}
	var config demoConfig
	if err := json.Unmarshal(data, &config); err != nil {
		t.Fatal(err)
	}
	if config.Shared == nil || config.Assets["main.wasm"] != config.Assets[config.Shared.Primary] || config.Assets["wasm.js"] != config.Assets[config.Shared.Startup] {
		t.Fatal("missing shared runtime aliases")
	}
	if config.Assets[config.Shared.FullWasm].Bytes == 0 || config.Assets[config.Shared.FullScript].Bytes == 0 {
		t.Fatal("missing complete fallback")
	}
	for _, name := range []string{"baseline-opt.wasm", "functions.txt", "shared-build.json"} {
		if _, ok := config.Assets[name]; ok {
			t.Fatal("shipped build-only artifact", name)
		}
	}
	if err := os.Remove(filepath.Join(input, "inputs.abc.wasm")); err != nil {
		t.Fatal(err)
	}
	if err := (&site{out: t.TempDir()}).copyDemo(input); err == nil {
		t.Fatal("accepted missing child")
	}
}
