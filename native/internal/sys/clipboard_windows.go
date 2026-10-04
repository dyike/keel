//go:build windows

package sys

import (
	"fmt"
	"runtime"
	"sync"
	"time"
	"unsafe"

	"github.com/dyike/keel/native"
	"golang.org/x/sys/windows"
)

var (
	clipboardMu        sync.Mutex
	clipboardOpen      = user32.NewProc("OpenClipboard")
	clipboardClose     = user32.NewProc("CloseClipboard")
	clipboardAvailable = user32.NewProc("IsClipboardFormatAvailable")
	clipboardGet       = user32.NewProc("GetClipboardData")
	clipboardRegister  = user32.NewProc("RegisterClipboardFormatW")
	clipboardKernel    = windows.NewLazySystemDLL("kernel32.dll")
	clipboardLock      = clipboardKernel.NewProc("GlobalLock")
	clipboardUnlock    = clipboardKernel.NewProc("GlobalUnlock")
	clipboardSize      = clipboardKernel.NewProc("GlobalSize")
	clipboardFiles     = windows.NewLazySystemDLL("shell32.dll").NewProc("DragQueryFileW")
)

func ClipboardRead(done func([]byte, error)) {
	go func() {
		raw, err := readWindowsClipboard()
		done(raw, err)
	}()
}

func readWindowsClipboard() ([]byte, error) {
	clipboardMu.Lock()
	defer clipboardMu.Unlock()
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	var last error
	opened := false
	for attempt := 0; attempt < 8; attempt++ {
		r, _, err := clipboardOpen.Call(0)
		if r != 0 {
			opened = true
			break
		}
		last = err
		if attempt < 7 {
			time.Sleep(15 * time.Millisecond)
		}
	}
	if !opened {
		return nil, fmt.Errorf("%w: OpenClipboard: %v", native.ErrFailed, last)
	}
	defer clipboardClose.Call()
	pngName, _ := windows.UTF16PtrFromString("PNG")
	pngFormat, _, err := clipboardRegister.Call(uintptr(unsafe.Pointer(pngName)))
	if pngFormat == 0 {
		return nil, fmt.Errorf("%w: RegisterClipboardFormat: %v", native.ErrFailed, err)
	}
	return collectWindowsClipboard(func(name string) ([]byte, error) {
		format := map[string]uintptr{"text": 13, "dib": 8, "png": pngFormat}[name]
		handle, err := clipboardHandle(format)
		if err != nil || handle == 0 {
			return nil, err
		}
		size, _, err := clipboardSize.Call(handle)
		if size == 0 || size > clipboardByteLimit {
			return nil, fmt.Errorf("%w: clipboard allocation size: %v", native.ErrFailed, err)
		}
		address, _, err := clipboardLock.Call(handle)
		if address == 0 {
			return nil, fmt.Errorf("%w: GlobalLock: %v", native.ErrFailed, err)
		}
		defer clipboardUnlock.Call(handle)
		// Copy the borrowed address through Win32, keeping it out of Go's
		// pointer graph and reporting invalid native memory as an error.
		data := make([]byte, int(size))
		var copied uintptr
		if err := windows.ReadProcessMemory(windows.CurrentProcess(), address, &data[0], size, &copied); err != nil || copied != size {
			return nil, fmt.Errorf("%w: clipboard memory copy: %v", native.ErrFailed, err)
		}
		return data, nil
	}, func() ([]string, error) {
		handle, err := clipboardHandle(15)
		if err != nil || handle == 0 {
			return nil, err
		}
		count, _, _ := clipboardFiles.Call(handle, 0xffffffff, 0, 0)
		if count > clipboardItemLimit {
			return nil, clipboardFormatError("too many files")
		}
		out := make([]string, 0, int(count))
		total := 0
		for i := uintptr(0); i < count; i++ {
			size, _, _ := clipboardFiles.Call(handle, i, 0, 0)
			if size == 0 || size >= clipboardByteLimit/2 {
				return nil, clipboardFormatError("invalid file path size")
			}
			buffer := make([]uint16, int(size)+1)
			copied, _, _ := clipboardFiles.Call(handle, i, uintptr(unsafe.Pointer(&buffer[0])), size+1)
			if copied != size {
				return nil, clipboardFormatError("file path changed while reading")
			}
			path := windows.UTF16ToString(buffer)
			total += len(path)
			if total > clipboardByteLimit {
				return nil, clipboardFormatError("file paths exceed 16MiB")
			}
			out = append(out, path)
		}
		return out, nil
	})
}

func clipboardHandle(format uintptr) (uintptr, error) {
	if ok, _, _ := clipboardAvailable.Call(format); ok == 0 {
		return 0, nil
	}
	h, _, err := clipboardGet.Call(format)
	if h == 0 {
		return 0, fmt.Errorf("%w: GetClipboardData: %v", native.ErrFailed, err)
	}
	return h, nil
}
