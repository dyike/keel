package kit

import (
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/theme"
)

// ColorPickerSize controls panel and trigger density; Medium preserves defaults.
type ColorPickerSize uint8

const (
	ColorPickerSizeMedium ColorPickerSize = iota
	ColorPickerSizeXSmall
	ColorPickerSizeSmall
	ColorPickerSizeLarge
)

func (p *ColorPickerView) Size(size ColorPickerSize) *ColorPickerView {
	if size <= ColorPickerSizeLarge {
		p.size = size
	}
	return p
}

// Label displays a caption above the inline picker or popup trigger.
func (p *ColorPickerView) Label(label string) *ColorPickerView { p.label = label; return p }

// Icon replaces the popup trigger's color swatch with an icon. IconNone restores it.
func (p *ColorPickerView) Icon(icon IconName) *ColorPickerView { p.icon = icon; return p }

// Popup opts into a built-in trigger and anchored panel. The default is inline.
func (p *ColorPickerView) Popup(on bool) *ColorPickerView {
	p.popup = on
	if !on && p.popover != nil {
		p.popover.SetValue(false)
		p.cancelPopupDraft()
	}
	return p
}

// SetOpen changes popup visibility without changing the color or invoking OnChange.
func (p *ColorPickerView) SetOpen(open bool) {
	if p.popup {
		p.ensurePopover()
		p.popover.SetValue(open && !p.disabled)
		if !p.popover.Value() {
			p.cancelPopupDraft()
		}
	}
}
func (p *ColorPickerView) IsOpen() bool { return p.popup && p.popover != nil && p.popover.Value() }

type colorPickerMetrics struct{ width, square, control, swatch float32 }

func (p *ColorPickerView) metrics() colorPickerMetrics {
	switch p.size {
	case ColorPickerSizeXSmall:
		return colorPickerMetrics{192, 112, 24, 20}
	case ColorPickerSizeSmall:
		return colorPickerMetrics{216, 130, 28, 22}
	case ColorPickerSizeLarge:
		return colorPickerMetrics{280, 180, 40, 28}
	default:
		return colorPickerMetrics{pickerWidth, 150, 32, 24}
	}
}
func (p *ColorPickerView) ensurePopover() {
	if p.popover == nil {
		p.popover = Popover(el.ViewFunc(p.renderPanel))
		if p.placed {
			p.popover.Placement(p.side, p.align)
		}
	}
}
func (p *ColorPickerView) cancelPopupDraft() {
	p.focused = false
	p.syncDrafts()
}
func (p *ColorPickerView) Render(cx *el.Context) el.Element {
	var content el.Element
	if p.popup {
		p.ensurePopover()
		p.popover.SetDisabled(p.disabled)
		id := autoID("color", p)
		p.popover.OnChange(func(open bool) {
			if open {
				cx.Focus(id + "/shade")
			} else {
				p.cancelPopupDraft()
				cx.Focus(id + "/trigger")
			}
		})
		label := p.label
		if label == "" {
			label = p.Text()
		}
		trigger := Button("", p.popover.Toggle).ID(id + "/trigger").Name(label).Variant(ButtonSecondary).Size(p.metrics().control)
		trigger.Content(el.ViewFunc(func(cx *el.Context) el.Element {
			row := el.Div().Row().Items(el.Center).Gap(theme.SpaceSm)
			if p.icon != IconNone {
				row.Child(Icon(p.icon).Size(16).Render(cx))
			} else {
				row.Child(el.Div().Size(el.Dp(16)).Rounded(theme.RadiusSm).Border(1, theme.Border).Bg(p.Value()))
			}
			return row.Child(el.Text(p.Text()).MaxLines(1))
		}))
		trigger.SetDisabled(p.disabled)
		p.popover.Trigger(trigger)
		content = p.popover.Render(cx)
	} else {
		content = p.renderPanel(cx)
	}
	if p.label == "" {
		return content
	}
	return el.Div().Gap(theme.SpaceSm).Items(el.Start).Child(el.Text(p.label).TextColor(theme.Muted).TextSize(theme.TextSm), content)
}
