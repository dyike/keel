# Dock

English | [简体中文](dock.zh-CN.md)

Like an IDE, dock tool panels to the left, right, and bottom of the center content.

```go
d := kit.Dock(editor).
    Panel(kit.DockPanel{ID: "files", Title: "文件", View: fileTree}, kit.DockLeft).
    Panel(kit.DockPanel{ID: "term", Title: "终端", View: terminal}, kit.DockBottom).
    OnLayoutChange(func(l kit.DockLayout) { save(l) })
if !d.SetLayout(loaded) { /* 不支持的版本或非法布局，保留原布局 */ }
```

- Multiple panels can be placed in each docking area and switched with tabs; the long tab bar can be scrolled horizontally; the docking area and the middle content can be dragged and resized.
- The menu to the right of the panel title bar can move the panel to the other side, or close it; `SetVisible(id, true)` will reopen in the original docking area.
- Maximize: Click "Maximize" in the menu or double-click the label to make this panel cover the entire Dock, with the middle content and other panels temporarily hidden; then click "Restore", double-click the label, or press Esc to restore. `Zoom(id)` / `Zoom("")` / `Zoomed()` are controlled in the program, and the maximized state is saved in `DockLayout.Zoomed`. Removing or closing a maximized panel automatically restores it.
- The panel will take the full height of the docking area. Components such as Tree, Table, and List can be filled with `Fill()` and have their own scrolling; longer ordinary content needs to be wrapped with a layer of `ScrollY`.
- `DockLayout` records which panels are in each docking area and their order, current label, size of each area, and closed panels. It can be directly encoded into JSON and saved. Version 2, the old versionless layout and version 1 are migrated to single label groups; unrecognized versions, duplicate span IDs, NaN/infinite sizes will be atomically rejected, `SetLayout` returns false and retains the original layout. `SetLayout` will ignore unknown panel IDs, and panels not mentioned in the layout will remain in place.
- `SetDisabled` Disables panel and layout operations. The separator bar can be adjusted by pressing 10dp with the direction keys, and Home/End to the border; cancel dragging to restore the original size. Normally clicking the separator bar will not trigger the layout callback.
- `OnLayoutChange` is called after the user moves, closes, switches panels, or drags to resize.
- When the window is too narrow, the left and right areas will be reduced proportionally, leaving at least 120dp for the middle content.

Agent: Each dock area is named `region` after the current panel title, the tab bar is `tablist`, the menu button is named "More Panel Title", and the menu items are "Dock to left", "Dock to right", "Dock to bottom" and "Close".

Verify: `go run ./examples/components -section dock`, add `-theme dark` to check the dark theme.

Repeatedly registering a panel ID will update the title and view and retain the area; empty IDs and illegal areas will not be added. Layout snapshots and restored inputs are copied and isolated, and `Visible` of unknown panels is false.

Nested splitting: `d.Split("search", "files", kit.DockPlacementBottom)` puts the search below the file group; also supports Left / Right / Top. The title menu provides "Split Right" and "Split Down" to separate the current tab from other tabs in the same group. The divider bar can be dragged, or adjusted with the arrow keys 5% at a time, from Home / End to 5% / 95%. Canceling a drag or disabling restores the pre-drag proportions. `Move` Move to the first label group in the specified area and select the moved label; empty groups are automatically merged. A closed group remains in the layout and returns to its original position when reopened.

`LeftTree` / `RightTree` / `BottomTree` Express nesting by `DockNode`: Panels / Active for leaves, First / Second / Axis / Ratio for splitting. While the tree exists, its panel order takes precedence over the old flat list; the old list still updates with the layout. Restoration deeply copies the tree, rejecting loops, over 32 levels, duplicate panels, missing branches, or out-of-bounds scale, and unknown panels culling and merging empty branches. Layout snapshots can be directly JSON round-tripped. Program calls `Split` / `Move` / `SetLayout` do not trigger layout callbacks.

When dragging a label, the title bar displays the insertion position; placing it in the middle of another group will merge it, and placing it on the four sides of the text will split it. Empty docking areas can be restored at the left, right, and bottom edges of the editor; if there are already panels in the corresponding area, they will be merged into one of the groups. The translucent color block represents the landing point. The layout is changed and called back once when released. The focus follows the moved label. Esc, Pointer Cancel, Drop Outside, or Disable all cancel the preview. The coordinates within the layout use dp, and drag and drop hits are judged according to the actual visible cropping range.

## Central Area Documentation

`Panel(p, kit.DockCenter)` Place the panel into the center area as a document. The center area, like the side areas, consists of label groups:

- Documents can be split (`Split`, title menu, drag to edge of group), dragged to merge, maximized, and layouts saved together via `CenterTree`, `Center`, `CenterActive`.
- When there is a document, the document replaces the center view passed in by `Dock(center)`; after all the documents are closed or moved, the center view reappears, which is suitable for empty states such as "no open documents".
- The side panel menu has an additional "move to the center" feature, and the center document menu can be docked to the left, right, or bottom. After using `DockCenter`'s Dock, when the center area is empty, drag the label to the middle and it will open as a document.
- When there is a document in the center area, only the 24dp narrow strip close to the edge is used to restore the empty side area, and the remaining edges are used to split the document group.

## Cross window

```go
d.OnDetach(func(p kit.DockPanel, reattach func()) {
    window.Open(window.Options{Title: p.Title, Content: el.Root(p.View), OnClose: reattach})
})
```

- After setting `OnDetach`, "Open in new window" appears in the panel menu. Drag the label out of the Dock and release it, and it will be detached.
- Detached panels are removed from the Dock, `Visible` is false, `Detached()` lists them, and `OnLayoutChange` is called once.
- `reattach` is called when the new window is closed, and the panel returns to its original docking area and tab group. To be called in the interface callback (`OnClose` originally), other goroutines use `core.Update`.
- Kit does not open windows by itself. The window size, title bar, etc. are determined by the application.
- Restoring the layout does not reopen the window: `SetLayout` Put the detached panel back into the Dock.
- When `OnDetach` is not set, dragging out of the Dock cancels as before.

The container should have a clear width and height, usually `el.Root(d)` is used directly. When embedding a normal page, specify the width and height for the outer layer; allocate the area inside the Dock according to the available space.

The first frame in the root viewport and each zoom will first allocate the docking area according to the current viewport, leaving 120dp for the center; when the viewport is smaller than this value, the docking area can be reduced to zero, and the center will use the remaining space. When embedding smaller containers, a redraw and correction will be requested after the actual size is measured on the first draw. "Preserve center" here does not mean that all panels are suitable for reading in mobile phone width.


## Panel status and factory

Just save the position and continue with `Layout/SetLayout`. To rebuild the panel in a new Dock, use `Snapshot/Restore` and register the factory by type:

```go
factory := func(s kit.DockPanelState) (kit.DockPanel, error) {
    input := kit.Input("搜索")
    var query string
    if len(s.State) > 0 {
        if err := json.Unmarshal(s.State, &query); err != nil {
            return kit.DockPanel{}, err
        }
    }
    input.SetValue(query)
    return kit.DockPanel{View: input, SaveState: func() (json.RawMessage, error) {
        return json.Marshal(input.Value())
    }}, nil // Restore fill in the original ID, Kind and Title
}
if err := d.RegisterPanel("search", factory); err != nil { /* 处理错误 */ }
state, err := d.Snapshot()
// json.Marshal(state) save; json.Unmarshal to kit.DockState after reading back.
if err == nil { err = anotherDock.Restore(state) } // anotherDock also needs to be registered search
```

`DockPanel.Kind: "search"` and `SaveState` are set when the panel is first added; the factory receives instance ID, type, title and JSON data and can restore multiple instances of the same type. The non-empty ID/Kind of the return value must match the snapshot, and the View must not be nil; the snapshot is inherited when the title is empty, otherwise the factory title is used. The application is responsible for data version migration, and the factory should also return the new `SaveState` to continue saving the edited value.

`DockState` is version 1, internally `DockLayout` is still version 2. Snapshots are sorted by ID, copy layout and JSON, including hidden, detached panels. The registry belongs to a single Dock, empty types, nil factories, and duplicate types return errors. Existing static panels without a Kind can be reused as-is by ID, but without carrying SaveState or data; to be restored across new instances, all panels should have a registered Kind.

Restore checks the entire list, JSON, and layout before calling the factory. Unknown types, duplicate IDs, panels with missing layout references, panels with no layout location in the manifest, illegal trees, or factory errors will all return errors that preserve the current Dock. After the restoration is successful, the panel collection will be based on the snapshot, and the current extra panels will be removed; the central empty view, appearance, disabled state, registry, and callbacks will be retained. The factory should only construct views, external side effects and created resources are managed by the application, and the Dock cannot roll back for the application.

These operations are called on the UI thread and do not trigger `OnLayoutChange`. Panel data editing does not trigger layout callbacks, apps should explicitly call `Snapshot` when saving the workspace or closing the window. When the split panel is restored to the Dock, the application window will not be automatically opened and closed; the old window closing callback will not affect the new instance after restoration.

## Independent appearance

```go
d.Skin(&kit.DockSkin{
    Header: func(e *el.DivEl) { e.Bg(theme.Bg) },
    Body: func(e *el.DivEl) { e.P(theme.SpaceLg) },
    Tab: func(e *el.DivEl, selected bool) {
        if selected { e.Bg(theme.Primary).TextColor(theme.PrimaryText) }
    },
    Separator: func(e *el.DivEl) { e.Bg(theme.Primary) },
})
```

`Panel` configures the outer frame of the label group, and `Header/Body/Tab/Separator` configures the title bar, body text, labels, and internal and external separators respectively. The callback is applied to a new element each frame, and the color, border, font size and panel blank can be changed; the element identity, sub-content and event handling are retained, and the separator bar maintains the 4dp geometry. Configuration objects can be shared and updated on the UI thread. `Skin(nil)` restores the default appearance without changing the layout or panel content. The skin does not enter the layout JSON, nor does it modify the global theme. It only cares about styling; the panel's own labels, toolbars, and menus are described in the next section.

The Component Library example's "Save Workspace/Restore Workspace" can verify that the search terms are restored with the layout, and "Toggle Dock Appearance" can be used to check the skin. Native windows, light-dark vision, and multi-window life cycles still require real-device validation.

## Panel labels, toolbars and menus

`DockPanel`'s optional fields let each panel determine how it looks and behaves in the Dock:

```go
d.Panel(kit.DockPanel{
    ID: "files", Title: "文件", View: files,
    Icon:    kit.IconFolder,                                     // Icons on labels
    Toolbar: kit.Button("", refresh).Name("刷新").Icon(kit.IconRetry).Variant(kit.ButtonGhost).Size(24),
    Menu:    func(m *kit.MenuView) { m.Item("全部折叠", "", collapseAll) },
    NoClose: true, // There is no "Close" in the menu
}, kit.DockLeft)
```

- `Icon` appears before the title; `Tab(selected)` completely replaces the label content (for example, with a status point), and `Title` remains the accessible name of the label.
- `Toolbar` Displays to the right of the title bar, before the menu button, when the panel is the current tab.
- `Menu` Add items to the panel menu, ranking before the Dock's own move, split, maximize, and close.
- `NoClose` removes "close"; `NoZoom` removes "maximize", double-clicking the label and `Zoom(id)` no longer maximizes it.
- `NoPadding` removes the white space in the text and is suitable for edge-drawing panels such as terminals and canvases.

## Collapse sidebar

```go
toolbar.Child(d.RegionButton(kit.DockLeft).Render(cx), d.RegionButton(kit.DockBottom).Render(cx))
```

`RegionButton(side)` returns a toggle button, which is selected when the sidebar is expanded and can be placed in the title bar or toolbar; the button is created once and then reused. `SetRegionOpen(side, open)` / `RegionOpen(side)` / `ToggleRegion(side)` are the corresponding program interfaces, and `ToggleRegion` will trigger OnLayoutChange. Collapse only hides the entire area, and the label, split and size of the panel are retained, and then expands it and returns it as it is; the collapsed state is stored in `LeftClosed/RightClosed/BottomClosed` of the layout. The center area cannot be retracted. Collapse of the sidebar where the maximized panel is located will exit maximization first.
