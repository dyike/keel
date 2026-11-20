// SPDX-License-Identifier: Unlicense OR MIT

package app

import (
	"bufio"
	"io"
	"log"
	"runtime"
	"strings"
	"sync"
	"unsafe"

	_ "github.com/dyike/keel/gio/internal/cocoainit"
	"github.com/ebitengine/purego"
	"github.com/ebitengine/purego/objc"
)

var (
	iosLogOnce     sync.Once
	iosNSLog       uintptr
	iosPoolPush    uintptr
	iosPoolPop     uintptr
	iosMsgSend     uintptr
	iosNSStringCls uintptr
	iosSelUTF8     uintptr
)

func initIOSLog() {
	iosLogOnce.Do(func() {
		foundation, err := purego.Dlopen("/System/Library/Frameworks/Foundation.framework/Foundation", purego.RTLD_NOW|purego.RTLD_GLOBAL)
		if err != nil {
			return
		}
		iosNSLog, _ = purego.Dlsym(foundation, "NSLog")
		iosPoolPush, _ = purego.Dlsym(purego.RTLD_DEFAULT, "objc_autoreleasePoolPush")
		iosPoolPop, _ = purego.Dlsym(purego.RTLD_DEFAULT, "objc_autoreleasePoolPop")
		iosMsgSend, _ = purego.Dlsym(purego.RTLD_DEFAULT, "objc_msgSend")
		iosNSStringCls = uintptr(objc.GetClass("NSString"))
		iosSelUTF8 = uintptr(objc.RegisterName("stringWithUTF8String:"))
	})
}

func nslogLine(line []byte) {
	initIOSLog()
	if iosNSLog == 0 || iosMsgSend == 0 || iosNSStringCls == 0 {
		return
	}
	// Escape '%' so NSLog receives a format string with zero variadic arguments,
	// avoiding platform-specific varargs ABI.
	escaped := strings.ReplaceAll(string(line), "%", "%%")
	buf := make([]byte, len(escaped)+1)
	copy(buf, escaped)
	var pool uintptr
	if iosPoolPush != 0 {
		pool, _, _ = purego.SyscallN(iosPoolPush)
	}
	str := iosLogCall(iosMsgSend, iosNSStringCls, iosSelUTF8, uintptr(unsafe.Pointer(unsafe.SliceData(buf))))
	if str != 0 {
		purego.SyscallN(iosNSLog, str)
	}
	if iosPoolPop != 0 {
		purego.SyscallN(iosPoolPop, pool)
	}
}

//go:uintptrescapes
func iosLogCall(fn, a, b, c uintptr) uintptr {
	r, _, _ := purego.SyscallN(fn, a, b, c)
	return r
}

func init() {
	// macOS Console already includes timestamps.
	log.SetFlags(log.Flags() &^ log.LstdFlags)
	log.SetOutput(newNSLogWriter())
}

func newNSLogWriter() io.Writer {
	r, w := io.Pipe()
	go func() {
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()
		// 1024 is an arbitrary truncation limit, taken from Android's
		// log buffer size.
		lineBuf := bufio.NewReaderSize(r, 1024)
		for {
			line, _, err := lineBuf.ReadLine()
			if err != nil {
				break
			}
			nslogLine(line)
		}
	}()
	return w
}
