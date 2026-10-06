# Migration

English | [简体中文](migration-kit.zh-CN.md)

The repository has removed `ui/widget` and `ui/layout` and does not provide compatible aliases. Use `ui/el` for application structures and `ui/kit` for general components; windows are still managed by `ui/window`. The interface and behavior changes that affect callers during upgrades are listed below.

## Code highlighting and network images are introduced on demand (v0.0.7)

In order to reduce the program size, these two functions are no longer linked by default, each about 4 MB:

```go
import (
    _ "github.com/dyike/keel/ui/highlight" // Syntax color for CodeEditor, TextView, Markdown code blocks
    _ "github.com/dyike/keel/ui/netimage"  // Image, Avatar, Attachment, Markdown image http(s) address
)
```

After the upgrade, if the code becomes plain text, or the network picture reports `core.ErrNoImageFetcher`, just add the corresponding line to the main package. Local files and images from data URLs are not affected.

## Pages and status

The old Column / Row / Card is changed to `el.Div()`, and Row, Gap, Padding, Bg and Border are set as needed; the kit view joins the element tree through `Render(cx)`, and the page is handed over to the window through `el.Root(view)`. For specific parameters from the old component to the new component, see [Component Index](kit.md).

Component instances are reused during the page life cycle, especially input boxes, overlays, Tables, Trees, Docks and virtual lists. Don't rebuild stateful components every time you render. Programmatic assignments such as `SetValue` do not trigger user callbacks; modifications are made directly within the callback, and background updates are put into `core.Update`.

Collection components will copy the data slices or layout trees they own. Externally modifying the original slice will not refresh the component and should call SetItems, SetRows, SetKeys, SetData or SetLayout. Views, pictures and business callbacks are still held by reference; copying the configuration does not mean copying the business objects.

## Chat list

`MessageScroller` is now built row by row, requiring a stable key, estimated row height, and row constructor:

```go
scroller := kit.MessageScroller(keys, 100, func(cx *el.Context, index int) el.Element {
    return messages[index].Render(cx)
})
// After new messages or historical messages arrive, messages are updated first, and then a complete and unique key list is submitted.
scroller.SetKeys(keys)
```

Removed old full list construction callback and `HistoryPrepended` call. Head inset with height changes by stabilizing key to maintain reading anchor; follow to bottom with `SetFollow` / `ScrollToEnd`. Do not use the current array index as a message ID that will be inserted or reordered. Full example at `examples/chat`.

## Tables, Selections, and Forms

Lists and trees with reordering, filtering, or dynamic updates should use the stable flag. Table column layout, selection, and data interaction see [Table](kit/table.md), and do not rely on aliases returning slices to modify internal state. Select / Combobox distinguishes labels and values, and provides interfaces for grouping, disabled items, multi-selection and asynchronous results.

Asynchronous forms and searches submit results via the corresponding request token; older requests cannot overwrite new requests. Disabling the area where the component belongs will also prevent users from modifying it, and the application does not have to register the disabling callback repeatedly for each sub-control. The caller remains responsible for canceling external network tasks.

## Dock persistence

`DockLayout` The current version is 2, adding LeftTree / RightTree / BottomTree nested trees. Old versionless layouts and version 1 can be migrated by `SetLayout`; its bool return value must be checked. Illegal or unknown versions will not partially overwrite the current layout. Save the snapshot returned by `Layout()` without caching the internal tree pointer. Drag and drop within the window is handled by the Dock; detaching to a new window requires applying the setting `OnDetach` and is responsible for opening the window, and calling `reattach` when closing. See [Cross-window](kit/dock.md#cross-window).

## Windows and themes

Custom `core.WindowControls` implementation needs to supplement `Focused()` and `TitleBarArea(...)`. The former represents the activation state of the native window, and the latter accepts the window dp coordinates; it is cleared every frame and then registered by the title bar. Application controls should not be included in the drag area.

Use `theme.Apply` for runtime themes and do not modify global colors one by one. Reduced motion override explicitly with `SetReducedMotion`, or `FollowSystemMotion` to restore system preferences; current system bridge supports macOS. Other platforms can have preferences specified by the app.

## Verify

After migration, run `go build ./...`, `go vet ./...`, `go test ./... -count=1` first. When the page involves the order form, chat scrolling or Dock state, the corresponding behavior test is retained; visual modifications are reviewed according to [Visual Specification](visual-guidelines.md) and [Screenshot Matrix](testing.md#full-component-screenshot-matrix).
