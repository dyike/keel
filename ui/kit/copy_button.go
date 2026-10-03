package kit

import (
	"time"

	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/locale"
	"github.com/dyike/keel/ui/theme"
)

// CopiedFeedback is how long a CopyButton shows 已复制.
const CopiedFeedback = 1500 * time.Millisecond

// CopyButtonView copies text to the clipboard and confirms with 已复制.
type CopyButtonView struct {
	text     func() string
	onCopied func(string)
	content  el.View
	copied   bool
	disabled bool
	revision uint64
}

// CopyButton copies what text returns when clicked; text runs at click time.
func CopyButton(text func() string) *CopyButtonView { return &CopyButtonView{text: text} }

// OnCopied runs after requesting a clipboard write, with the exact copied text.
// The platform does not acknowledge whether the system clipboard accepted it.
func (v *CopyButtonView) OnCopied(fn func(string)) *CopyButtonView {
	v.onCopied = fn
	return v
}

// Content replaces the default icon and label. Supply display-only content,
// without nested buttons or inputs.
// Nil restores the default. Use Copied inside a ViewFunc to customize feedback.
func (v *CopyButtonView) Content(content el.View) *CopyButtonView {
	v.content = content
	return v
}

// Copied reports whether the current copy feedback interval is active.
func (v *CopyButtonView) Copied() bool { return v.copied }

func (v *CopyButtonView) copy() {
	if v.disabled || v.text == nil {
		return
	}
	text := v.text()
	el.WriteClipboard(text)
	v.copied = true
	v.revision++
	if v.onCopied != nil {
		v.onCopied(text)
	}
}

func (v *CopyButtonView) Render(cx *el.Context) el.Element {
	id := autoID("copy", v)
	label, icon := locale.Current().Copy, IconCopy
	if v.copied {
		label, icon = locale.Current().Copied, IconCheck
		cx.AfterEnabled(id, copyKey{id, v.revision}, CopiedFeedback, func() { v.copied = false })
	}
	if v.content != nil {
		return el.Div().ID(id).Role("button").Name(label).Focusable(true).
			OnClick(v.copy).Disabled(v.disabled).CursorPointer().MinH(el.Dp(28)).MaxW(el.Full).
			Px(theme.SpaceSm).Py(theme.SpaceXs).Rounded(theme.RadiusMd).
			TextSize(theme.TextSm).TextColor(theme.Text).
			Hover(func(s *el.Style) { s.Bg(theme.Subtle) }).
			FocusStyle(func(s *el.Style) { s.BorderColor(theme.PrimaryText) }).
			Child(v.content.Render(cx))
	}
	button := Button(label, v.copy).ID(id).Icon(icon).Variant(ButtonGhost).Size(28)
	button.SetDisabled(v.disabled)
	return button.Render(cx)
}

type copyKey struct {
	id       string
	revision uint64
}

func (v *CopyButtonView) SetDisabled(on bool) {
	v.disabled = on
	if on {
		v.copied = false
	}
}
