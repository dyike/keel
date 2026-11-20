// SPDX-License-Identifier: Unlicense OR MIT

package app

import (
	"bufio"
	"log"
	"os"
	"runtime"
	"sync"
	"syscall"
	"unsafe"

	"github.com/ebitengine/purego"
)

// 1024 is the truncation limit from android/log.h, plus a \n.
const (
	logLineLimit   = 1024
	androidLogInfo = 4 // ANDROID_LOG_INFO
)

var (
	androidLogOnce sync.Once
	androidLogSym  uintptr
	logTag         = append([]byte(ID), 0)
)

func androidLogWrite(tag, msg []byte) {
	androidLogOnce.Do(func() {
		lib, err := purego.Dlopen("liblog.so", purego.RTLD_NOW|purego.RTLD_LOCAL)
		if err != nil {
			return
		}
		androidLogSym, _ = purego.Dlsym(lib, "__android_log_write")
	})
	if androidLogSym == 0 {
		return
	}
	androidLogSyscall(androidLogSym, androidLogInfo, uintptr(unsafe.Pointer(unsafe.SliceData(tag))), uintptr(unsafe.Pointer(unsafe.SliceData(msg))))
}

//go:uintptrescapes
func androidLogSyscall(fn, prio, tag, msg uintptr) {
	purego.SyscallN(fn, prio, tag, msg)
}

func init() {
	// Android's logcat already includes timestamps.
	log.SetFlags(log.Flags() &^ log.LstdFlags)
	log.SetOutput(new(androidLogWriter))

	// Redirect stdout and stderr to the Android logger.
	logFd(os.Stdout.Fd())
	logFd(os.Stderr.Fd())
}

type androidLogWriter struct {
	// buf has room for the maximum log line, plus a terminating '\0'.
	buf [logLineLimit + 1]byte
}

func (w *androidLogWriter) Write(data []byte) (int, error) {
	n := 0
	for len(data) > 0 {
		msg := data
		// Truncate the buffer, leaving space for the '\0'.
		if max := len(w.buf) - 1; len(msg) > max {
			msg = msg[:max]
		}
		buf := w.buf[:len(msg)+1]
		copy(buf, msg)
		// Terminating '\0'.
		buf[len(msg)] = 0
		androidLogWrite(logTag, buf)
		n += len(msg)
		data = data[len(msg):]
	}
	return n, nil
}

func logFd(fd uintptr) {
	r, w, err := os.Pipe()
	if err != nil {
		panic(err)
	}
	if err := syscall.Dup3(int(w.Fd()), int(fd), syscall.O_CLOEXEC); err != nil {
		panic(err)
	}
	go func() {
		lineBuf := bufio.NewReaderSize(r, logLineLimit)
		// The buffer to pass to C, including the terminating '\0'.
		buf := make([]byte, lineBuf.Size()+1)
		for {
			line, _, err := lineBuf.ReadLine()
			if err != nil {
				break
			}
			copy(buf, line)
			buf[len(line)] = 0
			androidLogWrite(logTag, buf[:len(line)+1])
		}
		// The garbage collector doesn't know that w's fd was dup'ed.
		// Avoid finalizing w, and thereby avoid its finalizer closing its fd.
		runtime.KeepAlive(w)
	}()
}
