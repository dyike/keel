# Combobox

English | [简体中文](combobox.zh-CN.md)

Inputable and filterable drop-down box.

```go
customer := kit.Combobox("Customer", customers...).Placeholder("Type to filter")
tags := kit.Combobox("Tags", "Urgent", "VIP").AllowCustom()
```

- Open the list and filter when typing (not case sensitive, matches if included), click the option to select.
- Rules for carriage return:
  - If the entered text happens to be a certain option, select it;
  - When `AllowCustom` is set, the entered text is retained;
  - Otherwise the first match is chosen.
- When `AllowCustom` is not set, if the text is not an option after leaving the input box, it will revert to the last selection.
- ↓ Open the list and move the highlight, ↑ move back, and press Enter to select the highlighted item.
- `Value()` / `SetValue`, `SetOptions`, `SetDisabled`, `SetError`. `el.Root` REQUIRED.

Agent: Container roles `combobox`, `value` are currently selected; there are `textbox` and expand buttons; the lists are `listbox` and `option`.

Verify: `go run ./examples/components -section combobox`, add `-theme dark` to check the dark theme.

`Multiple()` displays the options as removable labels, leaving the input area for the next search; `AllowCustom()` can be combined with it. `Values()` returns a copy of the selection sequence, `SetValues` replaces the selection after deduplication and does not trigger a callback; `OnValuesChange` receives a copy of the user's additions and deletions. `Values()` should be used for multiple selections, `Value()` retains the most recent leading value. `SetValue` Replaced with a single selection, procedural assignment closes the results list and invalidates the old query. Constructed with `SetOptions` to copy option slices; updating search results without removing selected tags.

`OnSearch(func(query string, token uint64))` takes over the result source. There is a new token for each input, open or retry; call `SetResults(token, options...)` or `SetSearchError(token, message)` with `core.Update` after asynchronous completion. Expired and closed results return false, and remote results are not subject to secondary local filtering. Old results cannot be executed during loading or failure, and retries are provided after failure. Closing, programmatic assignment, or disabling will invalidate the request; defocusing due to disabling will not submit the draft.

Results are cached using virtual lists and filters. ↑ ↓, PageUp/PageDown will scroll to reveal the highlighted items, and the height of the result area is limited by the window. The example includes tens of thousands of customers, free input of multiple tags, and asynchronous multiple selection that can simulate failure.


`DisableOption(value, true)` prohibits the user from selecting the specified value. Candidates are still displayed with disabled semantics. Click, up and down arrow keys, PageUp/PageDown, Enter, and out-of-focus submissions all respect disabling; AllowCustom cannot bypass disabling with the same value. When all candidates are disabled, there are no highlights and candidates cannot be submitted. Pass false to restore optional.

Configurations are saved as string values, filtering, SetOptions and asynchronous SetResults will not clear the configuration; duplicate value candidates are disabled together, and values that have not yet appeared in the candidates can also be pre-configured. Disabling selected values will not automatically delete or trigger callbacks, and multi-select labels can still be removed. SetValue/SetValues retains the ability of the application to actively assign values. When the current highlighted item is dynamically disabled, it is changed to the next available item, and when all are disabled, the highlight is cleared.


In the multi-select list, clicking again or pressing Enter to confirm the selected candidate will deselect the candidate and keep the pop-up layer open; the search text is cleared, and asynchronous mode applies for a new token for the cleared query. OnValuesChange is triggered once for each switch; Value falls back to the last item of the remaining selection when the most recent primary value is removed, or is empty if there is no remaining selection. OnChange is only triggered when the main value changes. Removing other selected options will not repeatedly notify the main value. Disabled candidates cannot be toggled from the list, and selected tags can still be removed using the Remove button.


`Footer(view)` displays the persistent operation area below the candidate scroll area, and is also retained during loading, failure, and empty results; `Footer(nil)` is removed. The content can contain buttons or input boxes. Candidate updates retain the focus and state of the same content instance; the footer operation does not select the candidate and does not automatically close. The application can call SetValue/SetValues to apply the new selection and close. Disabling a trigger field or ancestor closes the entire popup layer.

The footer occupies at most one-third of the height of the window after deducting 80dp. Super-high content scrolls in its own area; the candidate area reserves height for the footer. The first frame is reserved according to the upper limit, and the next frame converges to the actual height after measurement. The example's "Add Example Tag" button demonstrates active update selection.


`Clearable(true)` displays an independent clear button when there is a selected value, and is available for both single and multiple selections. Clearing by the user will close the pop-up layer, cancel old queries, clear drafts and errors, and return the focus to the input box; OnChange will receive an empty string, and multi-select OnValuesChange will receive an empty slice, and each will be notified once. The result of re-SetValue in the callback is retained; the callback parameters describe this clearing event. Don't show the button when only search drafts are left without a selection, disabled state follows fields and ancestors; `Clearable(false)` hides the button. The procedure SetValue/SetValues continues without triggering callbacks.


`Size(dp)` sets the minimum height of the field and synchronously scales the font size, spacing, expand/clear buttons, multi-select labels and candidate rows. 28/36/48dp is recommended; 0 returns to default, negative numbers and non-finite values are ignored. Multi-select content can be wrapped, the actual height may be higher, and the label retains a minimum height of 16dp. footer maintains its own size. Modifying the size during opening retains input focus and re-exposes the highlighted candidate.

`CheckIcon(kit.Icon(kit.IconCheck))` Replaces the icon of the selected candidate, also accepts VectorIcon; saves a copy of the icon configuration, with color and base size from the incoming icon, scaled with Size. Unselected candidates retain the same width space. `CheckIcon(nil)` Restore default; pass IconNone to hide the pattern and retain the size.


`SetItems(...ComboboxItem)` accepts `Value`, `Label`, `Disabled`, separating the stable value from the display name. Empty Value is ignored, empty Label uses Value, repeated Value retains the first item, and the input slice will be copied. Local searches match both names and values; callbacks, Value/Values, and program assignments use stable values, and candidate and multi-select labels display names. The name of the candidate semantic is the display name and the value is the stable value.

`SetItemResults(token, items...)` is used for structured asynchronous results. It verifies the request status like SetResults and does not perform local secondary filtering. Replacing the candidate will not delete the selection; the selected value will not retain the original name in the new result, and the name will be updated after returning the candidate. SetOptions/SetResults Restore string mode. The item Disabled and DisableOption are logically ORed. To modify the item status, SetItems/SetItemResults must be re-set.

When submitting input text, the exact value takes precedence, followed by the first name match; different values with the same name are recommended to be selected from the candidates. Unmodified radio selection display text retains its original value when out of focus, to avoid accidental selection if the name happens to be equal to the value of another item. Drafts being edited will not be overwritten by name updates.


`RenderItem(func(ComboboxItem, bool) el.View)` Customize the candidate text, the parameters include stable value, display name, merged disabled state and whether to select. The callback is nil, or the default name is restored when nil is returned. Row selection semantics, disabled, background and trailing ticks are still maintained by the component. Independent buttons within the text will not select the entire line; clicking on ordinary text or blank spaces will still select candidates. Disabling a row also disables its child operations.

Only visible virtual rows call rendering functions. When you need to retain the input/button state, press Value to reuse the View instance to avoid creating a new state object every frame; candidate rearrangement uses value identity, and repeated strings with the same value are distinguished according to the order of appearance. Focus is not guaranteed to be retained after leaving the virtual viewport, and persistent input values should be placed in app state.

`RowHeight(dp)` sets the unified candidate line height with top and bottom spacing. Positive values are independent of Size. 0 restores 30dp scaled with Size; illegal values are ignored. Minimum height is 1dp text plus current top and bottom spacing. Rich content will not automatically measure different line-by-line heights, and the application should set sufficient line height; content that exceeds the line height will be clipped by the virtual viewport.


`SetGroups(...ComboboxGroup)` sets ordered groups, each group contains stable ID, Label and Items; empty IDs are ignored, empty Labels use IDs, repeating group IDs and cross-group repeating candidate values retain the first item. Enter slice copy and choose not to clear. The group title is only displayed and does not participate in the selection; groups without candidates and groups with no matching items after searching are automatically hidden. The search matches candidate names and values, not group titles.

`SetGroupResults(token, groups...)` submits the asynchronous grouping result and still verifies the request token; SetItems/SetOptions and its asynchronous entry return to non-grouping mode. The group title and candidates are both in the virtual list and use the same RowHeight; the keyboard skips the title and disabled items, and the scroll positioning includes the space occupied by the title. The title scrolls with the content and is not fixed at the top. Custom RenderItem only handles candidate body text.


`Searchable(false)` Replaces the input box with a selection-only button, defaults to Searchable(true). Click or Enter to open, arrow keys to navigate, Enter/Space to confirm the highlighted candidate, and Esc to close; multi-select can still switch candidates and remove tags. Entering text does not change the query, candidates are not filtered locally, and AllowCustom does not take effect in this mode.

Switching the mode closes the pop-up layer, discards the draft and invalidates the old asynchronous request, retaining the selected value; repeatedly setting the same mode does not change the state. When opening or retrying without search mode, OnSearch receives an empty query, and the returned results are still subject to token verification; the input box is redisplayed after restoring Searchable(true). Clear buttons, footers, custom candidates, and grouping continue to be available.


`OnConfirm(func([]string))` Notifies once when the user ends an open selection session: applies to single value selection, Esc, external click, expand button close, and clear while open. Multi-select item-by-item switching only triggers the change callback and is confirmed when it is closed; clearing the pop-up layer that is not open does not trigger confirmation. Confirm that after this OnValuesChange/OnChange, the parameters are the selected snapshots at the time of completion, and modifying the slices does not affect the components.

Confirmation does not necessarily change the selection, nor does it change the existing Esc/External Close draft submission rules. Programmatic assignment, disabling, removal and Searchable switching do not trigger confirmation. The value can be reassigned in the callback, and the original operation will not overwrite the new value after the callback; a new search will not be started when the multi-select change callback closes or disables the component. `OnConfirm(nil)` removes the listener.

`RenderTrigger(func(ComboboxTriggerContext) el.View)` replaces the entire default field appearance; nil callback or nil content restores default. The context provides a copy of Selection (value, name, candidate disabled state), Open, the component itself Disabled, Size, Placeholder, and Toggle/Clear that can be called in UI events. Custom triggers draw borders, labels, and clear entries by themselves; default multi-select labels, expand icons, and Clearable buttons are no longer automatically inserted. The component still provides label/error copy, selection semantics, focus entry and keyboard behavior, and ancestor disabling is performed uniformly by the outer layer.

Ordinary display content can be turned on and off by clicking on the background; custom buttons can be bound to contextual actions, and sub-buttons do not trigger the background. Actions are only used in events and cannot be called during rendering. Enter/Space/arrow keys can be opened from the trigger; when the search is turned on, the search box appears in the panel and is automatically focused. After selection, it returns to the trigger (return to the search box when the multi-select is still open). Navigate candidates directly when closing search. Enabling or removing a RenderTrigger for the first time cancels the draft and closes the old spring layer; replacing a non-nil renderer preserves the selection and expanded state. Reuse stateful Views to preserve child control state.
