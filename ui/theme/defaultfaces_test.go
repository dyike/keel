package theme

import "testing"

func TestDefaultFontFacesShareTablesAndOwnList(t *testing.T) {
	a, b := DefaultFontFaces(), DefaultFontFaces()
	if len(a) == 0 || len(a) != len(b) {
		t.Fatal("missing default faces")
	}
	for i := range a {
		if a[i].Face.Face().Font != b[i].Face.Face().Font {
			t.Fatalf("face %d reloaded", i)
		}
	}
	original := b[0].Font
	a[0].Font.Typeface = "changed"
	if b[0].Font != original || DefaultFontFaces()[0].Font != original {
		t.Fatal("shared mutable face list")
	}
}
