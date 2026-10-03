package el

import (
	"image/color"
	"testing"

	"gioui.org/text"
	"gioui.org/unit"
)

func TestTextAlignmentInheritanceAndLabel(t *testing.T) {
	c := color.NRGBA{A: 255}
	for _, tc := range []struct {
		align Align
		want  text.Alignment
	}{{Start, text.Start}, {Center, text.Middle}, {End, text.End}} {
		parent := Div().TextAlign(tc.align).node().style.text
		child := Text("long first line\nshort").node()
		child.textStyle = child.style.text.inherit(parent)
		child.textStyle.color = &c
		child.textStyle.size = unit.Sp(14)
		e := engine{}
		if got := e.label(child, child.text).Alignment; got != tc.want {
			t.Fatal("label alignment", got, tc.want)
		}
		override := Text("override").TextAlign(Start).node().style.text.inherit(parent)
		if override.align == nil || *override.align != Start {
			t.Fatal("explicit start did not override")
		}
	}
}
