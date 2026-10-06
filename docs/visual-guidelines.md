# Visual guidelines

English | [简体中文](visual-guidelines.zh-CN.md)

This specification is used for review of new components and changes. For the source of colors and fonts, see [Theme](theme.md), and for component interfaces and status conventions, see [Component Development Specification](component-development.md). When a component has special dimensions, describe it in its own documentation.

## Fonts and alignment

Use `theme.BodySize` (15sp) for main text, `SmallSize` (13sp) for auxiliary text, and `HeadingSize` (22sp) for page titles. The font continues to be `theme.Face`, and Chinese, Latin and numbers use the same typesetting device; the icons use kit's Icon set, and no text characters are used to replace universal operation icons.

Text, cursor, selection, and placeholder text must share the typesetting origin and line height. Do not add offsets to controls separately based on whether the string contains Chinese characters. The visual correction of glyphs is handled uniformly at the theme / el / editorstyle layer; for troubleshooting steps and historical regression, see [Text and Cursor Fault Record](troubleshooting.md). For multi-line input, the viewport is calculated based on the actual measured line height. Empty lines also occupy one line; scroll after exceeding the viewport.

Icons and single lines of text are aligned to the center of the container. Buttons with sizes 28 / 32 / 40dp use 12 / 14 / 16sp text and 14 / 16 / 20dp icons respectively. Digital corner marks need to be checked separately for `4`, `99+` and Chinese. Chinese screenshots cannot be used in place of digital acceptance.

## Sizing and spacing

The spacing takes the constants in [theme scale](theme.md#spacing-and-sizing) (`theme.SpaceMd`, etc.), and does not write bare numbers.

| Purpose | Agreement |
| --- | --- |
| Action Buttons | Compact 28dp, Standard 32dp, Large 40dp; Rounded 6dp |
| Icons and labels | Spacing 6–8dp, icons do not squeeze the text |
| Form input | Standard single-line field height `theme.ControlHeight` (36dp), rounded corners 6dp, horizontal padding 10dp; multi-line fields calculate viewport based on row height |
| Controls in the same group | Usually spaced 8–12dp |
| Content blocks | Typically spaced 16–24dp; card or page padding 16–24dp |
| Separation and focus | Use semantic colors for borders, scale by dp, do not mix physical pixels |

Narrow windows prioritize wrapping, folding, or providing operable scrolling areas. The body text cannot be obscured; single-line labels are truncated while retaining the full accessible name. Components such as charts and tables that have a minimum readable size should specify their own scrolling or cropping behavior.

## Colors and states

The color is read from theme on each Render. `Text` is used for the main text, and `Muted` is used for description; `Highlight` is used to select the background, and `PrimaryText` is used for the text above it. Solid Primary / Danger buttons use `OnColor`. Success / Warning / Info are suitable for status text and icons. White text cannot be placed on these backgrounds by default.

| Status | Visible effects and behaviors |
| --- | --- |
| Hover / press | Use the hover / active color of the current theme; cannot change layout size |
| Focus | Tab, arrow keys and the program automatically focus to display the focus box; mouse clicks on ordinary controls retain the actual focus but do not display the focus box; the input box always has a border when focused; the Tab order is consistent with the visual order |
| Check | Color matches check, border, or explicit semantic state; cannot rely solely on red and green differences |
| Disabled | The appearance is weakened and input, dragging and callbacks are blocked; disabling the corresponding area also takes effect |
| Read-only | Reserved for reading, selection and copying; prohibiting modification of values, distinguished from disabled |
| Busy | Display progress or spinner, retain original width; prevent repeated submissions |
| Error | DangerText matches the error text and associates the corresponding fields; clear it promptly after recovery |
| Reduced motion | Follow the theme's valid preferences, retain the final state and necessary state prompts |

## Validation records

Each component covers at least light/dark, 1×/2×, standard/narrow containers. Components with Size interface are rechecked for compact and standard sizes; text is mixed in Chinese, English, and `0123456789`. Record what static screenshots and interactive tests have verified respectively, and do not write "can generate pictures" as "visual acceptance passed".

When adding screenshot issues, first locate the specific component, fix the problem and revert it back, and then submit it independently. Both the first frame and after interaction are checked; native window behavior and off-screen rendering are recorded separately and cannot replace each other.
