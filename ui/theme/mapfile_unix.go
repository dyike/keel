//go:build unix

package theme

import (
	"os"
	"syscall"
)

// mapFile maps the file at path read-only, for good, or reads it when it
// cannot be mapped.
func mapFile(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	st, err := f.Stat()
	if err != nil {
		return nil, err
	}
	if size := st.Size(); size > 0 && size == int64(int(size)) {
		if b, err := syscall.Mmap(int(f.Fd()), 0, int(size), syscall.PROT_READ, syscall.MAP_SHARED); err == nil {
			return b, nil
		}
	}
	return os.ReadFile(path)
}
