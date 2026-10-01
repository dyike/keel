package kit

import (
	"github.com/dyike/keel/ui/theme"
	"image/color"
)

// Tone selects a semantic status color, resolved each time a view renders.
type Tone uint8

const (
	ToneNeutral Tone = iota
	ToneInfo
	ToneSuccess
	ToneWarning
	ToneDanger
)

func (t Tone) name() string {
	switch t {
	case ToneInfo:
		return "info"
	case ToneSuccess:
		return "success"
	case ToneWarning:
		return "warning"
	case ToneDanger:
		return "danger"
	default:
		return "neutral"
	}
}
func (t Tone) color() color.NRGBA {
	switch t {
	case ToneInfo:
		return theme.Info
	case ToneSuccess:
		return theme.Success
	case ToneWarning:
		return theme.Warning
	case ToneDanger:
		return theme.DangerText
	default:
		return theme.Text
	}
}
