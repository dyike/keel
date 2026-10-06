# Components

English | [简体中文](kit.zh-CN.md)

`ui/kit` provides ready-to-use el views. Browse by purpose below. Each page describes usage, APIs, and interaction behavior, with an interactive example on the documentation site.

## Basics

| Components | Purpose |
| --- | --- |
| [Label](kit/label.md) | Wrapping labels, secondary text and keyword highlighting |
| [Icon](kit/icon.md) | vector icon |
| [Image](kit/image.md) | Image: asynchronous loading, SVG, GIF animation, memory and disk cache, preview |
| [Avatar](kit/avatar.md) | Fixed size avatar and name fallback |
| [AvatarGroup](kit/avatar_group.md) | Stacked avatar groups, maximum number of people and overflow mark |
| [Badge](kit/badge.md) | Number/dot/icon badge, size and custom color |
| [Tag](kit/tag.md) | Select/remove tags, stroke, size, rounded corners, custom content and color |
| [Marker](kit/marker.md) | Geometric markers |
| [StatusMarker](kit/status_marker.md) | Message status, timeline boundaries and system prompts |
| [Kbd](kit/kbd.md) | Shortcut key caps, platform formats, independent/inherited font sizes and custom styles |
| [DescriptionList](kit/description_list.md) | Multiple columns/spanning columns, horizontal and vertical labels, custom values and separators |
| [GroupBox](kit/group_box.md) | Title/description grouping, appearance, outside-box footer and style configuration |

## Actions

| Components | Purpose |
| --- | --- |
| [Button](kit/button.md) | Operation buttons, icons, keyboard and loading state, selected state |
| [ButtonGroup](kit/button_group.md) | Several buttons connected into one, horizontal or vertical |
| [Link](kit/link.md) | Link |
| [CopyButton](kit/copy_button.md) | Copy and show feedback |
| [Tabs](kit/tabs.md) | Four visual variants, icons/rich tags, single item disable, upper width limit, scrolling/overflow and drag rearrangement |
| [Accordion](kit/accordion.md) | Folding panel, four sizes, optional border |
| [Collapsible](kit/collapsible.md) | Independent trigger and content, interruptible expansion animation |
| [Pagination](kit/pagination.md) | Pagination |
| [Stepper](kit/stepper.md) | Horizontal and vertical step progress, icon, size, single step disable |
| [Command](kit/command.md) | Command Panel |

## Inputs

| Components | Purpose |
| --- | --- |
| [Input / TextArea](kit/input.md) | Text box, suffix, clear, error |
| [Input Group](kit/input_group.md) | Input/multiline editor with four-way add-ons |
| [Checkbox](kit/checkbox.md) | Checkbox with an indeterminate state |
| [Switch](kit/switch.md) | switch |
| [Radio](kit/radio.md) | A single radio button that can be placed anywhere |
| [RadioGroup](kit/radio_group.md) | Radio group |
| [Toggle](kit/toggle.md) | Button with persistent pressed state |
| [ToggleGroup](kit/toggle_group.md) | Single- or multi-select button group |
| [Select](kit/select.md) | Drop-down selection, searchable |
| [Combobox](kit/combobox.md) | Filterable input |
| [NumberInput](kit/number_input.md) | Number input |
| [OtpInput](kit/otp_input.md) | Verification code, password mask, grouping and size |
| [TimeField](kit/time_field.md) | Time input |
| [Calendar](kit/calendar.md) | calendar, range |
| [DatePicker](kit/date_picker.md) | Date field |
| [Slider](kit/slider.md) | Linear/logarithmic slider, range selection and end callback |
| [Rating](kit/rating.md) | Star rating, points deduction for filled stars, custom size and color |
| [ColorPicker](kit/color_picker.md) | Color selection |
| [Form](kit/form.md) | Forms and validation |
| [Questionnaire](kit/questionnaire.md) | Paginated questionnaire and answer model |

## Overlays

| Components | Purpose |
| --- | --- |
| [Popover](kit/popover.md) | Trigger non-modal panel next to element |
| [Tooltip](kit/tooltip.md) | Rich content tips, action keys and positioning |
| [HoverCard](kit/hover_card.md) | Hover preview, instance delay and anchor positioning |
| [Menu](kit/menu.md) | Command menu and submenu |
| [DropdownButton](kit/dropdown_button.md) | Button with menu, split button |
| [Dialog](kit/dialog.md) | Modal dialog box, confirmation box; `Show(cx)` does not need to be placed in the view tree |
| [Sheet](kit/sheet.md) | A modal panel that slides in and can be dragged to adjust the size; `Show(cx)` does not need to be placed in the view tree |
| [Notifier](kit/notifier.md) | Eight-directional notification stack, rich content and operations, system notification delivery; `WindowNotifier(cx)` is the one that comes with the window |
| [Alert](kit/alert.md) | Inline/banner alerts, sizes, icons and rich content, supports light and dark colors |
| [Empty](kit/empty.md) | Empty state rich content, media, actions/footer and per-part styling |
| [Spinner](kit/spinner.md) | Indeterminate progress, reduced animation, custom icon/color/cycle |
| [Skeleton](kit/skeleton.md) | Placeholder, circle, custom rounded corner, secondary color level and Shimmer sweep |
| [Progress](kit/progress.md) | Progress bar, custom height/color/rounded corners and track style |
| [ProgressCircle](kit/progress_circle.md) | Circle progress, center content and indeterminate state |
| [ShimmerText](kit/shimmer_text.md) | Text sweep |

## Lists and tables

| Components | Purpose |
| --- | --- |
| [List](kit/list.md) | Long list: single/multiple selection, search, grouping, load more, status slots, drag to rearrange |
| [VirtualList](kit/virtual_list.md) | Equal height virtual list |
| [VariableList](kit/variable_list.md) | Natural height virtual list, stable key and reading position maintenance |
| [Tree](kit/tree.md) | Tree: lazy loading, multi-selection, dragging and rearranging (automatic expansion and scrolling) |
| [Table](kit/table.md) | Table: sorting, column width, cell slots |

## Chat

| Components | Purpose |
| --- | --- |
| [Message](kit/message.md) | Conversation message |
| [MessageContent](kit/message_content.md) | Mixing multiple bubbles, attachments and views in one message |
| [MessageGroup](kit/message_group.md) | Continuous message grouping |
| [Bubble](kit/bubble.md) | Chat Bubble |
| [BubbleGroup](kit/bubble_group.md) | Continuous bubble grouping |
| [MessageScroller](kit/message_scroller.md) | Conversation scrolling area |
| [Attachment](kit/attachment.md) | Attachment Card |
| [AttachmentGroup](kit/attachment_group.md) | Horizontally arranged, scrollable attachment group |
| [CodeEditor](kit/code_editor.md) | Code editor: highlighting, completion, diagnosis, multi-cursor, folding, soft line wrapping, find and replace |

## Charts

| Components | Purpose |
| --- | --- |
| [Chart](kit/chart.md) | Line chart, area chart, column chart, stacked column chart |
| [PieChart](kit/pie_chart.md) | Pie chart, donut chart, interactive legend |
| [CandlestickChart](kit/candlestick_chart.md) | Open/high/low/close data, dense-data aggregation, data table |
| [RadarChart](kit/radar_chart.md) | Radar Chart |
| [SankeyChart](kit/sankey_chart.md) | Sankey Chart |
| [Plot](kit/plot.md) | x/y plotting with zoom and pan |
| [Shared plotting primitives](kit/plot_primitives.md) | `ui/plot`: Scales and immediate-mode drawing for custom charts |

## Layout

| Components | Purpose |
| --- | --- |
| [TitleBar](kit/title_bar.md) | The title bar of a borderless window |
| [Sidebar](kit/sidebar.md) | Navigation sidebar |
| [Toolbar](kit/toolbar.md) | Toolbar and overflow menu |
| [StatusBar](kit/status_bar.md) | Fixed 24dp left and right status bar |
| [Resizable](kit/resizable.md) | Draggable separated two columns |
| [ResizableGroup](kit/resizable_group.md) | Multi-panel splitting, vertical and horizontal and nesting |
| [Dock](kit/dock.md) | Dockable panels, splitting, maximizing, collapsing sidebars, saving across windows and layouts |
| [Settings](kit/settings.md) | Settings page |
| [Carousel](kit/carousel.md) | Carousel: multiple viewports, dragging, trackpad and wheel, loop track |

## Shared conventions

The component instance saves the state, which should be created once when constructing the view, saved in the field, and then used in `Render(cx)`. `SetValue` is a programmatic assignment and does not trigger `OnChange`; only user operations trigger the callback. Background updates comply with [threading rules](architecture.md#threading).

The overlay component needs to be placed in `el.Root`. For layout and combination, see [Elements and Views](el.md), for colors and sizes, see [Theme](theme.md), and for new frame components, see [Component Development Specification](component-development.md).

## Optional features and binary size

Code highlighting and network images are introduced on demand:

| Function | Import | When not imported |
| --- | --- | --- |
| Code highlighting: CodeEditor, TextView, Markdown code blocks | `_ "github.com/dyike/keel/ui/highlight"` | Display plain text |
| Network pictures: http(s) addresses for Image, Avatar, Attachment, Markdown | `_ "github.com/dyike/keel/ui/netimage"` | Local files and data URLs as usual, the network address returns `core.ErrNoImageFetcher` |

Both items add approximately 4 MB each. `keel build` removes debugging information by default, and the application using only kit is about 10 MB.
