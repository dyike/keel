package fontscan

import (
	"compress/gzip"
	"encoding/binary"
	"errors"
	"fmt"
	"hash/maphash"
	"io"
	"math"
	"os"
	"path/filepath"

	"github.com/dyike/keel/typesetting/font"
)

// defines the routines to serialize a font set to
// the disk

// assume len(dst) >= 4
func serializeFloat(f float32, dst []byte) {
	binary.BigEndian.PutUint32(dst, math.Float32bits(f))
}

// assume len(src) >= 4
func deserializeFloat(src []byte) float32 {
	return math.Float32frombits(binary.BigEndian.Uint32(src))
}

func serializeString(s string) []byte {
	L := len(s)
	if L > math.MaxUint16 { // never happen in practice
		L = math.MaxUint16
	}
	buffer := make([]byte, 2+L) // len as uint16 + data
	binary.BigEndian.PutUint16(buffer, uint16(L))
	copy(buffer[2:], s)
	return buffer
}

func deserializeString(s *string, data []byte) (int, error) {
	if len(data) < 2 {
		return 0, errors.New("invalid string (EOF)")
	}
	L := int(binary.BigEndian.Uint16(data))
	if len(data) < 2+L {
		return 0, errors.New("invalid string length (EOF)")
	}
	*s = string(data[2 : 2+L])
	return 2 + L, nil
}

const aspectSize = 1 + 4 + 4

// serializeTo serialize the Aspect in binary format
func serializeAspect(as font.Aspect) []byte {
	var buffer [aspectSize]byte
	buffer[0] = byte(as.Style)
	serializeFloat(float32(as.Weight), buffer[1:])
	serializeFloat(float32(as.Stretch), buffer[5:])
	return buffer[:]
}

// deserializeFrom reads the binary format produced by serializeTo
// it returns the number of bytes read from `data`
func deserializeAspectFrom(data []byte, as *font.Aspect) (int, error) {
	if len(data) < aspectSize {
		return 0, errors.New("invalid Aspect (EOF)")
	}
	as.Style = font.Style(data[0])
	as.Weight = font.Weight(deserializeFloat(data[1:]))
	as.Stretch = font.Stretch(deserializeFloat(data[5:]))
	return aspectSize, nil
}

// serializeTo serialize the Footprint in binary format,
// by appending to `dst` and returning the slice
func (fp Footprint) serializeTo(dst []byte) []byte {
	dst = append(dst, serializeString(fp.Location.File)...)

	var buffer [4]byte
	binary.BigEndian.PutUint16(buffer[:], fp.Location.Index)
	binary.BigEndian.PutUint16(buffer[2:], fp.Location.Instance)
	dst = append(dst, buffer[:]...)

	dst = append(dst, serializeString(fp.Family)...)
	dst = append(dst, fp.Runes.serialize()...)
	dst = append(dst, fp.Scripts.serialize()...)
	dst = append(dst, fp.Langs.serialize()...)
	dst = append(dst, serializeAspect(fp.Aspect)...)

	return dst
}

// deserializeFrom reads the binary format produced by serializeTo
// it returns the number of bytes read from `data`
func (fp *Footprint) deserializeFrom(data []byte) (int, error) {
	return fp.deserializeWithRunes(data, nil)
}

func (fp *Footprint) deserializeWithRunes(data []byte, runes *runeSetDecoder) (int, error) {
	n, err := deserializeString(&fp.Location.File, data)
	if err != nil {
		return 0, err
	}
	if len(data) < n+4 {
		return 0, errors.New("invalid Location (EOF)")
	}
	fp.Location.Index = binary.BigEndian.Uint16(data[n:])
	fp.Location.Instance = binary.BigEndian.Uint16(data[n+2:])
	n += 4

	read, err := deserializeString(&fp.Family, data[n:])
	if err != nil {
		return 0, err
	}
	n += read
	read, err = runes.decode(&fp.Runes, data[n:])
	if err != nil {
		return 0, err
	}
	n += read
	read, err = fp.Scripts.deserializeFrom(data[n:])
	if err != nil {
		return 0, err
	}
	n += read
	read, err = fp.Langs.deserializeFrom(data[n:])
	if err != nil {
		return 0, err
	}
	n += read
	read, err = deserializeAspectFrom(data[n:], &fp.Aspect)
	if err != nil {
		return 0, err
	}
	n += read

	return n, nil
}

// serialize into binary format, appending to `dst` and returning
// the updated slice
func serializeFootprintsTo(footprints []Footprint, dst []byte) []byte {
	for _, fp := range footprints {
		dst = fp.serializeTo(dst)
	}
	return dst
}

// parses the format written by `serializeFootprints`
func deserializeFootprints(src []byte) (out []Footprint, err error) {
	return deserializeFootprintsWithRunes(src, nil)
}

func deserializeFootprintsWithRunes(src []byte, runes *runeSetDecoder) (out []Footprint, err error) {
	for totalRead := 0; totalRead < len(src); {
		var fp Footprint
		read, err := fp.deserializeWithRunes(src[totalRead:], runes)
		if err != nil {
			return nil, fmt.Errorf("invalid footprints: %s", err)
		}
		totalRead += read

		out = append(out, fp)
	}

	return out, nil
}

func (ff fileFootprints) serializeTo(dst []byte) []byte {
	dst = append(dst, serializeString(ff.path)...)
	dst = append(dst, ff.modTime.serialize()...)
	// end by the variable length footprint list
	dst = serializeFootprintsTo(ff.footprints, dst)
	return dst
}

func (ff *fileFootprints) deserializeFrom(src []byte) error {
	return ff.deserializeWithRunes(src, nil)
}

func (ff *fileFootprints) deserializeWithRunes(src []byte, runes *runeSetDecoder) error {
	n, err := deserializeString(&ff.path, src)
	if err != nil {
		return err
	}
	if len(src) < n+8 {
		return errors.New("invalid fileFootprints (EOF)")
	}
	ff.modTime.deserialize(src[n:])
	n += 8
	ff.footprints, err = deserializeFootprintsWithRunes(src[n:], runes)
	if err != nil {
		return err
	}
	return nil
}

const cacheFormatVersion = 7 // include named weight instances

// serialize into binary format, compressed with gzip
func (index systemFontsIndex) serializeTo(w io.Writer) error {
	wr := gzip.NewWriter(w)
	defer wr.Close()
	var header [6]byte
	binary.BigEndian.PutUint16(header[:], cacheFormatVersion)
	binary.BigEndian.PutUint32(header[2:], uint32(len(index)))
	if _, err := wr.Write(header[:]); err != nil {
		return err
	}
	// The format is already record-based. Stream each file rather than
	// growing a second in-memory copy of the entire uncompressed index.
	var buffer []byte
	for _, ff := range index {
		buffer = ff.serializeTo(buffer[:0])
		binary.BigEndian.PutUint32(header[:4], uint32(len(buffer)))
		if _, err := wr.Write(header[:4]); err != nil {
			return err
		}
		if _, err := wr.Write(buffer); err != nil {
			return fmt.Errorf("serializing font footprints: %s", err)
		}
	}
	if err := wr.Close(); err != nil {
		return fmt.Errorf("compressing serialized font footprints: %s", err)
	}
	return nil
}

// parses the format written by `fontIndex.serializeTo`
func deserializeIndex(src io.Reader) (systemFontsIndex, error) {
	r, err := gzip.NewReader(src)
	if err != nil {
		return nil, fmt.Errorf("invalid compressed index file: %s", err)
	}
	defer r.Close()

	var (
		buf    [6]byte
		out    systemFontsIndex
		buffer []byte
	)

	// read the expected length
	if _, err := io.ReadFull(r, buf[:]); err != nil {
		return nil, fmt.Errorf("invalid index format: %s", err)
	}
	version := binary.BigEndian.Uint16(buf[:])
	if version != cacheFormatVersion {
		return nil, fmt.Errorf("different index version format: found %d", version)
	}
	L := binary.BigEndian.Uint32(buf[2:])
	runes := runeSetDecoder{seed: maphash.MakeSeed(), sets: make(map[uint64]RuneSet)}
	for i := uint32(0); i < L; i++ {
		// size of the encoded footprint
		if _, err := io.ReadFull(r, buf[:4]); err != nil {
			return nil, fmt.Errorf("invalid index: %s", err)
		}
		size := binary.BigEndian.Uint32(buf[:4])
		// buffer the fileFootprints segment
		buffer, err = readIndexRecord(r, size, buffer)
		if err != nil {
			return nil, fmt.Errorf("invalid index: %s", err)
		}

		var fp fileFootprints
		err = fp.deserializeWithRunes(buffer, &runes)
		if err != nil {
			return nil, fmt.Errorf("invalid index: %s", err)
		}

		out = append(out, fp)
	}

	return out, nil
}

// Read into retained storage without bytes.Buffer.ReadFrom's extra growth to
// discover EOF at each record boundary. Grow only as data arrives: an invalid
// length prefix must not cause a multi-gigabyte allocation for truncated input.
func readIndexRecord(r io.Reader, size uint32, buffer []byte) ([]byte, error) {
	buffer = buffer[:0]
	for uint64(len(buffer)) < uint64(size) {
		n := int(min(uint64(size)-uint64(len(buffer)), 32<<10))
		start := len(buffer)
		// After reading a prefix, reserve ordinary records once instead of
		// repeatedly growing and copying a multi-megabyte CJK record. Bound
		// speculative allocation; very large/truncated lengths stay incremental.
		if start != 0 && size <= 16<<20 {
			n = int(size) - start
		}
		buffer = append(buffer, make([]byte, n)...)
		if _, err := io.ReadFull(r, buffer[start:]); err != nil {
			return nil, err
		}
	}
	return buffer, nil
}

func deserializeIndexFile(cachePath string) (systemFontsIndex, error) {
	f, err := os.Open(cachePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	out, err := deserializeIndex(f)
	return out, err
}

func (index systemFontsIndex) serializeToFile(cachePath string) error {
	dir, _ := filepath.Split(cachePath)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("unable to create system font cache dir %q: %w", dir, err)
	}
	f, err := os.Create(cachePath)
	if err != nil {
		return err
	}

	err = index.serializeTo(f)
	if err != nil {
		return err
	}

	err = f.Close()
	return err
}
