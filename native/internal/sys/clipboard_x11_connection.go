package sys

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/dyike/keel/native"
	"github.com/jezek/xgb"
)

type clipboardDisplay struct {
	network, address, host, number string
	screen                         int
}

func parseClipboardDisplay(value string) (clipboardDisplay, error) {
	bad := fmt.Errorf("%w: invalid X11 DISPLAY", native.ErrUnsupported)
	i := strings.LastIndex(value, ":")
	if i < 0 {
		return clipboardDisplay{}, bad
	}
	prefix, number := value[:i], value[i+1:]
	screen := 0
	if n := strings.IndexByte(number, '.'); n >= 0 {
		var err error
		screen, err = strconv.Atoi(number[n+1:])
		if err != nil || screen < 0 {
			return clipboardDisplay{}, bad
		}
		number = number[:n]
	}
	n, err := strconv.Atoi(number)
	if err != nil || n < 0 || n > 59535 {
		return clipboardDisplay{}, bad
	}
	d := clipboardDisplay{number: strconv.Itoa(n), screen: screen}
	if strings.HasPrefix(prefix, "/") {
		d.network, d.address = "unix", prefix+":"+d.number
		return d, nil
	}
	protocol := "tcp"
	if i := strings.IndexByte(prefix, '/'); i >= 0 {
		protocol, prefix = prefix[:i], prefix[i+1:]
	}
	if protocol != "tcp" && protocol != "tcp4" && protocol != "tcp6" && protocol != "unix" {
		return clipboardDisplay{}, bad
	}
	if prefix == "" || prefix == "unix" || protocol == "unix" {
		d.network, d.address = "unix", "/tmp/.X11-unix/X"+d.number
		return d, nil
	}
	if protocol != "tcp" && protocol != "tcp4" && protocol != "tcp6" {
		return clipboardDisplay{}, bad
	}
	d.host = strings.Trim(prefix, "[]")
	d.network, d.address = protocol, net.JoinHostPort(d.host, strconv.Itoa(6000+n))
	return d, nil
}

func clipboardXCookie(data []byte, d clipboardDisplay, hostname string, remote net.IP) (string, error) {
	take := func() ([]byte, bool) {
		if len(data) < 2 {
			return nil, false
		}
		n := int(binary.BigEndian.Uint16(data))
		data = data[2:]
		if n > len(data) {
			return nil, false
		}
		v := data[:n]
		data = data[n:]
		return v, true
	}
	for len(data) > 0 {
		if len(data) < 2 {
			return "", clipboardFormatError("invalid Xauthority")
		}
		family := binary.BigEndian.Uint16(data)
		data = data[2:]
		fields := [4][]byte{}
		for i := range fields {
			var ok bool
			fields[i], ok = take()
			if !ok {
				return "", clipboardFormatError("truncated Xauthority")
			}
		}
		address, number, name, cookie := fields[0], fields[1], fields[2], fields[3]
		match := family == 65535 || family == 256 && (string(address) == hostname || string(address) == d.host)
		if family == 0 && len(address) == 4 || family == 6 && len(address) == 16 {
			match = remote != nil && remote.Equal(net.IP(address))
		}
		if match && (len(number) == 0 || string(number) == d.number) && string(name) == "MIT-MAGIC-COOKIE-1" && len(cookie) == 16 {
			return hex.EncodeToString(cookie), nil
		}
	}
	return "", nil
}

func openClipboardX11(value string, deadline time.Time) (*xgb.Conn, func(), error) {
	d, err := parseClipboardDisplay(value)
	if err != nil {
		return nil, nil, err
	}
	raw, err := net.DialTimeout(d.network, d.address, time.Until(deadline))
	if err != nil {
		return nil, nil, fmt.Errorf("%w: X11 connection: %v", native.ErrUnsupported, err)
	}
	fail := func(err error) (*xgb.Conn, func(), error) { raw.Close(); return nil, nil, err }
	if err = raw.SetDeadline(deadline); err != nil {
		return fail(err)
	}
	name := os.Getenv("XAUTHORITY")
	if name == "" {
		home, _ := os.UserHomeDir()
		name = filepath.Join(home, ".Xauthority")
	}
	cookie := ""
	if info, statErr := os.Stat(name); statErr == nil {
		if !info.Mode().IsRegular() || info.Size() > 1<<20 {
			return fail(clipboardFormatError("invalid Xauthority size or type"))
		}
		f, err := os.Open(name)
		if err != nil {
			return fail(err)
		}
		data, err := io.ReadAll(io.LimitReader(f, (1<<20)+1))
		f.Close()
		if err != nil {
			return fail(err)
		}
		if len(data) > 1<<20 {
			return fail(clipboardFormatError("Xauthority exceeds limit"))
		}
		hostname, _ := os.Hostname()
		var remote net.IP
		if addr, ok := raw.RemoteAddr().(*net.TCPAddr); ok {
			remote = addr.IP
		}
		cookie, err = clipboardXCookie(data, d, hostname, remote)
		if err != nil {
			return fail(err)
		}
	} else if !os.IsNotExist(statErr) {
		return fail(statErr)
	}
	var conn *xgb.Conn
	if cookie != "" {
		conn, err = xgb.NewConnNetWithCookieHex(raw, cookie)
	} else {
		conn, err = xgb.NewConnNet(raw)
	}
	if err != nil {
		return fail(fmt.Errorf("%w: X11 handshake: %v", native.ErrFailed, err))
	}
	conn.DefaultScreen = d.screen
	// Close the underlying socket first so pending replies cannot delay cleanup.
	return conn, func() { raw.Close(); conn.Close() }, nil
}
