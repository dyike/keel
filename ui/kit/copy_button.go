package kit

import (
	"github.com/dyike/keel/ui/locale"
	"time"

	"github.com/dyike/keel/ui/el"
)

// CopiedFeedback is how long a CopyButton shows 已复制.
const CopiedFeedback = 1500 * time.Millisecond

// CopyButtonView copies text to the clipboard and confirms with 已复制.
type CopyButtonView struct {
	text     func() string
	copied   bool
	disabled bool
	revision uint64
}

// CopyButton copies what text returns when clicked; text runs at click time.
func CopyButton(text func() string) *CopyButtonView { return &CopyButtonView{text: text} }

func (v *CopyButtonView) Render(cx *el.Context) el.Element {
	id := autoID("copy", v)
	label, icon := locale.Current().Copy, IconCopy
	if v.copied {
		label, icon = locale.Current().Copied, IconCheck
		cx.AfterEnabled(id, copyKey{id, v.revision}, CopiedFeedback, func() { v.copied = false })
	}
	button := Button(label, func() {
		if v.text != nil {
			el.WriteClipboard(v.text())
			v.copied = true
			v.revision++
		}
	}).ID(id).Icon(icon).Variant(ButtonGhost).Size(28)
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
