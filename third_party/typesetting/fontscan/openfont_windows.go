//go:build windows

package fontscan

import (
	"os"
	"unsafe"

	"golang.org/x/sys/windows"

	"github.com/dyike/keel/third_party/typesetting/font"
	ot "github.com/dyike/keel/third_party/typesetting/font/opentype"
)

// openFont maps a font file read-only. The view is never unmapped: the face
// parsed from it refers to it for the life of the process. release only
// closes a file that could not be mapped.
func openFont(path string) (font.Resource, func(), error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}
	st, err := f.Stat()
	if err == nil && st.Size() > 0 && st.Size() == int64(int(st.Size())) {
		m, err := windows.CreateFileMapping(windows.Handle(f.Fd()), nil, windows.PAGE_READONLY, 0, 0, nil)
		if err == nil {
			addr, err := windows.MapViewOfFile(m, windows.FILE_MAP_READ, 0, 0, uintptr(st.Size()))
			windows.CloseHandle(m) // the view keeps the mapping
			if err == nil {
				f.Close()
				// addr is outside the Go heap; convert it without a uintptr-to-pointer cast vet flags.
				data := unsafe.Slice((*byte)(*(*unsafe.Pointer)(unsafe.Pointer(&addr))), int(st.Size()))
				return ot.NewShared(data), func() {}, nil
			}
		}
	}
	return f, func() { f.Close() }, nil
}
