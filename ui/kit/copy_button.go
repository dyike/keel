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
	text   func() string
	copied bool
}

// CopyButton copies what text returns when clicked; text runs at click time.
func CopyButton(text func() string) *CopyButtonView { return &CopyButtonView{text: text} }

func (v *CopyButtonView) Render(cx *el.Context) el.Element {
	id := autoID("copy", v)
	label, icon := locale.Current().Copy, IconCopy
	if v.copied {
		label, icon = locale.Current().Copied, IconCheck
		cx.After(copyKey{id}, CopiedFeedback, func() { v.copied = false })
	}
	return Button(label, func() {
		if v.text != nil {
			el.WriteClipboard(v.text())
			v.copied = true
		}
	}).ID(id).Icon(icon).Variant(ButtonGhost).Size(28).Render(cx)
}

type copyKey struct{ id string }
