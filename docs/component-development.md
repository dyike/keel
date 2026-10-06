# Component guidelines

English | [简体中文](component-development.zh-CN.md)

Comply with this specification when adding or modifying the `ui/kit` component. For the implementation skeleton, see [Extension Guide](extending.md#add-new-components), for appearance conventions, see [Component Visual Specification](visual-guidelines.md), for user usage and classification entry, see [Component Reference](kit.md).

## Modules and APIs

The enumeration constants of kit are all prefixed with the type name, and the Name / Shape / Status / Variant suffix in the type name is removed: ToneNeutral / ToneInfo / ToneSuccess / ToneWarning / ToneDanger, AvatarOnline / AvatarBusy / AvatarOffline, ButtonPrimary / ButtonSecondary / ButtonGhost / ButtonDanger; IconCheck and MarkerDot continue to use the existing naming. Compatible aliases for old names are not retained. el is the underlying layout library, and layout short names such as el.Bottom and el.Start are not restricted by this rule.

kit only directly depends on Keel's `core`, `theme`, `locale`, `el`, `base`; it does not reference `window`. The basic capabilities that el lacks are implemented in el first, without duplicating Gio input routing, timing or overlay mechanisms in each component. Dependency tests register `internal/loop`, `internal/editorstyle` and `internal/inputcontent` as transitive dependencies, which are not direct dependencies of kit.

Constructor `Xxx(...)` returns `*XxxView`. The component is connected to el with `Render(*el.Context) el.Element`, the instance retains the business state, and Render generates an element tree based on the current state. Dynamic lists use stable IDs instead of array positions to represent movable items.

Valued interactive components provide `Value()`, `SetValue(...)`, chained `OnChange(...)`, `SetDisabled(bool)`. Programmatic assignment does not trigger callbacks, only user operations. Action components such as buttons use click callbacks and `SetDisabled`. Pure display components do not impose value and callback interfaces. Internal storage cannot be exposed when returning mutable values such as slices.

## State and layout

Loading and busy states do not change focus or tab order, only activation is ignored.

The callback changes the state within the frame lock; the background task returns to the UI through `core.Update`. The component cannot reset cross-frame state based on `gtx.Enabled()` being false: it may also indicate that there is no input source for el measurements or off-screen layout. Resets such as closing the menu and stopping dragging are placed in `SetDisabled(true)`; events, appearance and semantics respect the parent's disabled state.

The display, hiding, and content changes of decorations should not change the host size or baseline. For example, changes in Badge count should not make buttons jump. Ordinary content changes may result in reformatting. Constraints are adhered to in narrow containers; text checks for Chinese, Latin letters, numbers, mixed layout, 1×/2× scaling, the input cursor and selection area use unified glyph measurements, and no offset is forced according to a certain screenshot.

Only the scale of `theme` (`scale.go`) is used for rounded corners, font size, and shadow, and no numbers are written: the same character looks the same in each component, and only the scale is changed when the design is changed. Use `surface()` for cards, and `floating(level)` for layers floating on the page (menus, pop-ups, drop-downs, dialog boxes, notifications), with shadows.

All text fields (the hexadecimal boxes of Input, NumberInput, Select, Combobox, TimeField, DatePicker, InputGroup, ColorPicker, and the search boxes of Select, Command, and Settings) use the same `fieldFrame` in `ui/kit/field.go`. The outer frame is: high `theme.ControlHeight` (36dp), left and right inner margins 10dp, rounded corners. 6dp, select "Error → Focus → Normal" color for the border, use Subtle background color for disabled and read-only. The search box uses `searchField` with a search icon in front. The outer box surrounding the text input calls `FocusOnPress`. Clicking any blank space in the outer box will focus on the text. To change the appearance of the field, only change this part, do not write additional borders and padding in the component; `TestFieldsShareControlHeight` checks the equal height of each field.

The framework's own text (button copy, accessibility name, placeholder text, count) is always read from `locale.Current()` during Render, and is not saved during construction, nor hard-coded Chinese; the test of `internal/deps` will block hard-coded Chinese strings. Use `locale.Current().Name(action, target)` when concatenating names of the form "action + object". The text passed in by the application (title, menu item) is used as is.

The color is read from the theme semantic color when Rendering, and the theme snapshot is not saved in the constructor. Custom fixed colors are explicit overrides; theme switching does not infer their meaning for the app. The cache must contain the theme version, or use `cx.Cache` which supports theme invalidation. Local themes use `cx.Themed`, see [Elements and Views](el.md#local-topic).

Interactive components support mouse and keyboard; when disabled, they cannot be activated or receive focus, and Agent snapshots report disabled. Focus, keys, timing and overlays use API](el.md) provided by [el.

The overlay component must first create the panel and call `cx.Overlay`, then render the content of the Body / Footer and append it to the panel. The content itself may register the sub-overlay; the reverse order will cause the sub-menu to register earlier than the parent layer and be closed because the anchor is not yet available. Modal components use `Layer.Owner` to bind their own elements and inherit the outer layer's disabling and hiding. The acceptance must include true nested opening, layer-by-layer Esc, focus return and ancestor disabling, and cannot only test independent elastic layers.

## Files and validation

One component corresponds to `ui/kit/<name>.go`, `<name>_test.go`, `docs/kit/<name>.md`, `examples/components/<name>.go`. Register in the corresponding category of [component reference](kit.md); the site sidebar is generated from this index. The example registers a standalone `-section <name>`, showing common states, borders, and light-dark colors. The site automatically reads the demo code from the registration file; do not copy the entire example in the component documentation.

Component documentation includes purpose, minimal usage, public API, keyboard operations (when applicable), semantics, boundaries, and validation entries; tests verify user-observable behavior and do not replicate implementation algorithms.

`ui/kit/conventions_test.go` automatically checks the parts that can be judged by machines: each component has documentation, a sample section with the same name, and Agent tests in `ui/window`; enumeration constants are prefixed with type names; the public APIs of kit and el have no compatible entries or aliases. The new Agent role does not need to be registered in the automation code. Only container roles that need to list child elements separately are added `containerRoles`.

Check each item before submission:

- `uitest` drives layout and interaction; purely presentational components check for size, constraints, state, and color.
- `ui/window` Agent snapshot covers name, role, value and status; new role synchronizes `automation.go` and `docs/automation.md`.
- When there is interaction, overwriting the keyboard, disabling, restoring, and programmatic assignment will not trigger callbacks; after embedding el, multiple consecutive frames will not lose the state.
- The example works, the light and dark colors and text layout are checked with screenshots.
- The public API, component documentation, README, and examples are updated at the same time.
- `go build ./... && go vet ./ui/... && go test ./... -count=1` All passed, including `cmd/keel-mcp` end-to-end testing.

A change only addresses one component or one common capability, the verification requirements are in [TEST](testing.md).

## Documentation

- The complete getting started example is placed in [Quick Start](getting-started.md), each module README records responsibilities, dependencies and implementation files.
- The component classification is only maintained [component reference](kit.md), and new components are supplemented with documentation, examples and Agent tests.
- API change synchronization usage documentation; globally affecting trade-offs are documented in [Design Decisions](decisions.md).
- Development progress and historical acceptance results are placed in `docs/reports/`, and document navigation is not added.
- After modifying the index, run `go test ./internal/site` to check category coverage and intra-site links.
