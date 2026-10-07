//go:build darwin && !ios && cgo

package sys

/*
#include <unistd.h>
#include <errno.h>
#include <string.h>
#include <libproc.h>
#include <sys/proc_info.h>
static int keel_process_foreground(int fd){return tcgetpgrp(fd);}
static int keel_process_cwd(int pid,char *out,int size){
 struct proc_vnodepathinfo info;
 if(proc_pidinfo(pid,PROC_PIDVNODEPATHINFO,0,&info,sizeof(info)) != sizeof(info))return 0;
 strlcpy(out,info.pvi_cdir.vip_path,size);return 1;
}
*/
import "C"
import (
	"fmt"
	"github.com/dyike/keel/native"
	"unsafe"
)

func ProcessForegroundPID(fd uintptr) (int, error) {
	pid, err := C.keel_process_foreground(C.int(fd))
	if pid < 0 {
		return 0, fmt.Errorf("%w: foreground process: %v", native.ErrFailed, err)
	}
	return int(pid), nil
}
func ProcessDirectory(pid int) (string, error) {
	var buf [4096]byte
	ok, err := C.keel_process_cwd(C.int(pid), (*C.char)(unsafe.Pointer(&buf[0])), C.int(len(buf)))
	if ok == 0 {
		return "", fmt.Errorf("%w: process directory: %v", native.ErrFailed, err)
	}
	return C.GoString((*C.char)(unsafe.Pointer(&buf[0]))), nil
}
