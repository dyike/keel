package main

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

type demoAsset struct {
	URL       string `json:"url"`
	Bytes     int    `json:"bytes"`
	GzipBytes int    `json:"gzipBytes"`
}

// Content-addressed assets separate cache lifetimes from the entry HTML.
// Gzip siblings are for hosts supporting precompressed files; publishing a .gz
// alone does not make GitHub Pages serve it with Content-Encoding.
func (s *site) demoAsset(name string, data []byte) (demoAsset, error) {
	sum := sha256.Sum256(data)
	ext := filepath.Ext(name)
	filename := strings.TrimSuffix(name, ext) + "." + hex.EncodeToString(sum[:8]) + ext
	var compressed bytes.Buffer
	z, _ := gzip.NewWriterLevel(&compressed, gzip.BestCompression)
	if _, err := z.Write(data); err != nil {
		return demoAsset{}, err
	}
	if err := z.Close(); err != nil {
		return demoAsset{}, err
	}
	for path, contents := range map[string][]byte{filename: data, filename + ".gz": compressed.Bytes()} {
		if err := s.write("demo/"+path, contents); err != nil {
			return demoAsset{}, err
		}
	}
	return demoAsset{URL: filename, Bytes: len(data), GzipBytes: compressed.Len()}, nil
}

type demoGroup struct {
	Wasm     string   `json:"wasm"`
	Script   string   `json:"script"`
	Sections []string `json:"sections"`
}

// Shared build metadata names assets paired with one Go runtime. The full
// program remains available through ?bundle=all during rollout.
type demoShared struct {
	Primary      string              `json:"primary"`
	Loader       string              `json:"loader"`
	Startup      string              `json:"startup"`
	FullWasm     string              `json:"fullWasm"`
	FullScript   string              `json:"fullScript"`
	Sections     map[string]string   `json:"sections"`
	Categories   map[string]string   `json:"categories,omitempty"`
	Dependencies map[string][]string `json:"dependencies,omitempty"`
}

type demoConfig struct {
	Shared   *demoShared          `json:"shared,omitempty"`
	Assets   map[string]demoAsset `json:"assets"`
	Groups   map[string]demoGroup `json:"groups,omitempty"`
	Sections map[string]string    `json:"sections,omitempty"`
}

func (s *site) copyDemo(dir string) error {
	config := demoConfig{Assets: map[string]demoAsset{}, Sections: map[string]string{}}
	if data, err := os.ReadFile(filepath.Join(dir, "groups.json")); err == nil {
		if err := json.Unmarshal(data, &config); err != nil {
			return err
		}
	} else if !os.IsNotExist(err) {
		return err
	}
	allowed := map[string]bool{}
	if data, err := os.ReadFile(filepath.Join(dir, "shared-build.json")); err == nil {
		var shared demoShared
		if err := json.Unmarshal(data, &shared); err != nil {
			return err
		}
		if len(config.Groups) != 0 {
			return fmt.Errorf("cannot mix shared and independent category manifests")
		}
		for _, name := range []string{shared.Primary, shared.Loader, shared.Startup, shared.FullWasm, shared.FullScript} {
			if !fs.ValidPath(name) || name == "." {
				return fmt.Errorf("invalid shared asset %q", name)
			}
			allowed[name] = true
		}
		if len(shared.Sections) == 0 {
			return fmt.Errorf("shared build has no section ownership")
		}
		children := make([]string, 0, len(shared.Sections)+len(shared.Categories))
		for _, name := range shared.Sections {
			children = append(children, name)
		}
		for _, name := range shared.Categories {
			children = append(children, name)
		}
		for _, deps := range shared.Dependencies {
			children = append(children, deps...)
		}
		for _, name := range children {
			if !fs.ValidPath(name) || name == "." {
				return fmt.Errorf("invalid child asset %q", name)
			}
			allowed[name] = true
		}
		config.Shared = &shared
	} else if !os.IsNotExist(err) {
		return err
	}
	err := filepath.WalkDir(dir, func(path string, e fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if e.IsDir() {
			return nil
		}
		name, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		name = filepath.ToSlash(name)
		if config.Shared != nil && !allowed[name] && name != "font.otf" {
			return nil
		}
		if e.Name() == "index.html" || e.Name() == "groups.json" || strings.HasSuffix(name, ".gz") || strings.HasSuffix(name, ".br") {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		a, err := s.demoAsset(name, data)
		if err != nil {
			return err
		}
		config.Assets[name] = a
		return nil
	})
	if err != nil {
		return err
	}
	if shared := config.Shared; shared != nil {
		for name := range allowed {
			if a, ok := config.Assets[name]; !ok || a.Bytes == 0 {
				return fmt.Errorf("shared build missing %s", name)
			}
		}
		config.Assets["main.wasm"] = config.Assets[shared.Primary]
		config.Assets["wasm.js"] = config.Assets[shared.Startup]
	}
	for _, required := range []string{"main.wasm", "wasm.js"} {
		if a, ok := config.Assets[required]; !ok || a.Bytes == 0 {
			return fmt.Errorf("%s has no %s; build with gogio -target js", dir, required)
		}
	}
	for name, group := range config.Groups {
		for _, file := range []string{group.Wasm, group.Script} {
			if a, ok := config.Assets[file]; !ok || a.Bytes == 0 {
				return fmt.Errorf("category %s: missing %s", name, file)
			}
		}
		config.Sections[name] = name
		for _, section := range group.Sections {
			config.Sections[section] = name
		}
	}
	loader, err := assets.ReadFile("assets/demo-loader.js")
	if err != nil {
		return err
	}
	script, err := s.demoAsset("loader.js", loader)
	if err != nil {
		return err
	}
	page, err := assets.ReadFile("assets/demo.html")
	if err != nil {
		return err
	}
	data, err := json.Marshal(config)
	if err != nil {
		return err
	}
	page = bytes.ReplaceAll(page, []byte("__KEEL_DEMO_CONFIG__"), data)
	page = bytes.ReplaceAll(page, []byte("__KEEL_DEMO_LOADER__"), []byte(script.URL))
	if err := s.write("demo/assets.json", data); err != nil {
		return err
	}
	return s.write("demo/index.html", page)
}
