package el

import (
	"hash/fnv"
	"strconv"
	"strings"
	"testing"
)

// Identity must remain compatible with existing trees when optimizing hashing:
// scroll offsets, input contents and focus all live under these keys.
func TestChildKeyCompatibility(t *testing.T) {
	for _, parent := range []stateKey{0, 1, 0xffffffffffffffff} {
		for _, id := range []string{"", "button", "中文/入力", strings.Repeat("long-id/", 100)} {
			for _, index := range []int{0, 42, 123456, -1, -int(^uint(0)>>1) - 1, int(^uint(0) >> 1)} {
				h := fnv.New64a()
				var bytes [8]byte
				for i := range bytes {
					bytes[i] = byte(parent >> (8 * i))
				}
				h.Write(bytes[:])
				if id != "" {
					h.Write([]byte("#" + id))
				} else {
					h.Write([]byte(strconv.Itoa(index)))
				}
				if got, want := childKey(parent, id, index), stateKey(h.Sum64()); got != want {
					t.Fatalf("parent=%x id=%q index=%d: %x != %x", parent, id, index, got, want)
				}
			}
		}
	}
}
