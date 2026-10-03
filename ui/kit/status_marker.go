package kit

import (
	"math"
	"time"

	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/theme"
)

type StatusMarkerVariant uint8

const (
	StatusMarkerPlain StatusMarkerVariant = iota
	StatusMarkerSeparator
	StatusMarkerBorder
)

type StatusMarkerLoadingStyle uint8

const (
	StatusMarkerLoadingStyleSpinner StatusMarkerLoadingStyle = iota
	StatusMarkerLoadingStyleShimmer
)

type StatusMarkerPart uint8

const (
	StatusMarkerPartRoot StatusMarkerPart = iota
	StatusMarkerPartRow
	StatusMarkerPartIcon
	StatusMarkerPartContent
	StatusMarkerPartSeparator
)

// StatusMarkerView composes a status or timeline boundary with optional loading.
// It does not own a notification, counter, or row click action.
type StatusMarkerView struct {
	text          string
	icon, content el.View
	children      []el.View
	variant       StatusMarkerVariant
	align         el.Align
	alignSet      bool
	loading       bool
	loadingStyle  StatusMarkerLoadingStyle
	shimmer       *ShimmerTextView
	styles        [5]func(*el.DivEl)
	id, role      string
}

func StatusMarker(text string) *StatusMarkerView {
	return &StatusMarkerView{text: text, shimmer: ShimmerText(text)}
}
func (v *StatusMarkerView) SetText(text string)                { v.text = text; v.shimmer.SetText(text) }
func (v *StatusMarkerView) ID(id string) *StatusMarkerView     { v.id = id; return v }
func (v *StatusMarkerView) Role(role string) *StatusMarkerView { v.role = role; return v }
func (v *StatusMarkerView) Variant(value StatusMarkerVariant) *StatusMarkerView {
	if value <= StatusMarkerBorder {
		v.variant = value
	}
	return v
}

// Alignment places the row contents; only Start, Center and End are accepted.
// Without an override, separators center and other variants start.
func (v *StatusMarkerView) Alignment(value el.Align) *StatusMarkerView {
	if value == el.Start || value == el.Center || value == el.End {
		v.align = value
		v.alignSet = true
	}
	return v
}
func (v *StatusMarkerView) ResetAlignment() *StatusMarkerView { v.alignSet = false; return v }

// Icon supplies the compact icon slot; nil restores automatic spinner behavior.
func (v *StatusMarkerView) Icon(view el.View) *StatusMarkerView { v.icon = view; return v }

// Content adds rich content after the text. Empty text allows a rich-only marker.
func (v *StatusMarkerView) Content(view el.View) *StatusMarkerView { v.content = view; return v }

// Children replaces direct row children, after the icon/content slots.
func (v *StatusMarkerView) Children(views ...el.View) *StatusMarkerView {
	v.children = append([]el.View(nil), views...)
	return v
}
func (v *StatusMarkerView) Loading(on bool) *StatusMarkerView {
	if v.loading != on {
		v.shimmer.Restart()
	}
	v.loading = on
	return v
}
func (v *StatusMarkerView) LoadingStyle(style StatusMarkerLoadingStyle) *StatusMarkerView {
	if style <= StatusMarkerLoadingStyleShimmer {
		if v.loadingStyle != style {
			v.shimmer.Restart()
		}
		v.loadingStyle = style
	}
	return v
}

// ShimmerStyle configures the persistent text effect (duration, spread, color,
// direction and single-sweep mode). The marker controls its text/enabled state.
func (v *StatusMarkerView) ShimmerStyle(fn func(*ShimmerTextView)) *StatusMarkerView {
	if fn != nil {
		fn(v.shimmer)
	}
	return v
}

// PartStyle refines fresh elements after defaults. IDs remain stable for slots.
func (v *StatusMarkerView) PartStyle(part StatusMarkerPart, fn func(*el.DivEl)) *StatusMarkerView {
	if int(part) < len(v.styles) {
		v.styles[part] = fn
	}
	return v
}
func (v *StatusMarkerView) part(part StatusMarkerPart, id string, box *el.DivEl) *el.DivEl {
	if fn := v.styles[part]; fn != nil {
		fn(box)
	}
	return box.ID(id)
}
func (v *StatusMarkerView) Render(cx *el.Context) el.Element {
	id := v.id
	if id == "" {
		id = autoID("status-marker", v)
	}
	align := v.align
	if !v.alignSet {
		align = el.Start
		if v.variant == StatusMarkerSeparator {
			align = el.Center
		}
	}
	root := el.Div().WFull().Items(el.Stretch).TextColor(theme.Muted).TextAlign(align)
	row := el.Div().Row().WFull().MinH(el.Dp(theme.TextMd)).Items(el.Center).Justify(align).Gap(theme.SpaceSm)
	line := func(suffix string) *el.DivEl {
		return v.part(StatusMarkerPartSeparator, id+suffix, el.Div().H(el.Dp(1)).W(el.Dp(0)).Grow().Bg(theme.Border))
	}
	if v.variant == StatusMarkerSeparator && align != el.Start {
		row.Child(line("/leading"))
	}
	if v.icon != nil || v.loading && v.loadingStyle == StatusMarkerLoadingStyleSpinner {
		slot := el.Div().Size(el.Dp(16)).NoShrink().Items(el.Center).Justify(el.Center)
		if v.icon != nil {
			slot.Child(v.icon.Render(cx))
		} else {
			slot.Child(spinnerRing(cx, 16, theme.Muted))
		}
		row.Child(v.part(StatusMarkerPartIcon, id+"/icon", slot))
	}
	if v.text != "" || v.content != nil {
		content := el.Div().Row().MinW(el.Dp(0)).MaxW(el.Full).Items(el.Center).Gap(theme.SpaceSm)
		if v.text != "" {
			v.shimmer.SetText(v.text)
			v.shimmer.Enabled(v.loading && v.loadingStyle == StatusMarkerLoadingStyleShimmer)
			content.Child(v.shimmer.Render(cx))
		}
		if v.content != nil {
			content.Child(v.content.Render(cx))
		}
		if v.loading && v.loadingStyle == StatusMarkerLoadingStyleShimmer && v.text == "" && !el.ReducedMotion() {
			phase := float64(cx.Now().UnixNano()%int64(2*time.Second)) / float64(2*time.Second)
			content.Opacity(float32(.85 + .15*math.Cos(2*math.Pi*phase)))
			cx.Animating()
		}
		row.Child(v.part(StatusMarkerPartContent, id+"/content", content))
	}
	for _, child := range v.children {
		if child != nil {
			row.Child(child.Render(cx))
		}
	}
	if v.variant == StatusMarkerSeparator && align != el.End {
		row.Child(line("/trailing"))
	}
	if v.variant == StatusMarkerBorder {
		row.Pb(theme.SpaceSm)
	}
	root.Child(v.part(StatusMarkerPartRow, id+"/row", row))
	if v.variant == StatusMarkerBorder {
		root.Child(v.part(StatusMarkerPartSeparator, id+"/border", el.Div().WFull().H(el.Dp(1)).Bg(theme.Border)))
	}
	root = v.part(StatusMarkerPartRoot, id, root)
	if v.role != "" {
		root.Role(v.role)
	}
	return root
}
