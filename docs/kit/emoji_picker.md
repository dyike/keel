# EmojiPicker

English | [简体中文](emoji_picker.zh-CN.md)

A searchable emoji panel with categories, skin tones, recent choices and a virtualized grid. EmojiPicker() uses Unicode 17.0: all 3,944 fully-qualified sequences are available as entries or skin-tone variants. Mixed-tone sequences remain separate entries. The catalog is parsed on first use, not at package import; runtime search uses no network.

Use EmojiPicker().OnSelect(func(EmojiSelection)) for a composer: OnSelect fires for every choice, even selecting the same emoji twice. EmojiSelection contains exact Emoji output, the catalog Base and its Name. OnChange(func(string)) only fires when Value changes. SetValue is silent, accepts known sequences, ignores unknown values, and clears with an empty string. SetDisabled prevents interaction and closes popup presentation; programmatic assignments remain available.

Popup(true) renders a trigger. Viewports below a text-scaled 600dp use a bottom sheet; wider viewports use an anchored popover. The sheet uses the available width and places category tabs below the grid. A selection, Escape or outside dismissal closes it. SetOpen is silent; Open reports state. Inline mode is the default. Name labels the panel/trigger. Height sets the grid viewport, 264dp by default; enclosing panels handle remaining controls.

SetQuery and Query control search. Search matches whitespace-separated terms against names/keywords, case-insensitively; all terms must match. Exact emoji and registered skin-tone sequences also match; an exact variant query selects that variant. Default Unicode names are English. Category controls and picker UI follow locale; applications may supply localized emoji names and Keywords.

SetSkinTone accepts 0 (default) or 1–5 (light to dark). Only explicitly registered variants are emitted; unsupported tones fall back to the base. Changing tone does not mutate the already selected value. Groups returns an independent list; SetGroup selects a group or all groups with an empty string. ShowRecent selects the recent group. Recent returns a copy of up to 18 unique bases, newest first. SetRecent normalizes known variants to bases and drops unknown values. Choices update recents automatically; storage across sessions is application-owned. Recents use the current preferred tone.

EmojiPicker(entries...) and Data([]EmojiEntry) support custom catalogs. Data copies entries and keyword slices; empty/duplicate outputs are ignored, unsupported selected/recent values are removed silently. Data(nil) explicitly empties the catalog. EmojiEntry has Emoji, Name, Group, Keywords and five SkinTones. DefaultEmojis returns an independent catalog copy for localization. The generator scripts/generate-emoji.py accepts a cached --input file and verifies the pinned source hash; --check checks reproducibility. Source and license are recorded under third_party/unicode.

Each cell has its full name, exact value and selected/disabled semantics. Arrow keys navigate, Home/End reach the first/last match, Enter/Space choose. Tab enters the active grid cell. Search is focused only through normal focus/input actions. Cells remain at least 44dp and grow with text scale; the number of columns follows available width. Category/tone strips scroll horizontally. Hover/focus previews the name below. Only visible grid rows and nearby overscan are built.

Rendering uses system emoji fonts, not bundled images. Platform font coverage and color glyph support determine appearance; unsupported newer emoji can show missing glyphs. Device/IME and cross-platform font parity require separate acceptance.

Validation: go test ./ui/kit ./ui/window -run TestEmojiPicker; go run ./examples/components -section emoji_picker.
