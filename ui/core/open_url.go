package core

import (
	"fmt"
	"net/url"
	"strings"
)

// OpenURL opens an absolute HTTP, HTTPS or mailto URL using the platform handler.
// It reports launch errors, not whether the destination subsequently loaded.
func OpenURL(raw string) error {
	u, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("open URL: %w", err)
	}
	switch strings.ToLower(u.Scheme) {
	case "http", "https":
		if u.Host == "" {
			return fmt.Errorf("open URL: missing host")
		}
	case "mailto":
		if u.Opaque == "" && u.Path == "" {
			return fmt.Errorf("open URL: missing address")
		}
	default:
		return fmt.Errorf("open URL: unsupported scheme %q", u.Scheme)
	}
	return openURL(raw)
}
