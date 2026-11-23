//go:build !js

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/dyike/keel/ui/window"
)

func main() {
	output := flag.String("render", "", "write the composition to a new MP4 file and exit")
	scene := flag.String("scene", "Orbit", "Orbit, Signal, Bloom or Keel Film")
	palette := flag.String("palette", "", "Cobalt, Tangerine or Iris (defaults to the template palette)")
	web := flag.String("web", "", "serve a packaged WebAssembly site from this directory")
	port := flag.Int("port", 4173, "local web server port")
	flag.Parse()
	if *web != "" {
		log.Fatal(serveWebsite(*web, *port))
	}
	if *output != "" {
		if err := renderFile(*output, *scene, *palette); err != nil {
			log.Fatal(err)
		}
		return
	}
	window.Open(options(newStudio()))
	window.Main()
}

func renderFile(path, scene, palette string) error {
	c := composition{-1, -1}
	for i, t := range templates {
		if strings.EqualFold(t.name, scene) {
			c = composition{i, t.defaultPalette}
			break
		}
	}
	if palette != "" {
		c.Palette = -1
		for i, p := range palettes {
			if strings.EqualFold(p.name, palette) {
				c.Palette = i
				break
			}
		}
	}
	if !c.valid() {
		return fmt.Errorf("unknown scene %q or palette %q", scene, palette)
	}
	if _, err := os.Lstat(path); err == nil {
		return fmt.Errorf("%s already exists", path)
	} else if !os.IsNotExist(err) {
		return err
	}
	tmp, err := renderMovie(context.Background(), filepath.Dir(path), c)
	if err != nil {
		return err
	}
	defer os.Remove(tmp)
	return os.Link(tmp, path) // Atomically publish without replacing an existing movie.
}
