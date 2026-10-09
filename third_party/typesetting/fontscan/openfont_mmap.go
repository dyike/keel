//go:build unix && !js && !wasip1

package fontscan

import (
	"os"
	"syscall"

	"github.com/go-text/typesetting/font"
	ot "github.com/go-text/typesetting/font/opentype"
)

// openFont maps a font file read-only. The mapping is never removed: the
// face parsed from it refers to it for the life of the process. release only
// closes a file that could not be mapped.
func openFont(path string) (font.Resource, func(), error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}
	st, err := f.Stat()
	if err == nil && st.Size() > 0 && st.Size() == int64(int(st.Size())) {
		data, err := syscall.Mmap(int(f.Fd()), 0, int(st.Size()), syscall.PROT_READ, syscall.MAP_SHARED)
		if err == nil {
			f.Close()
			return ot.NewShared(data), func() {}, nil
		}
	}
	return f, func() { f.Close() }, nil
}
