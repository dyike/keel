package fontscan

import (
	"reflect"
	"testing"
)

func TestShareSystemFontCoverage(t *testing.T) {
	a := RuneSet{{ref: 0, set: pageSet{1, 2, 3}}, {ref: 1, set: pageSet{4}}}
	b := append(RuneSet(nil), a...)
	c := append(RuneSet(nil), a...)
	c[1].set[0] = 5
	index := systemFontsIndex{{footprints: []Footprint{{Runes: a}, {Runes: c}}}, {footprints: []Footprint{{Runes: b}}}}
	index.shareRuneSets()
	shared := index[1].footprints[0].Runes
	if &shared[0] != &a[0] {
		t.Fatal("identical coverage not shared")
	}
	if !reflect.DeepEqual(shared, b) || !reflect.DeepEqual(index[0].footprints[1].Runes, c) {
		t.Fatal("coverage changed")
	}
	if &index[0].footprints[1].Runes[0] == &a[0] {
		t.Fatal("different coverage shared")
	}
}
