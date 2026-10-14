package tables

import (
	"encoding/binary"
	"testing"
)

func TestFvarNamedInstanceStride(t *testing.T) {
	for _, stride := range []int{8, 10} {
		data := make([]byte, 36+2*stride)
		put16 := func(i, v int) { binary.BigEndian.PutUint16(data[i:], uint16(v)) }
		put32 := func(i, v int) { binary.BigEndian.PutUint32(data[i:], uint32(v)) }
		put16(0, 1)
		put16(4, 16)
		put16(6, 2)
		put16(8, 1)
		put16(10, 20)
		put16(12, 2)
		put16(14, stride)
		copy(data[16:], "wght")
		put32(20, 100<<16)
		put32(24, 500<<16)
		put32(28, 900<<16)
		for i, w := range []int{400, 600} {
			off := 36 + i*stride
			put16(off, 256+i)
			put32(off+4, w<<16)
			if stride == 10 {
				put16(off+8, 300+i)
			}
		}
		fv, _, err := ParseFvar(data)
		if err != nil {
			t.Fatal(err)
		}
		for i, w := range []float32{400, 600} {
			r := fv.Instances[i]
			if r.Coordinates[0] != w || int(r.SubfamilyNameID) != 256+i {
				t.Fatalf("stride %d: wrong instance %+v", stride, r)
			}
			if stride == 8 && r.PostScriptNameID != 0 {
				t.Fatal("read name from next instance")
			}
			if stride == 10 && int(r.PostScriptNameID) != 300+i {
				t.Fatal("lost PostScript name")
			}
		}
		if _, _, err := ParseFvar(data[:len(data)-1]); err == nil {
			t.Fatal("accepted truncated instance")
		}
	}
}
