// SPDX-License-Identifier: Unlicense OR MIT
package opentype

import (
	"bytes"
	"golang.org/x/image/font/gofont/goregular"
	"testing"
)

func TestNativeSourceOwnership(t *testing.T) {
	for _, collection := range []bool{false, true} {
		src := bytes.Clone(goregular.TTF)
		var face Face
		if collection {
			faces, err := ParseCollection(src)
			if err != nil {
				t.Fatal(err)
			}
			face = faces[0].Face.(Face)
		} else {
			var err error
			face, err = Parse(src)
			if err != nil {
				t.Fatal(err)
			}
		}
		data, index := face.FontData()
		if index != 0 || !bytes.Equal(data, src) {
			t.Fatal("native source lost its face identity")
		}
		clear(src)
		if !bytes.Equal(data, goregular.TTF) {
			t.Fatal("Parse did not own its bytes")
		}
		if _, ok := face.Face().NominalGlyph('A'); !ok {
			t.Fatal("owned tables were corrupted")
		}
	}
	src := bytes.Clone(goregular.TTF)
	faces, err := ParseCollectionShared(src)
	if err != nil {
		t.Fatal(err)
	}
	data, _ := faces[0].Face.(Face).FontData()
	if &data[0] != &src[0] {
		t.Fatal("Shared copied immutable font bytes")
	}
}
