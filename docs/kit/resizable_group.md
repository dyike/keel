# ResizableGroup

English | [简体中文](resizable_group.zh-CN.md)

Independent multi-panel split container supports horizontal and vertical arrangement and nesting. Coexists with double-sided Resizable.

```go
g := kit.ResizableGroup(
    kit.ResizablePanel{ID: "files", Content: files, Size: 180, Min: 100, Max: 300},
    kit.ResizablePanel{ID: "editor", Content: editor, Min: 160},
    kit.ResizablePanel{ID: "preview", Content: preview, Size: 240, Min: 100},
)
```

The default is horizontal layout; Vertical changes to vertical layout. The ID must be non-empty and unique. Empty IDs and subsequent entries with duplicate IDs are ignored. Size is the initial dp, 0 initially divides the remaining space outside the specified size panel; Min defaults to 0, and Max is 0, which means there is no upper limit. When the upper limit is less than Min, Min shall prevail. Put into a parent container of certain size.

Dragging only changes the two adjacent visible panels, leaving other panels unchanged. Arrow keys 16dp each time, Home/End to adjacent range boundary; cancel to retain the last valid size. SetDisabled disables both the handle and the content, and HandleAppearance uses the same handle configuration as Resizable.

Window changes allocate margin from the last visible panel forward, retaining the size of the front panel; leaving the tail empty when the upper limit is reached. When there is not enough space to accommodate the minimum size, it will be compressed according to the Min ratio, but each 6dp handle will still be retained. Request next frame convergence after first measurement.

Sizes Returns a copy of the size indexed by ID. SetSizes sets a limited non-negative size of a known ID, does not trigger a callback, and converges to the container range in the next frame; explicit 0 is no longer used as an automatic equalization. OnChange only returns a full size copy after the user has actually resized it.

SetVisible(id, visible) controls the visibility of a single panel, and the hidden content continues to be declared and retains its status and size; the remaining panels reallocate space, and one side still adheres to its own Max. SetPanels replaces/rearranges the configuration, retains the size of the same ID, and uses Size for new IDs; use SetSizes to change the size of existing IDs. The status of removed panels and handles has been cleaned up. OnChange is not triggered by visibility, configuration, and window changes.
