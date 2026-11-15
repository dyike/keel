# Component capability gap inventory

[English](gallery-gaps.md) | [简体中文](gallery-gaps.zh-CN.md)

Compared against the 41 base component demos in the [NSH Gallery source](https://github.com/longbridge/nsh/blob/main/examples/gallery/App.tsx), retrieved locally on 2026-11-14. An entry point is not evidence of equivalent behavior, performance or mobile usability.

| Reference component | Keel example | Status |
| --- | --- | --- |
| Accordion | [accordion](../examples/components/accordion.go) | Entry exists; detailed parity unverified |
| Avatar | [avatar](../examples/components/avatar.go) | Entry exists; detailed parity unverified |
| Button | [button](../examples/components/button.go) | Entry exists; detailed parity unverified |
| Carousel | [carousel](../examples/components/carousel.go) | Entry exists; detailed parity unverified |
| Checkbox | [checkbox](../examples/components/checkbox.go) | Entry exists; detailed parity unverified |
| Collapsible | [collapsible](../examples/components/collapsible.go) | Entry exists; detailed parity unverified |
| ColorPicker | [color_picker](../examples/components/color_picker.go) | Entry exists; detailed parity unverified |
| Combobox | [combobox](../examples/components/combobox.go) | Entry exists; detailed parity unverified |
| Command | [command](../examples/components/command.go) | Entry exists; detailed parity unverified |
| DataTable | [table](../examples/components/table.go) | Entry exists; detailed parity unverified |
| DatePicker | [date_picker](../examples/components/date_picker.go) | Entry exists; detailed parity unverified |
| DiffView | `kit.Diff` | Unified/split review and patches implemented; synchronized scrolling and grapheme-aware intraline differences implemented; single-file patch input implemented; split context folding implemented; gutter callbacks implemented; embedded custom hunk actions implemented; broader device/visual parity unverified |
| Dock | [dock](../examples/components/dock.go) | Entry exists; detailed parity unverified |
| Editor | [code_editor](../examples/components/code_editor.go) | Entry exists; detailed parity unverified |
| EmojiPicker | [emoji_picker](../examples/components/emoji_picker.go) | Unicode catalog, search, tones, recents, virtual grid and adaptive popover/sheet implemented; physical-device/font parity unverified |
| HoverCard | [hover_card](../examples/components/hover_card.go) | Entry exists; detailed parity unverified |
| Input | [input](../examples/components/input.go) | Entry exists; detailed parity unverified |
| Link | [link](../examples/components/link.go) | Entry exists; detailed parity unverified |
| Menu | [menu](../examples/components/menu.go) | Entry exists; detailed parity unverified |
| NavStack | [nav_stack](../examples/components/nav_stack.go) | Stack/input/scroll, modal dismissal/focus and edge-swipe implemented; physical-device acceptance pending |
| NumberInput | [number_input](../examples/components/number_input.go) | Entry exists; detailed parity unverified |
| OtpInput | [otp_input](../examples/components/otp_input.go) | Entry exists; detailed parity unverified |
| Pagination | [pagination](../examples/components/pagination.go) | Entry exists; detailed parity unverified |
| Plot | [plot](../examples/components/plot.go) | Entry exists; detailed parity unverified |
| Popover | [popover](../examples/components/popover.go) | Entry exists; detailed parity unverified |
| Progress | [progress](../examples/components/progress.go) | Entry exists; detailed parity unverified |
| Radio | [radio_group](../examples/components/radio_group.go) | Entry exists; detailed parity unverified |
| Resizable | [resizable_group](../examples/components/resizable_group.go) | Entry exists; detailed parity unverified |
| RichEditor | — | Missing equivalent |
| Select | [select](../examples/components/select.go) | Entry exists; detailed parity unverified |
| Sheet | [sheet](../examples/components/sheet.go) | Entry exists; detailed parity unverified |
| Slider | [slider](../examples/components/slider.go) | Entry exists; detailed parity unverified |
| Switch | [switch](../examples/components/switch.go) | Entry exists; detailed parity unverified |
| Tabs | [tabs](../examples/components/tabs.go) | Entry exists; detailed parity unverified |
| Terminal | [terminal](../examples/components/terminal.go) | ANSI display, application-owned I/O, selection/copy and scrolling implemented; full VT compatibility, real PTY/IME and device acceptance remain unverified |
| TextView | [text_view](../examples/components/text_view.go) | Entry exists; detailed parity unverified |
| Toast | [notifier](../examples/components/notifier.go) | Entry exists; detailed parity unverified |
| Toggle | [toggle](../examples/components/toggle.go) | Entry exists; detailed parity unverified |
| Toolbar | [toolbar](../examples/components/toolbar.go) | Entry exists; detailed parity unverified |
| Tooltip | [tooltip](../examples/components/tooltip.go) | Entry exists; detailed parity unverified |
| Tree | [tree](../examples/components/tree.go) | Entry exists; detailed parity unverified |

NavStack now provides retained navigation and back behavior; modal dismissal/focus and edge-swipe are implemented; physical-device acceptance remains pending. Diff now provides review actions and patch export; synchronized split scrolling and grapheme-aware intraline highlighting are implemented; single-file patch input implemented; split context folding implemented; gutter callbacks implemented; embedded custom hunk actions implemented; broader device/visual parity unverified. EmojiPicker now includes a Unicode catalog, search, tones, recents and adaptive presentation; physical-device/font parity remains unverified. RichEditor remains missing. Terminal now has an ANSI display and interaction tests; full VT compatibility and real PTY/IME/device acceptance are not established. Each new component requires bilingual docs, a gallery entry, interaction tests and window snapshot tests.

Track primitive-level gaps separately: scroll paging/zoom, sticky lists, Canvas/Shader extensions and Video. The reference WebView dependency is not an equivalent in Keel’s pure-Go drawing architecture; adding HTML/WebView is not a valid shortcut to parity.

Diff scope was checked against the [reference DiffDemo](https://github.com/longbridge/nsh/blob/main/examples/gallery/demos/DiffDemo.tsx). The demo includes input patches, context expansion/collapse, line statistics and per-change/gutter actions; it does not demonstrate an editable merge pane. Earlier work notes treating editable merge as a closure gate were too broad. Keel now has statistics and unified context expansion/collapse; the remaining gaps above follow the actual reference.
