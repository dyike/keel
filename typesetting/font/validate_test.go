package font

import (
	"bytes"
	"encoding/binary"
	ot "github.com/dyike/keel/typesetting/font/opentype"
	"golang.org/x/image/font/gofont/goregular"
	"testing"
)

func TestValidateRequiredTablesMatchesNewFont(t *testing.T) {
	for _, tag := range []string{"", "head", "maxp", "cmap"} {
		data := bytes.Clone(goregular.TTF)
		if tag != "" {
			n := int(binary.BigEndian.Uint16(data[4:6]))
			found := false
			for i := 0; i < n; i++ {
				r := 12 + i*16
				if string(data[r:r+4]) == tag {
					off := int(binary.BigEndian.Uint32(data[r+8 : r+12]))
					size := int(binary.BigEndian.Uint32(data[r+12 : r+16]))
					clear(data[off : off+size])
					found = true
					break
				}
			}
			if !found {
				t.Fatal("test font lacks required table")
			}
		}
		ld, err := ot.NewLoader(bytes.NewReader(data))
		if err != nil {
			t.Fatal(err)
		}
		validErr := Validate(ld)
		_, fullErr := NewFont(ld)
		if (validErr == nil) != (fullErr == nil) {
			t.Fatalf("%s validation=%v full parse=%v", tag, validErr, fullErr)
		}
		if tag == "" && validErr != nil {
			t.Fatal(validErr)
		}
	}
}
