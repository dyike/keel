package markdown

import (
	"image"
	"image/color"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/dyike/keel/third_party/gio/font"
	"github.com/dyike/keel/third_party/gio/op"
	"github.com/dyike/keel/third_party/gio/op/clip"
	"github.com/dyike/keel/third_party/gio/op/paint"
	"github.com/dyike/keel/third_party/gio/unit"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/theme"
)

type fadeStyle struct {
	font              font.Font
	size, rise        unit.Sp
	color, background color.NRGBA
	strike, underline bool
}

type streamFadeState struct {
	active      bool
	enabled     bool
	duration    time.Duration
	durationSet bool
	styles      []fadeStyle
	text        []rune
	times       []time.Time
	revision    uint64
	source      string
	ready       bool
}

// StreamFade opts in to fading newly appended text over 350ms with cubic
// ease-out. Replacing content shows it immediately. Existing text, selection,
// search highlights and card controls do not fade. Reduced motion disables it.
func (d *Doc) StreamFade(enabled bool) *Doc {
	if d.fade.enabled != enabled {
		d.fade.active = false
		d.fade.enabled = enabled
		d.fade.ready = false
		d.fade.times, d.fade.styles, d.fade.text = nil, nil, nil
	}
	return d
}

// StreamFadeDuration sets a duration from zero (instant) through ten seconds.
// Values outside that range are ignored. The default is 350ms.
func (d *Doc) StreamFadeDuration(duration time.Duration) *Doc {
	if duration >= 0 && duration <= 10*time.Second {
		d.fade.duration, d.fade.durationSet = duration, true
	}
	return d
}

func (f *streamFadeState) length() time.Duration {
	if f.durationSet {
		return f.duration
	}
	return 350 * time.Millisecond
}

func (d *Doc) syncStreamFade(gtx core.C) {
	f := &d.fade
	d.selection.fade = f
	if !f.enabled {
		return
	}
	changed := !f.ready || f.revision != d.selection.revision || f.source != d.src
	if changed {
		text := d.selection.text
		styles := make([]fadeStyle, len(text))
		for _, p := range d.selection.parts {
			pos := p.start
			for _, r := range p.r.runs {
				style := fadeStyle{font: r.font, size: r.size, rise: r.rise, color: r.color, strike: r.strike, underline: r.underline}
				if r.bg != nil {
					style.background = *r.bg
				}
				end := min(p.end, pos+utf8.RuneCountInString(r.text))
				for i := pos; i < end; i++ {
					styles[i] = style
				}
				pos = end
			}
		}
		next := make([]time.Time, len(text))
		appended := f.ready && len(d.src) > len(f.source) && strings.HasPrefix(d.src, f.source)
		if !theme.ReducedMotion && f.length() > 0 {
			for i, ch := range text {
				same := i < len(f.text) && f.text[i] == ch && f.styles[i] == styles[i]
				if same && (appended || f.source == d.src) {
					next[i] = f.times[i]
				} else if appended {
					next[i] = gtx.Now
				}
			}
		}
		f.text, f.styles, f.times = append(f.text[:0], text...), styles, next
		f.source, f.revision, f.ready = d.src, d.selection.revision, true
	}
	if !changed && !f.active {
		return
	}
	active := false
	for i, start := range f.times {
		if start.IsZero() {
			continue
		}
		if theme.ReducedMotion || f.length() <= 0 || gtx.Now.Sub(start) >= f.length() {
			f.times[i] = time.Time{}
		} else {
			active = true
		}
	}
	if active {
		gtx.Execute(op.InvalidateCmd{})
	}
	f.active = active
}

func (f *streamFadeState) alpha(now, start time.Time) float32 {
	if start.IsZero() || !f.enabled || theme.ReducedMotion || f.length() <= 0 {
		return 1
	}
	t := max(0, min(1, float32(now.Sub(start))/float32(f.length())))
	return 1 - (1-t)*(1-t)*(1-t)
}

func (r *richText) paintFadedPiece(gtx core.C, p piece, draw func()) {
	if r.document == nil || r.document.fade == nil || r.decoration || p.run >= r.textRuns || r.runs[p.run].image != nil || r.runs[p.run].object != nil {
		draw()
		return
	}
	f := r.document.fade
	begin, end := r.offset+p.start, r.offset+p.start+p.runes
	if !f.enabled || !f.active || begin < 0 || end > len(f.times) || begin == end {
		draw()
		return
	}
	// Split only the paint operation; the text is shaped once, so animation does
	// not change wrapping, glyph clusters, selection or link hit areas.
	for at := begin; at < end; {
		next := at + 1
		for next < end && f.times[next] == f.times[at] {
			next++
		}
		alpha := f.alpha(gtx.Now, f.times[at])
		if at == begin && next == end && alpha == 1 {
			draw()
			return
		}
		x0, x1 := p.rect.Min.X+p.xAt(at-begin), p.rect.Min.X+p.xAt(next-begin)
		if at == begin {
			x0 = -codeWidth
		}
		if next == end {
			x1 = codeWidth
		}
		if x1 > x0 && alpha > 0 {
			area := clip.Rect(image.Rect(x0, -codeWidth, x1, codeWidth)).Push(gtx.Ops)
			if alpha == 1 {
				// Keep settled text on its original rendering path; an opacity
				// layer can change glyph antialiasing even at full opacity.
				draw()
			} else {
				opacity := paint.PushOpacity(gtx.Ops, alpha)
				draw()
				opacity.Pop()
			}
			area.Pop()
		}
		at = next
	}
}
