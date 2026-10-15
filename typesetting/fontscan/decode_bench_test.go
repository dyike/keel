package fontscan

import (
	"bytes"
	"testing"
)

func BenchmarkFontIndexDeserialize(b *testing.B) {
	for _, distinct := range []bool{false, true} {
		name := "shared"
		if distinct {
			name = "distinct"
		}
		b.Run(name, func(b *testing.B) {
			index := make(systemFontsIndex, 64)
			for i := range index {
				index[i].path = "font.ttc"
				for j := range 4 {
					runes := make(RuneSet, 512)
					for k := range runes {
						runes[k] = runePage{ref: uint16(k), set: pageSet{1, 2, 3}}
					}
					if distinct {
						runes[0].set[0] = uint32(1 + i*4 + j)
					}
					index[i].footprints = append(index[i].footprints, Footprint{Family: "family", Runes: runes})
				}
			}
			var encoded bytes.Buffer
			if err := index.serializeTo(&encoded); err != nil {
				b.Fatal(err)
			}
			b.ReportAllocs()
			for b.Loop() {
				if _, err := deserializeIndex(bytes.NewReader(encoded.Bytes())); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
