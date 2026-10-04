package core

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	_ "golang.org/x/image/webp"
)

const maxImageBytes = 16 << 20

// MaxImagePixels is the largest image, in pixels, that decoding accepts.
const MaxImagePixels = 32_000_000

// DecodeImage reads PNG, JPEG, GIF (first frame) and WebP. Size limits apply to
// encoded input and decoded dimensions, before allocating a pixel buffer.
// It blocks until completion; call from a worker with a deadline, not during layout.
func DecodeImage(ctx context.Context, source string) (image.Image, error) {
	data, err := ReadImageSource(ctx, source)
	if err != nil {
		return nil, err
	}
	return DecodeImageBytes(data)
}

// ReadImageSource returns the encoded bytes of an image source, as
// DecodeImage reads it: a local path, a file URL, an HTTP(S) URL or a data
// URL, up to 16MiB. Use it to decode formats DecodeImage does not, such as
// SVG or every frame of a GIF. It blocks; call it from a worker.
func ReadImageSource(ctx context.Context, source string) ([]byte, error) {
	var reader io.ReadCloser
	switch {
	case strings.HasPrefix(source, "data:"):
		if len(source) > maxImageBytes*2 {
			return nil, fmt.Errorf("image data URL too large")
		}
		header, data, ok := strings.Cut(source, ",")
		if !ok {
			return nil, fmt.Errorf("invalid image data URL")
		}
		var raw []byte
		var err error
		if strings.HasSuffix(header, ";base64") {
			raw, err = base64.StdEncoding.DecodeString(data)
		} else {
			var value string
			value, err = url.PathUnescape(data)
			raw = []byte(value)
		}
		if err != nil {
			return nil, err
		}
		reader = io.NopCloser(bytes.NewReader(raw))
	case strings.HasPrefix(source, "https://") || strings.HasPrefix(source, "http://"):
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, source, nil)
		if err != nil {
			return nil, err
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, err
		}
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			resp.Body.Close()
			return nil, fmt.Errorf("image HTTP status %d", resp.StatusCode)
		}
		reader = resp.Body
	default:
		path := source
		if decoded, err := url.PathUnescape(source); err == nil {
			path = decoded
		}
		if strings.HasPrefix(source, "file:") {
			u, err := url.Parse(source)
			if err != nil {
				return nil, err
			}
			if u.Host != "" && u.Host != "localhost" {
				return nil, fmt.Errorf("nonlocal file URL")
			}
			path = u.Path
		} else if strings.Contains(source, "://") {
			return nil, fmt.Errorf("unsupported image URL")
		}
		f, err := os.Open(path)
		if err != nil {
			return nil, err
		}
		reader = f
	}
	defer reader.Close()
	data, err := io.ReadAll(io.LimitReader(reader, maxImageBytes+1))
	if err != nil {
		return nil, err
	}
	if len(data) > maxImageBytes {
		return nil, fmt.Errorf("image exceeds %d bytes", maxImageBytes)
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return data, nil
}

// DecodeImageBytes decodes PNG, JPEG, GIF (first frame) or WebP bytes,
// refusing images larger than DecodeImage allows before allocating them.
func DecodeImageBytes(data []byte) (image.Image, error) {
	if len(data) > maxImageBytes {
		return nil, fmt.Errorf("image exceeds %d bytes", maxImageBytes)
	}
	cfg, _, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	if cfg.Width <= 0 || cfg.Height <= 0 || int64(cfg.Width)*int64(cfg.Height) > MaxImagePixels {
		return nil, fmt.Errorf("image dimensions exceed limit")
	}
	img, _, err := image.Decode(bytes.NewReader(data))
	return img, err
}
