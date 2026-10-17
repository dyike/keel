package main

import (
	"fmt"
	"maps"
	"math"
	"slices"
	"strconv"
	"strings"
)

// plistEntry is one top-level Info.plist key with its value as plist XML.
type plistEntry struct {
	key, value string
}

// infoPlistReserved maps keys keel writes from keel.json to the field that
// controls them. info_plist may not set these: the bundle would disagree with
// keel.json, and signing, permissions and notifications follow the bundle.
var infoPlistReserved = map[string]map[string]string{
	"macos": {
		"CFBundleDisplayName":           "name",
		"CFBundleName":                  "name",
		"CFBundleExecutable":            "binary",
		"CFBundleIconFile":              "icon",
		"CFBundleIdentifier":            "appid",
		"CFBundleInfoDictionaryVersion": "keel",
		"CFBundlePackageType":           "keel",
		"CFBundleShortVersionString":    "version",
		"CFBundleVersion":               "build",
		"LSMinimumSystemVersion":        "keel (native capabilities need macOS 14)",
		"LSUIElement":                   "macos.accessory",
	},
	"ios": {
		"CFBundleDisplayName":        "name",
		"CFBundleName":               "name",
		"CFBundleExecutable":         "binary",
		"CFBundleIdentifier":         "appid",
		"CFBundlePackageType":        "keel",
		"CFBundleShortVersionString": "version",
		"CFBundleVersion":            "build",
		"CFBundleSupportedPlatforms": "keel (simulator or device build)",
		"MinimumOSVersion":           "ios.minimum_version",
		"UIApplicationSceneManifest": "keel (KeelSceneDelegate)",
		"CFBundleIcons":              "icon",
		"CFBundleIcons~ipad":         "icon",
	},
}

// validateInfoPlist checks one platform's info_plist from keel.json.
func validateInfoPlist(platform string, extra map[string]any) []error {
	var errs []error
	for _, k := range slices.Sorted(maps.Keys(extra)) {
		where := fmt.Sprintf("%s.info_plist[%q]", platform, k)
		if strings.TrimSpace(k) == "" {
			errs = append(errs, fmt.Errorf("%s: empty key", platform+".info_plist"))
			continue
		}
		if field, ok := infoPlistReserved[platform][k]; ok {
			errs = append(errs, fmt.Errorf("%s is written by keel from %s; it cannot be overridden", where, field))
			continue
		}
		if _, err := plistValue(extra[k]); err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", where, err))
		}
	}
	return errs
}

// mergeInfoPlist overrides or appends info_plist keys. Keys keel only uses
// as defaults (orientations, device family, category) keep their position;
// new keys follow in sorted order so output is stable. validateInfoPlist
// has already rejected reserved keys and bad values.
func mergeInfoPlist(base []plistEntry, extra map[string]any) []plistEntry {
	out := slices.Clone(base)
	for _, k := range slices.Sorted(maps.Keys(extra)) {
		v, err := plistValue(extra[k])
		if err != nil {
			panic(fmt.Sprintf("info_plist %q passed validation: %v", k, err))
		}
		if i := slices.IndexFunc(out, func(e plistEntry) bool { return e.key == k }); i >= 0 {
			out[i].value = v
		} else {
			out = append(out, plistEntry{k, v})
		}
	}
	return out
}

func plistDocument(entries []plistEntry) string {
	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
`)
	for _, e := range entries {
		fmt.Fprintf(&b, "\t<key>%s</key>%s\n", xmlText(e.key), e.value)
	}
	b.WriteString("</dict>\n</plist>\n")
	return b.String()
}

// plistValue converts a JSON value to plist XML: strings, booleans, numbers
// (integral values become <integer>, others <real>), arrays and objects.
// JSON has no dates or binary data, so neither is supported.
func plistValue(v any) (string, error) {
	switch v := v.(type) {
	case string:
		return "<string>" + xmlText(v) + "</string>", nil
	case bool:
		if v {
			return "<true/>", nil
		}
		return "<false/>", nil
	case float64:
		if math.IsInf(v, 0) || math.IsNaN(v) {
			return "", fmt.Errorf("number %v is not finite", v)
		}
		if v == math.Trunc(v) && math.Abs(v) < 1<<53 {
			return "<integer>" + strconv.FormatInt(int64(v), 10) + "</integer>", nil
		}
		return "<real>" + strconv.FormatFloat(v, 'g', -1, 64) + "</real>", nil
	case []any:
		var b strings.Builder
		b.WriteString("<array>")
		for i, e := range v {
			s, err := plistValue(e)
			if err != nil {
				return "", fmt.Errorf("[%d]: %w", i, err)
			}
			b.WriteString(s)
		}
		b.WriteString("</array>")
		return b.String(), nil
	case map[string]any:
		var b strings.Builder
		b.WriteString("<dict>")
		for _, k := range slices.Sorted(maps.Keys(v)) {
			s, err := plistValue(v[k])
			if err != nil {
				return "", fmt.Errorf("%q: %w", k, err)
			}
			b.WriteString("<key>" + xmlText(k) + "</key>" + s)
		}
		b.WriteString("</dict>")
		return b.String(), nil
	case nil:
		return "", fmt.Errorf("null has no plist form; remove the key instead")
	default:
		return "", fmt.Errorf("unsupported value %T", v)
	}
}
