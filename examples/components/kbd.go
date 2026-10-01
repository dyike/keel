package main

import (
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/layout"
	"github.com/dyike/keel/ui/widget"
)

func init() { registerSection("kbd", "controls", kbdGallery) }
func kbdGallery() core.Widget {
	return layout.Card(widget.Heading("快捷键键帽"), layout.Row(widget.Text("命令面板"), widget.Kbd("mod+shift+p"), widget.Text("确认"), widget.Kbd("enter")), layout.Row(widget.Kbd("mod+s").Plain(), widget.Kbd("esc").Size(widget.Large), widget.Kbd("alt+backspace").Size(widget.Small)))
}
