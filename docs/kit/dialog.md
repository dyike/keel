# Dialog

English | [简体中文](dialog.zh-CN.md)

Modal dialog box. When opened, the page dims, does not respond to the pointer, and the focus is limited to the dialog box; after closed, the focus returns to the position before opening.

**Standard message**: Reuse the same instance.

```go
dlg := kit.Dialog("")
dlg.Confirm("Save changes", "Save before leaving?", save)                  // Cancel/OK, focus on OK
dlg.ConfirmDanger("Delete order", "This cannot be undone.", "Delete", remove)  // Focus on cancellation
dlg.Alert("Export complete", "36 records in total.", nil)                      // Only sure
```

**Custom content**:

```go
edit := kit.Dialog("Edit customer").Body(form).Footer(cancelButton, saveButton).Width(480)
edit.SetValue(true)
```

- Choose one of two hanging methods: render `dlg.Render(cx)` in the view tree (no content is displayed in the original position, only the overlay is declared); or call `dlg.Show(cx)` in the callback, and the dialog box is directly hung to the root of the window without being placed in the view tree. It will be automatically removed after closing, just like GPUI's `window.open_dialog`. Do not use both in the same instance. `el.Root` REQUIRED.
- Close method: Esc, click mask, cancel button. `OnClose(fn)` is called on shutdown; the standard message "OK" does not count as a shutdown and just runs its own callback.
- The dialog boxes of `ConfirmDanger` and `Persistent()` will not be closed when clicked on the mask to prevent accidental touches; Esc is equivalent to "cancel".
- The button of the custom dialog box is provided by the caller, and the button callback calls `SetValue(false)` to close the dialog box.
- `Value()` / `SetValue(bool)` reads or sets whether to open; `SetTitle` modifies the title; `Width(dp)` sets the width, the default is 420, and does not exceed the window width in a narrow window.

Agent: The role of the ordinary dialog box is `dialog`, the role of `ConfirmDanger` and `Persistent()` is `alertdialog`, the name is the title, and the elements inside are listed separately. While the dialog box is open, the following pages are not visible in the snapshot.

Verification: `go run ./examples/components -section dialog`.

The dialog box is limited to the window, the long text is scrolled separately, the footer buttons of the narrow window are changed to be arranged vertically; the extremely small window allows the entire panel to be scrolled. Non-limited widths are ignored and Footer saves a copy of the view list.

`SetDisabled(true)` is closed directly and prevents SetValue/standard messages from reopening; the user close callback is not triggered. When the containing element is disabled, hidden, or no longer available, the declared modal layer requests closure and calls OnClose once. After closing, it will not pop up again due to resumption.

The body and footer can contain a Menu, Popover, or another Dialog. The parent overlay is registered first, and the child overlay is above; Esc closes each layer layer by layer, and each layer restores the corresponding previous focus. The custom preset in the example "Custom" is used to validate nested menus.

The shutdown configuration can be set independently and retained when reusing standard messages:

- `Keyboard(false)` disables Esc to close and still consumes this key to avoid accidentally closing the lower dialog box; it is enabled by default.
- `Overlay(false)` Hides the mask color, still blocking background operations and constraining focus; shown by default.
- `OverlayClosable(bool)` Explicitly sets whether external clicks are turned off, taking precedence over Persistent/ConfirmDanger's default value.
- `CloseButton(true)` Displays the title bar close button, calling the same close logic as Cancel/Esc. Hidden by default, retaining the old layout; button names switch with the language. Can be displayed without title.

Close buttons, headers, body text, and footers use stable identities; switching display configurations does not rebuild the body input state. The above configuration does not restrict the program from calling `SetValue(false)`, nor does it restrict the custom Footer button.

`BeforeConfirm(func() bool)` is run before the confirm operation of the standard Confirm / ConfirmDanger / Alert. Return false to keep it open, retain focus, and not execute the original onOK, which is suitable for synchronization verification or waiting for background tasks; after returning true, close first and then execute onOK in the original order, without calling OnClose. nil clears the check and retains the configuration when reusing standard messages. Cancel/Esc/Mask/Close buttons are not affected by this validation; custom footers are still controlled by the app.

```go
dlg.BeforeConfirm(func() bool { return formIsValid() })
dlg.Confirm("Submit", "Submit now?", submit)
```

If SetValue, SetDisabled or another standard message is opened in the verification callback, the old confirmation operation will not continue to be closed or onOK will be executed. The verification will not automatically start the goroutine, nor will it automatically display the loading state.

`BeforeCancel(func() bool)` runs before Esc, mask, cancel button and title close button execute. false maintains the modal and focus, does not call OnClose, and the user can try again; true continues the original closing process. nil is cleared and retained when standard messages are reused. SetValue/disable/replace messages within callbacks abort old cancellation operations.

The program calls SetValue(false) and SetDisabled(true) without cancellation verification. The overlay cleaning when the belonging element is hidden, disabled or removed also bypasses the validation and uses the OnClose notification; the validation cannot allow the dialog box to continue to exist outside the view tree. The underlying `el.Layer.BeforeDismiss` provides the same user-closed verification entry.

## Title icon

`Icon(kit.IconWarning)` Display the icon before the title and keep it when reusing the same dialog box to display different messages. Colors follow the tone: `ConfirmDanger` defaults to dangerous colors, others default to reminder colors, `IconTone(kit.ToneSuccess)` etc. can be specified. `Icon(kit.IconNone)` Remove the icon. The icon is just for decoration and does not broadcast separately. The title already explains the meaning.

```go
dlg.Icon(kit.IconWarning).ConfirmDanger("Delete order", "This cannot be undone.", "Delete", remove)
dlg.Icon(kit.IconCheck).IconTone(kit.ToneSuccess).Alert("Export complete", "36 records in total.", nil)
```
