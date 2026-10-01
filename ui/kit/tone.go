package kit

import (
	"github.com/dyike/keel/ui/theme"
	"image/color"
)

// Tone selects a semantic status color, resolved each time a view renders.
type Tone uint8

const (
	Neutral Tone = iota
	Info
	Success
	Warning
	Danger
)

func (t Tone) name() string {
	switch t {
	case Info:
		return "info"
	case Success:
		return "success"
	case Warning:
		return "warning"
	case Danger:
		return "danger"
	default:
		return "neutral"
	}
}
func (t Tone) color() color.NRGBA {
	switch t {
	case Info:
		return theme.Info
	case Success:
		return theme.Success
	case Warning:
		return theme.Warning
	case Danger:
		return theme.DangerText
	default:
		return theme.Text
	}
}
