package theme

import (
	"embed"
	"encoding/json"
	"fmt"
	"image/color"
	"reflect"
	"strconv"
	"strings"
)

// themes holds the registered palettes by name, in registration order.
var (
	themeNames = []string{"light", "dark"}
	themes     = map[string]func() Palette{"light": Light, "dark": Dark}
)

// builtin are the theme files Keel registers besides light and dark.
//
//go:embed themes/*.json
var builtin embed.FS

func init() {
	for _, name := range []string{"nord", "paper", "solarized-dark", "high-contrast"} {
		data, err := builtin.ReadFile("themes/" + name + ".json")
		if err != nil {
			panic(err)
		}
		n, p, err := ParseTheme(data)
		if err != nil {
			panic(err)
		}
		Register(n, p)
	}
}

// Register adds a named palette, or replaces one, for Named and Names; an app
// offers them in its theme picker. Built in: light, dark, nord, paper,
// solarized-dark and high-contrast.
func Register(name string, p Palette) {
	if _, ok := themes[name]; !ok {
		themeNames = append(themeNames, name)
	}
	themes[name] = func() Palette { return p }
}

// Named returns a copy of a registered palette.
func Named(name string) (Palette, bool) {
	f, ok := themes[name]
	if !ok {
		return Palette{}, false
	}
	return f(), true
}

// Names lists the registered palettes in registration order.
func Names() []string { return append([]string(nil), themeNames...) }

// ParseTheme reads a theme file: a base palette ("light" or "dark", or any
// registered name) and the colors it changes, by Palette field name in any
// case, as #rgb, #rgba, #rrggbb or #rrggbbaa. Chart takes up to eight colors.
//
//	{"name": "Nord", "base": "dark",
//	 "colors": {"bg": "#2e3440", "surface": "#3b4252", "primary": "#88c0d0"}}
func ParseTheme(data []byte) (name string, p Palette, err error) {
	var file struct {
		Name   string                     `json:"name"`
		Base   string                     `json:"base"`
		Colors map[string]json.RawMessage `json:"colors"`
	}
	if err := json.Unmarshal(data, &file); err != nil {
		return "", Palette{}, err
	}
	if file.Name == "" {
		return "", Palette{}, fmt.Errorf("theme has no name")
	}
	base := file.Base
	if base == "" {
		base = "light"
	}
	p, ok := Named(base)
	if !ok {
		return "", Palette{}, fmt.Errorf("theme %q: unknown base %q", file.Name, base)
	}
	fields := map[string]reflect.Value{}
	v := reflect.ValueOf(&p).Elem()
	for i := range v.NumField() {
		fields[strings.ToLower(v.Type().Field(i).Name)] = v.Field(i)
	}
	for key, raw := range file.Colors {
		f, ok := fields[strings.ToLower(key)]
		if !ok {
			return "", Palette{}, fmt.Errorf("theme %q: unknown color %q", file.Name, key)
		}
		if f.Type() == reflect.TypeOf(p.Chart) {
			var hexes []string
			if err := json.Unmarshal(raw, &hexes); err != nil || len(hexes) > len(p.Chart) {
				return "", Palette{}, fmt.Errorf("theme %q: chart takes up to %d colors", file.Name, len(p.Chart))
			}
			for i, h := range hexes {
				c, err := parseHex(h)
				if err != nil {
					return "", Palette{}, fmt.Errorf("theme %q: chart %d: %w", file.Name, i, err)
				}
				p.Chart[i] = c
			}
			continue
		}
		var h string
		if err := json.Unmarshal(raw, &h); err != nil {
			return "", Palette{}, fmt.Errorf("theme %q: %s: want a hex string", file.Name, key)
		}
		c, err := parseHex(h)
		if err != nil {
			return "", Palette{}, fmt.Errorf("theme %q: %s: %w", file.Name, key, err)
		}
		f.Set(reflect.ValueOf(c))
	}
	return file.Name, p, nil
}

func parseHex(s string) (color.NRGBA, error) {
	h := strings.TrimPrefix(s, "#")
	if len(h) == 3 || len(h) == 4 { // #rgb, #rgba
		var b []byte
		for i := range len(h) {
			b = append(b, h[i], h[i])
		}
		h = string(b)
	}
	if len(h) == 6 {
		h += "ff"
	}
	n, err := strconv.ParseUint(h, 16, 32)
	if len(h) != 8 || err != nil {
		return color.NRGBA{}, fmt.Errorf("bad color %q", s)
	}
	return color.NRGBA{R: uint8(n >> 24), G: uint8(n >> 16), B: uint8(n >> 8), A: uint8(n)}, nil
}
