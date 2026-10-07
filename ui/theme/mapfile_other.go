//go:build !unix

package theme

import "os"

// mapFile reads the file at path.
func mapFile(path string) ([]byte, error) { return os.ReadFile(path) }
