package main

import (
	"bufio"
	"net"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

// reloadControl serves the socket watched apps connect to (KEEL_RUN_CONTROL).
// Before a restart keel run asks the app over it to save its state
// (KEEL_RUN_STATE) and exit; ui/window restores that state when the new
// build opens its windows. Apps built with a Keel that does not connect are
// stopped as before.
type reloadControl struct {
	ln    net.Listener
	path  string
	mu    sync.Mutex
	conns map[int]net.Conn // by process ID
}

// newReloadControl listens in dir; nil if the platform refuses.
func newReloadControl(dir string) *reloadControl {
	path := filepath.Join(dir, "control.sock")
	ln, err := net.Listen("unix", path)
	if err != nil {
		return nil
	}
	r := &reloadControl{ln: ln, path: path, conns: map[int]net.Conn{}}
	go r.serve()
	return r
}

func (r *reloadControl) serve() {
	for {
		conn, err := r.ln.Accept()
		if err != nil {
			return
		}
		go func() {
			sc := bufio.NewScanner(conn)
			if !sc.Scan() {
				conn.Close()
				return
			}
			pid, err := strconv.Atoi(strings.TrimPrefix(sc.Text(), "hello "))
			if err != nil {
				conn.Close()
				return
			}
			r.mu.Lock()
			r.conns[pid] = conn
			r.mu.Unlock()
			for sc.Scan() { // the app sends nothing more; wait for it to go
			}
			r.mu.Lock()
			if r.conns[pid] == conn {
				delete(r.conns, pid)
			}
			r.mu.Unlock()
			conn.Close()
		}()
	}
}

// env is what a watched app needs to take part in reloads.
func (r *reloadControl) env(state string) []string {
	env := []string{"KEEL_RUN_STATE=" + state}
	if r != nil {
		env = append(env, "KEEL_RUN_CONTROL="+r.path)
	}
	return env
}

// ask requests a reload from the app with process ID pid; false if it is not
// connected.
func (r *reloadControl) ask(pid int) bool {
	if r == nil {
		return false
	}
	r.mu.Lock()
	conn := r.conns[pid]
	r.mu.Unlock()
	if conn == nil {
		return false
	}
	_, err := conn.Write([]byte("reload\n"))
	return err == nil
}

func (r *reloadControl) close() {
	if r != nil {
		r.ln.Close()
	}
}

// reloadRunProcess stops cmd for a restart: a connected app saves its state
// and runs its windows' OnClose first. It falls back to stopRunProcess.
func reloadRunProcess(cmd *exec.Cmd, done <-chan error, ctl *reloadControl) {
	if cmd.Process == nil || !ctl.ask(cmd.Process.Pid) {
		stopRunProcess(cmd, done)
		return
	}
	timer := time.NewTimer(5 * time.Second) // OnClose may persist files
	defer timer.Stop()
	select {
	case <-done:
	case <-timer.C:
		killRunProcess(cmd)
		<-done
	}
}
