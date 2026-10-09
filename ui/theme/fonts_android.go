package theme

import (
	"os"

	"gioui.org/font"
	"gioui.org/font/opentype"
)

// Gio scans system fonts through os.UserCacheDir, which has no usable default
// in Android apps. Load the platform's CJK collection directly so Chinese text
// works before the first window and in every independent Keel shaper.
func platformFaces() []font.FontFace {
	for _, path := range []string{
		"/system/fonts/NotoSansCJK-Regular.ttc",
		"/system/fonts/NotoSansSC-Regular.otf",
	} {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		faces, err := opentype.ParseCollection(data)
		if err == nil {
			return faces
		}
	}
	return nil // Apps can supply OEM-specific or bundled fonts with LoadFonts.
}
