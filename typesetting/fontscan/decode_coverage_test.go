package fontscan

import (
	"bytes"
	"hash/maphash"
	"reflect"
	"testing"

	"github.com/dyike/keel/typesetting/font"
)

func TestIndexDecodeSharesCoverageNotIdentities(t *testing.T) {
	runes := RuneSet{{ref: 0, set: pageSet{1, 2}}, {ref: 3, set: pageSet{7}}}
	index := systemFontsIndex{
		{path: "a.ttc", footprints: []Footprint{
			{Location: Location{File: "a.ttc", Index: 2, Instance: 1}, Family: "regular", Runes: runes, Aspect: font.Aspect{Weight: 400}},
			{Location: Location{File: "a.ttc", Index: 2, Instance: 3}, Family: "bold", Runes: runes, Aspect: font.Aspect{Weight: 700}},
		}},
		{path: "b.ttf", footprints: []Footprint{{Location: Location{File: "b.ttf"}, Runes: runes}}},
	}
	var buf bytes.Buffer
	if err := index.serializeTo(&buf); err != nil {
		t.Fatal(err)
	}
	decode := func() systemFontsIndex {
		t.Helper()
		got, err := deserializeIndex(bytes.NewReader(buf.Bytes()))
		if err != nil {
			t.Fatal(err)
		}
		return got
	}
	got := decode()
	first := &got[0].footprints[0].Runes[0]
	for i, file := range got {
		for j, fp := range file.footprints {
			want := index[i].footprints[j]
			if !reflect.DeepEqual(fp.Runes, want.Runes) || &fp.Runes[0] != first {
				t.Fatal("coverage not shared at decode")
			}
			if fp.Location != want.Location || fp.Aspect != want.Aspect || fp.Family != want.Family {
				t.Fatal("font identity changed")
			}
		}
	}
	other := decode()
	if &other[0].footprints[0].Runes[0] == first {
		t.Fatal("independent index reads share mutable decoder state")
	}
	got[0].footprints[0].Family = "changed"
	if got[0].footprints[1].Family != "bold" {
		t.Fatal("face metadata aliased")
	}
}

func TestRuneSetDecodeOwnershipCollisionAndTruncation(t *testing.T) {
	want := RuneSet{{ref: 1, set: pageSet{1, 2, 3}}, {ref: 2, set: pageSet{4}}}
	d := runeSetDecoder{seed: maphash.MakeSeed(), sets: make(map[uint64]RuneSet)}
	encoded := want.serialize()
	var first RuneSet
	if _, err := d.decode(&first, encoded); err != nil {
		t.Fatal(err)
	}
	clear(encoded)
	if !reflect.DeepEqual(first, want) {
		t.Fatal("coverage borrowed the record buffer")
	}
	encoded = want.serialize()
	var again RuneSet
	if got := testing.AllocsPerRun(20, func() {
		if _, err := d.decode(&again, encoded); err != nil {
			t.Fatal(err)
		}
	}); got != 0 {
		t.Fatalf("duplicate coverage allocated %.0f times", got)
	}
	if &first[0] != &again[0] {
		t.Fatal("duplicate coverage copied")
	}
	// Force different content, reference, and length into the candidate bucket.
	differentSets := []RuneSet{
		{{ref: 9, set: want[0].set}, want[1]},
		{want[0]},
	}
	for word := range 8 {
		changed := append(RuneSet(nil), want...)
		changed[0].set[word] ^= 8
		differentSets = append(differentSets, changed)
	}
	for _, different := range differentSets {
		raw := different.serialize()
		key := maphash.Bytes(d.seed, raw)
		d.sets[key] = first
		var got RuneSet
		if _, err := d.decode(&got, raw); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(got, different) || &got[0] == &first[0] {
			t.Fatal("hash collision reused different coverage")
		}
		var repeated RuneSet
		if _, err := d.decode(&repeated, raw); err != nil || &repeated[0] != &got[0] {
			t.Fatal("collision entry not reused", err)
		}
	}
	for n := range len(encoded) {
		var got RuneSet
		if _, err := d.decode(&got, encoded[:n]); err == nil {
			t.Fatalf("accepted truncation at %d", n)
		}
	}
	var empty RuneSet
	if n, err := d.decode(&empty, []byte{0, 0}); err != nil || n != 2 || empty == nil || len(empty) != 0 {
		t.Fatal("empty coverage changed")
	}
}
