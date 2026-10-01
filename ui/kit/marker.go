package kit

import (
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/theme"
)

// MarkerView pairs a fixed-size colored dot with a readable status label.
type MarkerView struct {
	label string
	tone  Tone
}

func Marker(label string) *MarkerView         { return &MarkerView{label: label, tone: Info} }
func (v *MarkerView) SetLabel(s string)       { v.label = s }
func (v *MarkerView) Tone(t Tone) *MarkerView { v.tone = t; return v }
func (v *MarkerView) Render(*el.Context) el.Element {
	return el.Div().W(el.Full).Role("marker").Name(v.label).Value(v.tone.name()).Row().Gap(8).Items(el.Center).Child(el.Div().Size(el.Dp(8)).NoShrink().Rounded(4).Bg(v.tone.color()), el.Text(v.label).Grow().TextColor(theme.Text))
}
