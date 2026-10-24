// SPDX-License-Identifier: Unlicense OR MIT

package ops

import "testing"

// Writing ops into a reused buffer allocates nothing, also under -race
// (Keel patch), and every write returns zeroed bytes.
func TestWritesReuseBuffer(t *testing.T) {
	var o Ops
	ref := new(int)
	frame := func() {
		Reset(&o)
		for _, n := range []int{4, 17, 64, 9} {
			b := Write(&o, n)
			for i := range b {
				if b[i] != 0 {
					t.Fatalf("Write(%d) returned dirty byte %d", n, i)
				}
				b[i] = 0xff
			}
			Write2(&o, n, ref, ref)
		}
		BeginMulti(&o)
		for range 8 {
			WriteMulti(&o, 24)
		}
		EndMulti(&o)
	}
	frame()
	if allocs := testing.AllocsPerRun(50, frame); allocs != 0 {
		t.Fatalf("a frame of op writes allocated %.0f times", allocs)
	}
}
