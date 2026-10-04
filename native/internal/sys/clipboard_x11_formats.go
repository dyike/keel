package sys

import (
	"encoding/json"
	"net/url"
	"path"
	"strings"
	"unicode/utf8"
)

// collectXClipboard only requests advertised, side-effect-free formats. A nil
// target list means a legacy owner refused TARGETS; probe known formats.
func collectXClipboard(targets map[string]bool, read func(string) ([]byte, error)) ([]byte, error) {
	allowed := func(name string) bool { return targets == nil || targets[name] }
	var out clipboardSnapshot
	for _, name := range []string{"UTF8_STRING", "text/plain;charset=utf-8", "text/plain", "STRING"} {
		if !allowed(name) {
			continue
		}
		data, err := read(name)
		if err != nil {
			return nil, err
		}
		if data == nil {
			continue
		}
		if len(data) > clipboardByteLimit {
			return nil, clipboardFormatError("text exceeds 16MiB")
		}
		if name == "STRING" {
			var b strings.Builder
			for _, c := range data {
				b.WriteRune(rune(c))
			}
			out.Text = b.String()
		} else {
			if !utf8.Valid(data) {
				return nil, clipboardFormatError("invalid UTF-8 text")
			}
			out.Text = string(data)
		}
		break
	}
	for _, name := range []string{"text/uri-list", "x-special/gnome-copied-files"} {
		if !allowed(name) {
			continue
		}
		data, err := read(name)
		if err != nil {
			return nil, err
		}
		if data == nil {
			continue
		}
		out.Files, err = clipboardURIs(data, name == "x-special/gnome-copied-files")
		if err != nil {
			return nil, err
		}
		break
	}
	size := len(out.Text)
	for _, file := range out.Files {
		size += len(file)
	}
	if size > clipboardByteLimit {
		return nil, clipboardFormatError("snapshot exceeds 16MiB")
	}
	if len(out.Files) == 0 {
		for _, mime := range []string{"image/png", "image/jpeg", "image/tiff", "image/bmp", "image/webp"} {
			if !allowed(mime) {
				continue
			}
			data, err := read(mime)
			if err != nil {
				return nil, err
			}
			if data == nil {
				continue
			}
			if len(data) > clipboardByteLimit-size {
				return nil, clipboardFormatError("snapshot exceeds 16MiB")
			}
			out.Images = []clipboardImage{{MIME: mime, Data: data}}
			break
		}
	}
	return json.Marshal(out)
}

func clipboardURIs(data []byte, gnome bool) ([]string, error) {
	if len(data) > clipboardByteLimit || !utf8.Valid(data) {
		return nil, clipboardFormatError("invalid URI list")
	}
	lines := strings.Split(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n")
	if gnome {
		if len(lines) == 0 || lines[0] != "copy" && lines[0] != "cut" {
			return nil, clipboardFormatError("invalid file operation header")
		}
		lines = lines[1:] // References only: never perform the copy/cut operation.
	}
	var files []string
	for _, line := range lines {
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		u, err := url.Parse(line)
		if err != nil {
			return nil, clipboardFormatError("invalid file URI")
		}
		if u.Scheme != "file" || u.Host != "" && !strings.EqualFold(u.Host, "localhost") {
			continue
		}
		if u.User != nil || u.RawQuery != "" || u.Fragment != "" || !path.IsAbs(u.Path) || strings.ContainsRune(u.Path, 0) || !utf8.ValidString(u.Path) {
			return nil, clipboardFormatError("invalid local file URI")
		}
		if len(files) >= clipboardItemLimit {
			return nil, clipboardFormatError("too many files")
		}
		files = append(files, u.Path)
	}
	return files, nil
}
