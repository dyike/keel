package sys

import (
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net"
	"reflect"
	"strings"
	"testing"

	"github.com/jezek/xgb"
	"github.com/jezek/xgb/xproto"
)

func TestXClipboardFormats(t *testing.T) {
	for _, tc := range []struct {
		name    string
		formats map[string][]byte
		text    string
		files   []string
		mime    string
	}{
		{"unicode", map[string][]byte{"UTF8_STRING": []byte("中文🙂"), "STRING": []byte("wrong"), "image/png": {1, 2, 3}}, "中文🙂", nil, "image/png"},
		{"latin1", map[string][]byte{"STRING": {0x63, 0x61, 0x66, 0xe9}}, "café", nil, ""},
		{"files before preview", map[string][]byte{"text/uri-list": []byte("#comment\r\nfile:///tmp/a%20b\r\nfile://localhost/tmp/%E4%B8%AD\nhttps://example.com\nfile://remote/tmp/a"), "image/png": {1}}, "", []string{"/tmp/a b", "/tmp/中"}, ""},
		{"gnome cut references", map[string][]byte{"x-special/gnome-copied-files": []byte("cut\nfile:///tmp/a")}, "", []string{"/tmp/a"}, ""},
		{"empty preferred text", map[string][]byte{"UTF8_STRING": {}, "STRING": []byte("wrong")}, "", nil, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			targets := map[string]bool{}
			for k := range tc.formats {
				targets[k] = true
			}
			raw, err := collectXClipboard(targets, func(name string) ([]byte, error) {
				if !targets[name] {
					t.Fatalf("unadvertised %s", name)
				}
				if len(tc.files) > 0 && strings.HasPrefix(name, "image/") {
					t.Fatal("requested file preview")
				}
				return tc.formats[name], nil
			})
			if err != nil {
				t.Fatal(err)
			}
			var got clipboardSnapshot
			if err = json.Unmarshal(raw, &got); err != nil {
				t.Fatal(err)
			}
			if got.Text != tc.text || !reflect.DeepEqual(got.Files, tc.files) {
				t.Fatalf("unexpected snapshot: %+v", got)
			}
			if tc.mime != "" {
				if len(got.Images) != 1 || got.Images[0].MIME != tc.mime || !reflect.DeepEqual(got.Images[0].Data, tc.formats[tc.mime]) {
					t.Fatalf("image mismatch: %+v", got.Images)
				}
			} else if len(got.Images) != 0 {
				t.Fatal("unexpected image")
			}
		})
	}
	for _, data := range []string{"file:///tmp/%00", "file:///tmp/%ff", "file:///tmp/a?q=1", "file:relative", strings.Repeat("file:///a\n", 129)} {
		if _, err := clipboardURIs([]byte(data), false); err == nil {
			t.Fatalf("accepted invalid URI list %q", data)
		}
	}
	if _, err := clipboardURIs([]byte("move\nfile:///a"), true); err == nil {
		t.Fatal("accepted invalid GNOME operation")
	}
	for _, formats := range []map[string][]byte{{"UTF8_STRING": {0xff}}, {"UTF8_STRING": []byte("a"), "image/png": make([]byte, clipboardByteLimit)}, {"STRING": []byte(strings.Repeat("\xff", clipboardByteLimit/2+1))}} {
		raw, err := collectXClipboard(nil, func(name string) ([]byte, error) { return formats[name], nil })
		if err == nil || raw != nil {
			t.Fatal("accepted invalid/oversized snapshot")
		}
	}
	sentinel := errors.New("owner failed")
	raw, err := collectXClipboard(nil, func(string) ([]byte, error) { return nil, sentinel })
	if !errors.Is(err, sentinel) || raw != nil {
		t.Fatal("lost error")
	}
}

func TestXClipboardTransfer(t *testing.T) {
	notify := xproto.SelectionNotifyEvent{Requestor: 1, Selection: 2, Target: 3, Property: 4}
	chunkEvent := xproto.PropertyNotifyEvent{Window: 1, Atom: 4, State: xproto.PropertyNewValue}
	header := make([]byte, 4)
	xgb.Put32(header, 6)
	prop := func(kind xproto.Atom, format byte, data []byte) *xproto.GetPropertyReply {
		return &xproto.GetPropertyReply{Type: kind, Format: format, Value: data}
	}
	for _, tc := range []struct {
		name   string
		events []xgb.Event
		props  []*xproto.GetPropertyReply
		want   string
		acks   int
		fail   bool
	}{
		{"normal", []xgb.Event{chunkEvent, notify}, []*xproto.GetPropertyReply{prop(3, 8, []byte("中文"))}, "中文", 1, false},
		{"empty", []xgb.Event{notify}, []*xproto.GetPropertyReply{prop(3, 8, []byte{})}, "", 1, false},
		{"incremental", []xgb.Event{notify, chunkEvent, chunkEvent, chunkEvent}, []*xproto.GetPropertyReply{prop(5, 32, header), prop(3, 8, []byte("abc")), prop(3, 8, []byte("def")), prop(3, 8, nil)}, "abcdef", 4, false},
		{"changed format", []xgb.Event{notify, chunkEvent, chunkEvent}, []*xproto.GetPropertyReply{prop(5, 32, header), prop(3, 8, []byte("abc")), prop(3, 16, nil)}, "", 2, true},
		{"truncated header", []xgb.Event{notify}, []*xproto.GetPropertyReply{prop(5, 32, []byte{1})}, "", 0, true},
		{"disconnect", []xgb.Event{notify, chunkEvent}, []*xproto.GetPropertyReply{prop(5, 32, header), prop(3, 8, []byte("abc"))}, "", 2, true},
		{"wrong requestor", []xgb.Event{xproto.SelectionNotifyEvent{Requestor: 9, Selection: 2, Target: 3, Property: 4}}, nil, "", 0, true},
		{"partial property", []xgb.Event{notify}, []*xproto.GetPropertyReply{{Type: 3, Format: 8, BytesAfter: 1}}, "", 0, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ei, pi, acks := 0, 0, 0
			tr := clipboardXTransfer{window: 1, selection: 2, target: 3, property: 4, incr: 5,
				next: func() (xgb.Event, error) {
					if ei == len(tc.events) {
						return nil, nil
					}
					e := tc.events[ei]
					ei++
					return e, nil
				},
				fetch: func() (*xproto.GetPropertyReply, error) {
					if pi == len(tc.props) {
						t.Fatal("unexpected fetch")
					}
					p := tc.props[pi]
					pi++
					return p, nil
				},
				remove: func() error { acks++; return nil },
			}
			got, err := tr.receive()
			if (err != nil) != tc.fail || acks != tc.acks {
				t.Fatalf("err=%v acks=%d", err, acks)
			}
			if !tc.fail && (string(got.data) != tc.want || got.data == nil || got.kind != 3 || got.format != 8) {
				t.Fatalf("unexpected value %+v", got)
			}
			if tc.fail && got.data != nil {
				t.Fatal("returned partial data")
			}
		})
	}
}

func TestXClipboardConnectionParsing(t *testing.T) {
	for _, tc := range []struct {
		value, network, address string
		screen                  int
	}{
		{":0", "unix", "/tmp/.X11-unix/X0", 0}, {"unix:1.2", "unix", "/tmp/.X11-unix/X1", 2},
		{"host:10", "tcp", "host:6010", 0}, {"tcp6/[::1]:0", "tcp6", "[::1]:6000", 0},
		{"/tmp/launch/socket:0", "unix", "/tmp/launch/socket:0", 0},
	} {
		d, err := parseClipboardDisplay(tc.value)
		if err != nil || d.network != tc.network || d.address != tc.address || d.screen != tc.screen {
			t.Fatalf("%s: %+v %v", tc.value, d, err)
		}
	}
	for _, value := range []string{"", ":-1", ":1.-1", ":60000", "bogus/:0", "host:abc"} {
		if _, err := parseClipboardDisplay(value); err == nil {
			t.Fatalf("accepted %q", value)
		}
	}
	cookie := []byte("0123456789abcdef")
	record := func(family uint16, addr, display string) []byte {
		b := binary.BigEndian.AppendUint16(nil, family)
		for _, v := range [][]byte{[]byte(addr), []byte(display), []byte("MIT-MAGIC-COOKIE-1"), cookie} {
			b = binary.BigEndian.AppendUint16(b, uint16(len(v)))
			b = append(b, v...)
		}
		return b
	}
	d := clipboardDisplay{number: "1"}
	data := append(record(256, "host", "0"), record(256, "host", "1")...)
	got, err := clipboardXCookie(data, d, "host", nil)
	if err != nil || got != hex.EncodeToString(cookie) {
		t.Fatal("local cookie lookup failed", err)
	}
	got, err = clipboardXCookie(record(0, string([]byte{127, 0, 0, 1}), "1"), d, "host", net.ParseIP("127.0.0.1"))
	if err != nil || got == "" {
		t.Fatal("IP cookie lookup failed")
	}
	got, err = clipboardXCookie(data, d, "other", nil)
	if err != nil || got != "" {
		t.Fatal("wrong host accepted")
	}
	if _, err = clipboardXCookie(data[:len(data)-1], d, "host", nil); err == nil {
		t.Fatal("truncated authority accepted")
	}
}

func TestXClipboardTransferFailures(t *testing.T) {
	sentinel := errors.New("transport failed")
	notify := xproto.SelectionNotifyEvent{Requestor: 1, Selection: 2, Target: 3, Property: 4}
	for _, phase := range []string{"next", "fetch", "remove", "refused"} {
		t.Run(phase, func(t *testing.T) {
			tr := clipboardXTransfer{window: 1, selection: 2, target: 3, property: 4, incr: 5,
				next: func() (xgb.Event, error) {
					if phase == "next" {
						return nil, sentinel
					}
					e := notify
					if phase == "refused" {
						e.Property = 0
					}
					return e, nil
				},
				fetch: func() (*xproto.GetPropertyReply, error) {
					if phase == "refused" {
						t.Fatal("fetched refused selection")
					}
					if phase == "fetch" {
						return nil, sentinel
					}
					return &xproto.GetPropertyReply{Type: 3, Format: 8, Value: []byte("abc")}, nil
				},
				remove: func() error {
					if phase == "remove" {
						return sentinel
					}
					return nil
				},
			}
			got, err := tr.receive()
			if got.data != nil {
				t.Fatal("partial result")
			}
			if phase == "refused" {
				if err != nil {
					t.Fatal(err)
				}
			} else if !errors.Is(err, sentinel) {
				t.Fatalf("lost error: %v", err)
			}
		})
	}
	for _, announced := range []uint32{clipboardByteLimit, clipboardByteLimit + 1} {
		header := make([]byte, 4)
		xgb.Put32(header, announced)
		step, acks := 0, 0
		tr := clipboardXTransfer{window: 1, selection: 2, target: 3, property: 4, incr: 5,
			next: func() (xgb.Event, error) {
				if step == 0 {
					return notify, nil
				}
				return xproto.PropertyNotifyEvent{Window: 1, Atom: 4, State: xproto.PropertyNewValue}, nil
			},
			fetch: func() (*xproto.GetPropertyReply, error) {
				step++
				p := &xproto.GetPropertyReply{Type: 3, Format: 8}
				switch step {
				case 1:
					p.Type = 5
					p.Format = 32
					p.Value = header
				case 2:
					p.Value = make([]byte, clipboardByteLimit)
				case 3:
					p.Value = []byte{1}
				default:
					t.Fatal("read past overflow")
				}
				return p, nil
			},
			remove: func() error { acks++; return nil },
		}
		got, err := tr.receive()
		if err == nil || got.data != nil {
			t.Fatal("accepted oversized transfer")
		}
		if announced > clipboardByteLimit && acks != 0 {
			t.Fatal("acknowledged oversized header")
		}
	}
}

func FuzzClipboardURIs(f *testing.F) {
	f.Add([]byte("file:///tmp/a%20b\r\n"), false)
	f.Add([]byte("copy\nfile:///tmp/中"), true)
	f.Fuzz(func(t *testing.T, data []byte, gnome bool) {
		files, err := clipboardURIs(data, gnome)
		if err == nil {
			if len(files) > clipboardItemLimit {
				t.Fatal("file count overflow")
			}
			for _, p := range files {
				if !strings.HasPrefix(p, "/") || strings.ContainsRune(p, 0) {
					t.Fatal("invalid path")
				}
			}
		}
	})
}
