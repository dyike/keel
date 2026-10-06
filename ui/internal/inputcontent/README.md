# ui/internal/inputcontent

English | [简体中文](README.zh-CN.md)

Pure content layer for atomic reference inputs, used by el.InputDocument. Saves UTF-8 commit text, reference IDs/display names/ranges, and selections and undo transactions; no dependencies on Gio or other Keel packages.

Content is an independent draft, and the reference range must match the original text and fall on grapheme boundaries. Presentation maps the presentation text to the byte coordinates of the display name. Session processes substitutions according to clear editing intervals to prevent text diff from missing the operation of "replacing the same text but removing the reference"; only one undo of the input method combination transaction is recorded.

The UI layer is responsible for focus, read-only/disabled, platform events, copy-paste, hit and draw. The combined input test here only proves the transaction behavior, not the native candidate window or the correct display of the combined underline.

LayoutPresentation specifies an editor-internal display substitution string for a reference, such as a single private character for a placeholder glyph. It does not modify the commit text and reference metadata; use SourceRange to map the display editing range to the original text, and then call Session.ReplaceSource to continue inputting and undoing transactions using the original combination. The copy still takes the original text from Session.SelectedText, and the layout placeholder string cannot be copied. The object layout event path of Input/Textarea has been connected; the platform IME uses readable label projection, and the internal editor uses object projection, both of which are mapped to the original text.
