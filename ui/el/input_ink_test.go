package el

import (
	"github.com/dyike/keel/third_party/gio/font"
	"github.com/dyike/keel/third_party/gio/unit"
	"github.com/dyike/keel/third_party/gio/widget"
	"github.com/dyike/keel/third_party/gio/widget/material"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/theme"
	"image/color"
	"testing"
)

func TestInputKeepsDescenderInk(t *testing.T) {
	for _, face := range []font.Typeface{theme.Material.Face, "Go"} {
		old := theme.Material.Face
		theme.Material.Face = face
		for _, scale := range []int{1, 2} {
			for _, size := range []float32{12, 15, 20, 32} {
				for _, value := range []string{"gypqj", "国Agyp", "Égj"} {
					value := value
					got := redInkScaled(t, Input().Bind(&value).TextSize(size).TextColor(color.NRGBA{R: 255, A: 255}).W(Dp(280)).Border(0, color.NRGBA{}).P(0).MinH(Auto), scale)
					var editor widget.Editor
					editor.SingleLine = true
					editor.SetText(value)
					want := redInkScaled(t, Widget(core.Func(func(gtx core.C) core.D {
						style := material.Editor(theme.Material, &editor, "")
						style.TextSize = unit.Sp(size)
						style.Color = color.NRGBA{R: 255, A: 255}
						return style.Layout(gtx)
					})).W(Dp(280)).H(Dp(110)), scale)
					if got < want*97/100 {
						t.Errorf("face=%s size=%g text=%s ink=%d want=%d", face, size, value, got, want)
					}
				}
			}
		}
		theme.Material.Face = old
	}
}
