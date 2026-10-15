package fontscan

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"io"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"golang.org/x/image/font/gofont/goregular"
)

func TestSameFontFiles(t *testing.T) {
	base := systemFontsIndex{{path: "a.ttf", modTime: 1}, {path: "b.ttc", modTime: 2}}
	if !sameFontFiles(base, append(systemFontsIndex(nil), base...)) {
		t.Fatal("unchanged index requires rewrite")
	}
	for _, changed := range []systemFontsIndex{
		nil, base[:1], append(append(systemFontsIndex(nil), base...), fileFootprints{path: "c.ttf"}),
		{{path: "a.ttf", modTime: 2}, base[1]},
		{{path: "renamed.ttf", modTime: 1}, base[1]},
		{base[1], base[0]},
	} {
		if sameFontFiles(base, changed) {
			t.Fatalf("changed index not rewritten: %v", changed)
		}
	}
}

func TestFontIndexStreamingFormat(t *testing.T) {
	index := systemFontsIndex{
		{path: "regular.ttf", modTime: 1, footprints: []Footprint{{Family: "regular", Runes: RuneSet{{ref: 2, set: pageSet{3}}}, Scripts: ScriptSet{}}}},
		{path: "empty.ttf", modTime: 2},
	}
	// Construct the old whole-index encoding; streamed writes must retain
	// exactly that uncompressed format and must read old cache files.
	var old bytes.Buffer
	var head [6]byte
	binary.BigEndian.PutUint16(head[:], cacheFormatVersion)
	binary.BigEndian.PutUint32(head[2:], uint32(len(index)))
	old.Write(head[:])
	for _, file := range index {
		raw := file.serializeTo(nil)
		binary.BigEndian.PutUint32(head[:4], uint32(len(raw)))
		old.Write(head[:4])
		old.Write(raw)
	}
	var encoded bytes.Buffer
	if err := index.serializeTo(&encoded); err != nil {
		t.Fatal(err)
	}
	r, err := gzip.NewReader(bytes.NewReader(encoded.Bytes()))
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := io.ReadAll(r)
	r.Close()
	if err != nil || !bytes.Equal(decoded, old.Bytes()) {
		t.Fatalf("cache format changed: %v", err)
	}
	got, err := deserializeIndex(bytes.NewReader(encoded.Bytes()))
	if err != nil || !reflect.DeepEqual(got, index) {
		t.Fatalf("roundtrip: %v: %#v", err, got)
	}
}

func TestIndexRecordReusesStorageAndRejectsTruncation(t *testing.T) {
	src := bytes.Repeat([]byte{42}, 100000)
	buffer, err := readIndexRecord(bytes.NewReader(src), uint32(len(src)), nil)
	if err != nil || !bytes.Equal(buffer, src) {
		t.Fatal("record differs:", err)
	}
	first := &buffer[0]
	buffer, err = readIndexRecord(bytes.NewReader(src), uint32(len(src)), buffer)
	if err != nil || &buffer[0] != first {
		t.Fatal("storage not reused:", err)
	}
	if _, err := readIndexRecord(bytes.NewReader(nil), ^uint32(0), nil); err == nil {
		t.Fatal("truncated oversized record accepted")
	}
}

func BenchmarkFontIndexRecordRead(b *testing.B) {
	data := bytes.Repeat([]byte{42}, 3143610)
	for _, retained := range []bool{false, true} {
		name := "buffer"
		if retained {
			name = "record"
		}
		b.Run(name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				reader := bytes.NewReader(data)
				if retained {
					if _, err := readIndexRecord(reader, uint32(len(data)), nil); err != nil {
						b.Fatal(err)
					}
				} else {
					var buffer bytes.Buffer
					if _, err := io.CopyN(&buffer, reader, int64(len(data))); err != nil {
						b.Fatal(err)
					}
				}
			}
		})
	}
}

func BenchmarkFontIndexSerialize(b *testing.B) {
	runes := make(RuneSet, 512)
	index := make(systemFontsIndex, 64)
	for i := range index {
		index[i] = fileFootprints{path: "font.ttf", footprints: []Footprint{
			{Family: "regular", Runes: runes}, {Family: "bold", Runes: runes},
			{Family: "light", Runes: runes}, {Family: "medium", Runes: runes},
		}}
	}
	for _, streamed := range []bool{false, true} {
		name := "whole"
		if streamed {
			name = "streamed"
		}
		b.Run(name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				if streamed {
					if err := index.serializeTo(io.Discard); err != nil {
						b.Fatal(err)
					}
				} else {
					buffer := make([]byte, 6, 4+len(index)*(aspectSize+1+2))
					binary.BigEndian.PutUint16(buffer, cacheFormatVersion)
					binary.BigEndian.PutUint32(buffer[2:], uint32(len(index)))
					for _, ff := range index {
						n := len(buffer)
						buffer = append(buffer, 0, 0, 0, 0)
						buffer = ff.serializeTo(buffer)
						binary.BigEndian.PutUint32(buffer[n:], uint32(len(buffer)-n-4))
					}
					wr := gzip.NewWriter(io.Discard)
					if _, err := wr.Write(buffer); err != nil {
						b.Fatal(err)
					}
					if err := wr.Close(); err != nil {
						b.Fatal(err)
					}
				}
			}
		})
	}
}

func TestFontIndexOnlyWritesAfterFileChanges(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "regular.ttf")
	if err := os.WriteFile(path, goregular.TTF, 0600); err != nil {
		t.Fatal(err)
	}
	cache := filepath.Join(t.TempDir(), "index.cache")
	logger := log.New(io.Discard, "", 0)
	refresh := func() systemFontsIndex {
		t.Helper()
		index, err := refreshFontIndex(logger, cache, dir)
		if err != nil {
			t.Fatal(err)
		}
		return index
	}
	refresh()
	stamp := time.Unix(100, 0)
	if err := os.Chtimes(cache, stamp, stamp); err != nil {
		t.Fatal(err)
	}
	refresh()
	stat, _ := os.Stat(cache)
	if !stat.ModTime().Equal(stamp) {
		t.Fatal("unchanged cache rewritten")
	}
	if err := os.Chtimes(path, stamp, stamp); err != nil {
		t.Fatal(err)
	}
	updated := refresh()
	stat, _ = os.Stat(cache)
	if stat.ModTime().Equal(stamp) || updated[0].modTime != timeStamp(stamp.UnixNano()) {
		t.Fatal("changed font not written to cache")
	}
}
