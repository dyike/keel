# Carousel

English | [简体中文](carousel.zh-CN.md)

Carousel: Displays one picture at a time, with previous and next buttons and cue points.

```go
car := kit.Carousel(slide1, slide2, slide3).Height(200).Autoplay(4 * time.Second)
```

- Click the button or pointer to switch; after the carousel gets focus, use ← → horizontally and ↑ ↓ vertically to switch. The default is to cycle from beginning to end; Home/End selects the first/last image.
- `Autoplay(d)` Cut to the next picture every d. Pauses when the pointer hovers over the carousel; does not play automatically when the reduced motion is enabled.
- `SetDisabled(true)` disables button, cue point, and key switching, while stopping autoplay and removing focus; autoplay is also paused when the parent container is disabled, the content leaves the viewport, or is blocked by a modal layer, and the timer is restarted after recovery.
- `Value()` / `SetValue(i)` (does not trigger callback), `OnChange(fn)`.
- The names of "previous" and "next" come from locale.

Agent: container role `group`, the name is "Current/Total" (such as 2/3); the button names are "Previous" and "Next", and the name of the indication point is the serial number.

Verify: `go run ./examples/components -section carousel`, add `-theme dark` to check the dark theme.

`Vertical(true)` Place navigation to the right of the content: up arrow, vertical indicator dot, and down arrow. `Vertical(false)` Restores horizontal navigation below content. Height still controls content height (default 200dp), navigation remains at least 72dp; non-positive or non-finite heights are ignored. The cue point area can be scrolled vertically to accommodate more items. Switching direction retains the current index and existing keyboard focus, and does not trigger OnChange; the autoplay, hover pause, disable, and reduced motion rules remain unchanged.

The default is single item switching, and ItemsPerView can open multiple viewports of equal size; it already supports pointer dragging and snapping, and continuous scrolling will snap after a pause, see below. Before and after controls and pagination items can be combined independently, see below.

`Loop(false)` Closes the loop: disables the previous item when reaching the first item, disables the next item when reaching the last item, and keeps the original selection when the keyboard goes out of bounds. `Loop(true)` restores the default loop; switching modes does not change the index and does not trigger OnChange. The front and back buttons for empty and single-item lists are always disabled, and empty lists have an Agent count of 0/0.

`Previous()` / `Next()` shares the navigation logic with the built-in buttons, only calls OnChange when the index actually changes, and respects its own disabled state. `CanPrevious()` / `CanNext()` Returns availability combined with list number, loop mode, bounds and self-disabled state. They do not read the disabled state inherited from the parent container; external controls should be placed in the same disabled container or otherwise disabled by the app. `SetValue` is still used for program setup without callbacks and can be called even when disabled.

Non-loop automatic playback stops after the last item; after the program or user returns to the previous item, the timer starts again from the next frame. Other hover, reduced motion, and visibility pause rules remain unchanged.

```go
car.Loop(false)
previous := kit.Button("Previous item", car.Previous)
previous.SetDisabled(!car.CanPrevious()) // Update according to current status before rendering
```

## Independent combination

Keep the returned control instances, rendering them every frame:

```go
content := car.Content()
previous := car.PreviousControl(nil)
next := car.NextControl(kit.Button("Next item", nil).Variant(kit.ButtonSecondary).Size(40))
first := car.PaginationItem(0, nil)
second := car.PaginationItem(1, nil)
// Call content/previous/next/first/second.Render(cx) respectively in your own layout.
```

`Content()` Only displays the content area, retaining the arrow keys, Home/End, focus and autoplay. The same CarouselView only mounts Content or the complete Carousel once; independent controls can create multiple copies. When nil is passed in, the front and rear controls display arrows according to the current axis, the paging items display numbers starting from 1 and highlight the current page; the paging index starts from 0, and invalid indexes are disabled.

When passed in, the Button inherits size, text, icon, rich content, name, appearance, loading, and self-disable configuration. The component takes over the ID and click callback, does not execute the original button callback, and does not modify the original instance. Configuration is read every frame, so loading/disabling and appearance can be updated dynamically; custom pagination appearance configures itself, and the selected state is still exposed with toggle/selected semantics. Clicking the current page does not repeat the callback, and the program SetValue does not trigger the callback.

The carousel's own disabling will be synchronized to all associated controls; the parent container's disabling is only inherited along the actual layout tree. When combining, content and controls should be placed in the same disabled container. The automatic playback pause detection range is still the content area, and the hover of external controls does not automatically pause. Control button's Space/Enter performs actions; directional navigation is handled by content area focus.

## Multiple items on the same screen

`ItemsPerView(2)` Divides the current axis into two equal-sized slots; `Gap(8)` sets the item spacing (default SpaceMd). The horizontal allocation is based on available width, and the vertical allocation is based on Height; remeasure after the window size changes. Positive integers take effect, 1 restores the original single mode; the spacing only accepts limited non-negative values, and will be compressed in extremely narrow viewports. Free space is reserved when there are fewer items than the number of slots.

Buttons, paging, arrow keys, and autoplay still move by one item. In a limited track, the selected items are aligned as much as possible at the beginning, and the offset at the end is limited to fill the viewport, so the last items may share the same offset; Value and callbacks still distinguish each index. Each item in the cycle orbit has an independent snapping point, and the last item is followed by the first item. SetValue still updates the display when disabled.

By default, all projects use stable track identities, and content beyond the viewport is clipped and not virtualized. Track offset is controlled by selection or pointer dragging; continuous scrolling synchronizes selections after a pause.

## Fraction Proportions and Mixed Dimensions

`Basis(2.0/3)` makes each slot occupy approximately two-thirds of the viewport, exposing part of the item behind it; `ItemBasis(1, 0.5)` makes the second item alone half the viewport. The scale range is `(0,1]`, the scale does not exceed the viewport, and items beyond the viewport can use ItemSize; 0 clears the corresponding override, and illegal scales or indexes are ignored. The horizontal proportion corresponds to the width, and the vertical proportion corresponds to the height.

```go
car.Basis(2.0/3).ItemBasis(1, 0.5).Gap(8)
car.ItemBasis(1, 0) // Restore default ratio
car.ItemsPerView(2) // Clear global Basis, item-by-item override remains
```

Scale includes spacing share: project main axis size is `(viewport size + gap) × basis − gap`; e.g. 240dp viewport, 8dp spacing, 1/2 scale results in a 116dp project. Effective spacing converges to a minimum to fit narrow viewports. Navigation is accumulated based on the actual rounded pixel size of each item and spacing, and items are not assumed to be of equal width/height; end offsets are still limited by the viewport boundaries.

Adjusting the scale does not change the current index and does not trigger OnChange; the input state and focus are retained internally in track mode. Turning off Draggable and Scrollable, clearing all item-by-item overrides, and restoring the old single-item mode when ItemsPerView is 1 and Basis is 0 will rebuild the item hierarchy. Fixed dp size configured via ItemSize.

## Fixed size items

`ItemSize(index, dp)` Sets the main axis dimensions for the specified item, excluding spacing. The horizontal direction corresponds to the width, and the vertical direction corresponds to the height; the dp value remains unchanged when the window is scaled or the direction is switched. Fixed size takes precedence over ItemBasis and global scale, `ItemSize(index, 0)` clears the fixed value and restores the original scale. Invalid indexes, negative numbers, and non-finite values are ignored; modifications do not change the selection and do not trigger OnChange.

```go
car.Basis(0.5).ItemSize(0, 180).ItemSize(2, 640)
car.ItemSize(0, 0) // Restore half viewport
```

Fixed dimensions can exceed the viewport. When an oversized item is selected, the beginning of it is shown and the rest is cropped; when navigating to the next item, the offset is calculated as the full size of the oversized item plus spacing. This batch does not add free scrolling within the project. Oversized content requires the application to provide its own inner scroll container; continuous scrolling will be snaped after a pause, see below. The old single item mode is restored when Draggable and Scrollable are turned off, all itemized proportions/fixed sizes are cleared, and ItemsPerView is 1 and Basis is 0.

## Pointer drag snapping

`Draggable(true)` is enabled by default. Mouse or touch dragging can be initiated from the non-interactive area of the content area; the operating area for sliders, input boxes, clickable children, and inline widgets is reserved for child components. After moving at least 4dp along the current axis, the track follows the pointer, selecting the closest snap point when released. When the limited track starts dragging outward from the beginning and end, or the initial direction is closer to another axis, the carousel cancels its own drag and allows the outer layer to take over; the drag that has been taken over along the carousel axis will be retained until released and will not be handed over when it reaches the boundary halfway. Loop track dragging along the axis is still taken over by the carousel. The Value will not be changed during the move, and OnChange will not be called; it will only be called back once after the actual change of the index after release. For overlapping snapping points, the current item will be retained first, otherwise the first nearest item will be selected. Snapping uses a 180ms transition with no velocity inertia; it completes immediately when the reduced motion is turned on.

Canceling a drag, procedural SetValue, changing axis/dimension configuration, or disabling will restore the currently selected positioning without committing the old drag. Automatic playback is paused during dragging and restarts after completion. Draggable(false) disables dragging; buttons, paging, and keyboard can still be used. Sub-buttons, input boxes and other interactive controls prioritize their own events and do not initiate carousel dragging from these controls.

If the limited track (`Loop(false)`, or the track is not long enough to loop) is dragged to the beginning and end, the offset will be limited; the looping track can be dragged across the beginning and end. For the conditions, see "Cycling Track" below. See below for boundary routing of scroll events.

## Scroll wheel and continuous scrolling

`Scrollable(true)` is enabled by default. The horizontal carousel receives horizontal scrolling, the vertical carousel receives vertical scrolling, and the other axis is given to the outer layer. Continuous displacement updates the track in real time. After the gesture ends, the nearest snapping point is selected, and the actual change in the index triggers a callback. Pause autoplay during scrolling and restart when finished. Program switches, configuration changes, or disabling cancels old scrolling and does not allow its deferred callback to overwrite the new selection.

Gesture end and device determined by platform (`core.CurrentScrollGesture`):

- **macOS**: Keel reads native scroll events before Gio. The touchpad is immediately snaped when the finger is lifted, and will not be snaped in advance when the finger is stopped. The system inertia after lifting is ignored. The mouse wheel switches one item per grid along the carousel direction.
- **Linux Wayland**: Keel opens another pointer object on the window's own connection, reads the scroll source (wheel/finger) and finger lift event, and behaves the same as macOS; Wayland does not send system inertia, so there is no inertia to ignore. When the synthesizer's wl_seat is lower than version 5, you cannot get this information and return to the next one. This part has only been compiled and checked with real header files on macOS and has not yet been run on the Wayland desktop.
- **Windows, X11 and other platforms**: The system's wheel message does not have device and gesture stages. If there is no new event for 140ms, it is deemed to be over.

`WheelStep(true)` switches item by item on any device according to events, and the horizontal carousel also receives vertical scrolling; when it is not turned on, it switches item by item only when the mouse wheel is detected, and only responds to the direction of the carousel itself and does not grab the vertical scrolling of the outer page.

The continuous pattern of finite orbits hands events to the outer layer when rolling outward from the beginning to the end; scrolling that has started inside is retained until it stops, preventing the outer layer from suddenly moving when it reaches the boundary. The item-by-item mode continues to use Loop; the non-loop boundary is given to the outer layer, and the page is changed at the beginning and end during the loop. The scroll container in the interactive child receives its own scroll first. `Scrollable(false)` closes this input and does not close pointer dragging.

`Loop(true)` allows dragging and continuous scrolling across the beginning and end of the track when it is long enough. Each project is only mounted once, and the original project is moved to the adjacent cycle according to the current position; the input status, click area and overlay anchor point are retained with the project. The loop condition is "cycle length including the end spacing − maximum item length ≥ viewport length", otherwise it returns to the limited track and the navigation still loops according to the index. This restriction prevents the same interactive item from having to appear on both sides of the viewport at the same time. Changing the loop, orientation, or size configuration cancels the current gesture.

Sustains the current display offset when starting dragging from scroll. The native touchpad feel, system inertia and event merging of different devices still need to be accepted by the real device.

## Switching and snapping animation

Buttons, pagination, keyboard, and autoplay use 180ms ease-out transitions; Value and OnChange update immediately when the selection changes, without waiting for the animation to end. The next item of the circular orbit is always forward, and the previous item is always backward. After crossing the beginning and end, it is reset to the equivalent orbit coordinates without flashing back to the other end.

After dragging is released, canceled or scrolling is paused, the current offset transitions to the snapping point; returning to the original index will also snap, but the callback will not be repeated. Navigating again during the animation starts a new transition from the last displayed position; starting a drag or continuous scroll takes over that position immediately.

First display, SetValue (including repeatedly setting the same value), size or orientation changes, disabling and reducing animations all target the selected item directly. Turning off dragging and scrolling still retains the navigation animation.
