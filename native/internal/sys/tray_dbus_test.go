//go:build (linux && !android) || (darwin && !ios)

package sys

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"reflect"
	"testing"

	"github.com/godbus/dbus/v5"
)

func TestTrayDBusWireSignatures(t *testing.T) {
	for _, tc := range []struct {
		value     any
		signature string
	}{
		{[]trayPixmap{}, "a(iiay)"}, {trayTooltip{}, "(sa(iiay)ss)"},
		{trayLayout{}, "(ia{sv}av)"}, {[]trayItemProperties{}, "a(ia{sv})"}, {[]trayMenuEvent{}, "a(isvu)"},
	} {
		if got := dbus.SignatureOf(tc.value).String(); got != tc.signature {
			t.Fatalf("%T = %s, want %s", tc.value, got, tc.signature)
		}
	}
}

func TestTrayDBusPixmapStraightARGB(t *testing.T) {
	art := image.NewNRGBA(image.Rect(2, 3, 3, 4))
	art.SetNRGBA(2, 3, color.NRGBA{R: 0xe0, G: 0x60, B: 0x20, A: 0x80})
	var buf bytes.Buffer
	if err := png.Encode(&buf, art); err != nil {
		t.Fatal(err)
	}
	pixmaps, err := trayPixmaps(buf.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	if len(pixmaps) != 1 || pixmaps[0].Width != 1 || pixmaps[0].Height != 1 || !bytes.Equal(pixmaps[0].Pixels, []byte{0x80, 0xe0, 0x60, 0x20}) {
		t.Fatalf("pixmap = %+v", pixmaps)
	}
}

func TestTrayDBusMenuLayoutAndProperties(t *testing.T) {
	items := []TrayMenuItem{{ID: 1, Title: "Parent", Children: []TrayMenuItem{{ID: 2, Title: "Checked", Checked: true}, {ID: 3, Separator: true}}}, {ID: 4, Title: "Disabled", Disabled: true}}
	full, ok := trayMakeLayout(0, items, -1, nil)
	if !ok || len(full.Children) != 2 {
		t.Fatal("root layout incomplete")
	}
	parent := full.Children[0].Value().(trayLayout)
	if parent.ID != 1 || len(parent.Children) != 2 {
		t.Fatalf("parent = %+v", parent)
	}
	checked := parent.Children[0].Value().(trayLayout)
	if checked.Properties["toggle-state"].Value() != int32(1) {
		t.Fatal("checked state missing")
	}
	subtree, ok := trayMakeLayout(1, items, 1, []string{"label"})
	if !ok || len(subtree.Properties) != 1 || len(subtree.Children) != 2 {
		t.Fatal("subtree/filter not respected")
	}
	shallow, _ := trayMakeLayout(0, items, 0, nil)
	if len(shallow.Children) != 0 {
		t.Fatal("depth 0 expanded")
	}
	if _, ok := trayMakeLayout(99, items, -1, nil); ok {
		t.Fatal("unknown ID found")
	}
	disabled := trayMenuProps(items[1])
	if disabled["enabled"].Value() != false {
		t.Fatal("disabled property missing")
	}
	if signature := dbus.SignatureOf(full).String(); signature != "(ia{sv}av)" {
		t.Fatal(signature)
	}
}

func TestTrayDBusEvents(t *testing.T) {
	got := []uint32{}
	t1 := &dbusTray{options: TrayOptions{Menu: []TrayMenuItem{{ID: 1, Title: "Action"}, {ID: 2, Disabled: true}, {ID: 3, Separator: true}, {ID: 4, Children: []TrayMenuItem{{ID: 5}}}}, Event: func(id uint32) { got = append(got, id) }}}
	t1.activate()
	t1.menuEvent(1, "clicked")
	t1.menuEvent(1, "hovered")
	t1.menuEvent(2, "clicked")
	t1.menuEvent(3, "clicked")
	t1.menuEvent(4, "clicked")
	if t1.menuEvent(99, "clicked") {
		t.Fatal("unknown event accepted")
	}
	if !reflect.DeepEqual(got, []uint32{0, 1}) {
		t.Fatalf("events = %v", got)
	}
	t1.closed = true
	t1.activate()
	t1.menuEvent(1, "clicked")
	if len(got) != 2 {
		t.Fatal("closed tray emitted")
	}
}
