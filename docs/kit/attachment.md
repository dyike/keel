# Attachment

English | [简体中文](attachment.zh-CN.md)

Attachment card: file name, size, upload progress or errors, can be opened and removed.

```go
a := kit.Attachment("报价单.pdf", size).OnRemove(remove).OnOpen(open)
a.SetProgress(0.6)          // Uploading; negative number means completed
a.SetError("超过 10 MB 上限")
```

- Size formatted with `kit.FileSize` as B/KB/MB/GB. A progress bar and "Uploading 60%" are displayed during the upload. When an error occurs, the reason is displayed in a dangerous color.
- The "Uploading" text comes from locale.

Agent: Role `attachment`, the name is the file name; `value` is empty, "Uploading 60%" or `error`; the remove button is named "Remove file name".

Verify: `go run ./examples/components -section attachment`, add `-theme dark` to check the dark theme.

`OnCancel(fn)` displays a cancel button during the upload; click to mark canceled first and then notify the business to stop uploading. `OnRetry(fn)` displays retry after error or cancellation. Click to clear the error and reset the progress to 0 before calling the business callback. The callback is responsible for starting/stopping the real transfer; the background task updates the component via `core.Update` and discards late results from canceled tasks.

`SetProgress` Clears previous error and cancellation status. Progress greater than 1 is truncated to 1, NaN/Inf is ignored, and negative numbers mark completion. Negative file sizes appear as 0B. Only completed and error-free attachments can be opened, cancellation and removal will not trigger the open callback. `SetDisabled(true)` prohibits all operations on the card. The accessible names of the Cancel and Retry buttons include the file name; the Agent value for the Cancel state is `canceled`.

`Media(view)` Replaces the default file icon with a presentational View; nil restores the icon. You can pass in `kit.Image(pixels, alt).Size(width, height).Fit(kit.ImageCover)` to display the image, or you can use MediaSource to load the image from the component. Horizontal media uses the content's own size, and the maximum width is limited by the card; horizontal previews should use small thumbnails to leave space for file names and operations. Vertical layout uses square preview by default, see below.

`Vertical(true)` places the media above the text and overlays the operation in the upper right corner of the card. `Vertical(false)` restores the default horizontal layout. Switching layouts during opening maintains the keyboard identity of the open area; preview does not trigger OnOpen when uploading/fails, and cancellation, retry, and removal remain independent. Media is used for presentation, leaving the opening interaction to the attachment OnOpen, avoiding nesting buttons or another clickable image within the preview. This interface automatically adds a media status mask; the title during uploading/processing displays text sweep, and the size file is shown below.

`AttachmentGroup(items ...el.View)` Arrange attachments into a horizontally scrollable row without compressing the card width. `Gap(dp)` sets the non-negative spacing, defaults to SpaceSm; `Name` sets the accessible name of the group. The group width fills the parent container, and the items are top aligned.

`SetItems` replaces the list, and `Items` returns a copy. The two isolate slice modifications and ignore nil entries; the attachment instance is still shared, maintaining its own upload status and callbacks. Do not render the same instance repeatedly in a group. `SetDisabled` prohibits operations within the group and does not modify the attachment's own disabling settings. Removal is completed by the application calling SetItems; the group is only responsible for arrangement and scrolling, and does not take over file selection or upload tasks.

```go
files := kit.AttachmentGroup(report, photo).Name("附件").Gap(12)
photo.OnRemove(func() { files.SetItems(report) })
```

`SetStatus(AttachmentStatus...)` sets the explicit life cycle, and `Status()` reads the current effective state. The default is Complete; available are Pending, Uploading, Processing, Failed, Complete, and Keel reserved Canceled. The status value provides the IsPending/IsUploading/IsProcessing/IsFailed/IsComplete/IsInProgress query, IsInProgress includes uploading and processing.

Pending displays "to be uploaded" and Processing displays "processing"; during uploading and processing, the default media icon is replaced with a spinning circle, which can be canceled. Failed If there is no error description, "Upload failed" is displayed and you can try again; only Complete can be opened. The text switches with the locale, the Agent adds pending/processing values, and the old upload, error, canceled and null values remain compatible.

Explicit SetStatus clears error and cancellation status; entering Uploading retains valid progress or starts from 0, other status clears progress. SetProgress Non-negative values enter Uploading (including 1), negative values enter Complete; explicitly set Processing when processing is required after completing the transmission. SetError temporarily overwrites the current status and clears the error to restore the previous status; explicit Failed requires SetStatus or retry to exit. The canceled status continues to take precedence over error display. All program status updates do not call operation callbacks. When retrying, enter 0% upload before notifying the application.

`Content(view)` replaces the default file name, status description and progress bar; nil restores defaults. When OnOpen is enabled, the display content should be used here, and the interactive control should be placed in `Actions(views...)`. The custom meta-information reads Status by itself and displays the desired status, and the Agent name and lifecycle value of the card itself remain unchanged.

`Actions` Copies the incoming slice, ignores nil, and adds control before built-in cancel/retry. The built-in remove button is placed independently in the upper right corner of the card. Empty parameters clear custom controls; clearing OnCancel/OnRetry/OnRemove callbacks can remove the corresponding built-in buttons. Custom actions do not trigger OnOpen and respect card and ancestor disabled states.

`PartStyle(part, func(*el.DivEl))` adjusts the six partitions of Root, Media, Content, Title, Description, and Actions, and the constants all start with AttachmentPart. You can set the background, border, rounded corners, spacing, font size and color, and you can also use Hidden to hide the optional area. The style is applied after the default value, nil restores the default; the element is rebuilt every frame and should not save references or add child content in the style callback. The Root's ID, role, name, lifetime value, and window maximum width are maintained by the component. Title/Description only works on the default meta information, and the application controls the internal style when the Content is customized.

`Size(AttachmentSize...)` Select four levels: XSmall, Small, Medium and Large, the default is Medium. The card width is 176/200/232/272dp respectively, the default media side length is 28/32/38/44dp, the title font size is 11/12/13/14sp; the minimum height is 40/48/56/64dp, and will continue to increase when there is more content. Padding, spacing, and built-in action buttons adjust with gears.

The default Medium width has been adjusted from the original 280dp to 232dp. `PartStyle` is applied after size defaults, overriding width and media sizes; sizes explicitly set in custom Media/Content/Actions remain unchanged. Vertical default square preview and upper right corner operation; see MediaAspectRatio for custom ratio.

Default state appearance: Pending uses a dashed border, Failed uses a DangerText border; returns to normal borders when completed. When no custom Media is provided, a dangerous color background and icon are displayed on failure: an error icon is used for OnRetry, and a prohibition icon is used for OnRetry without. When custom media is uploaded, it is covered with a dark mask and a white progress ring, and an indeterminate progress ring is displayed during processing; when it fails, the mask is deepened, and OnRetry displays a circular retry button, otherwise a forbidden icon is displayed. The completed, pending upload and canceled statuses restore the original preview. Masking does not change the media size, and cropping follows the Media partition rounding. Media retry and operation area retry share status verification. Enter 0% upload first and then notify the application. Cards and ancestors are disabled; only the completion status can be opened.

Partition styles are applied after state defaults, color/width can be overridden via Root's `Border`, `BorderDashed(false)` restores solid lines; Media can override background. The underlying el's BorderDashed also applies to other elements and state styles, maintaining the original border width and rounded corners, and the dashed line is 4dp solid segment and 3dp spacing.

`MediaOverlay(view)` overlays the custom View in the center of the media area, draws it above the preview and built-in life cycle mask, and does not participate in media size calculation; nil removes it. You can place a play button, logo or custom progress. The content should fit the media size, and the excess parts should be cropped according to the media boundaries and rounded corners. If there is no custom Media, it can also be superimposed on the default icon.

Overlay buttons have independent click and keyboard focus, do not trigger attachment OnOpen, respect attachment and ancestor disabling. Completed attachments can still be opened in empty spaces in display content or overlays; attachments will not be opened during upload, processing, and failure. State switching preserves the overlay identity, and the app can decide what to display by pressing Status in the ViewFunc.

```go
a.MediaOverlay(kit.Button("播放", play).Size(24))
a.MediaOverlay(nil) // clear overlay
```

The default title displays ShimmerText sweep during uploading and processing, and returns to normal text in other states or when reduced motion; retains the font size, font weight, and line height inherited by the Title partition. Custom Content replaces the default title and can be combined with kit.ShimmerText if needed.

`PartStatus(part, status)` sets the independent display status for the default Title or Description; `ClearPartStatus(part)` restores the current valid status of inherited attachments. Illegal status and other partitions are ignored. The upload/processing status of Title controls scanning; the status of Description controls automatic text and failed color matching. The override persists after the parent state is updated, without modifying the attachment's own state, media, progress bar, open/cancel/retry logic, or Agent lifecycle values.

`Description(text)` replaces the default description text while retaining the color of the active description state; an empty string displays empty copy. `ClearDescription()` Restore automatic size/status text. Displays 0% when the override is Uploading but the attachment has no valid upload progress. PartStyle is still applied after state color matching. Custom Content replaces the entire default meta information. At this time, the title/description configuration is not displayed for the time being; it will be restored after clearing Content.

```go
a.SetError("当前版本上传失败")
a.Description("上一版本已上传").
    PartStatus(kit.AttachmentPartDescription, kit.AttachmentStatusComplete)
// The card is still in a failed state and the description is in normal color; retry operations are still available.
a.ClearPartStatus(kit.AttachmentPartDescription).ClearDescription()
```

`MediaAspectRatio(width/height)` configures the vertical preview ratio, the default is 1; for example, 2 means the width is twice the height. 0 restores the media to its natural size, negative and non-finite values are ignored. The vertical preview fills the inner width of the card. The loaded kit.Image is cropped in the center and covers the preview by default, without modifying the original Image instance; other custom Views maintain their own size and are centered. The explicit height of PartStyle(Media) takes precedence over the scale, which can also be overridden with AspectRatio.

The vertical operation area is superimposed on the upper right corner of the card, and the default offset changes with the padding of the size file; after customizing the Root padding, you can use PartStyle(Actions) to adjust Top/Right. The operation background is Surface, the buttons are disabled and do not trigger opening, and the button focus is retained when switching between horizontal and vertical directions. The operation area shrinks according to the content width. Too many operations may block the preview. The application should limit the number or use PartStyle(Actions) to configure line wrapping.

Layout changes: The original vertical natural preview and bottom operations are changed to the above default values. Set MediaAspectRatio(0) when you need to preserve the natural preview. Standalone example: `go run ./examples/components -section attachment_vertical`.

`AspectRatio(ratio)` of the underlying el derives height by aspect ratio when width is parsed and height is automatic; explicit height and max/min height constraints take precedence, 0 clears the scale. It does not automatically derive dimensions for content in both axes.

`ShowMedia`, `ShowContent`, `ShowActions` independently control the media, meta information and operation area, all displayed by default. After hiding, it does not occupy the layout, does not appear in the Agent element, and cannot be focused/operated; it will be displayed again using the original content, style, callback and status configuration. The filename and lifecycle semantics of attachments are always preserved. ShowActions controls custom and built-in buttons in the action area. The media mask's retry button and MediaOverlay are still controlled by ShowMedia.

When arranged vertically with media displayed and meta information hidden, the card becomes a pure image tile: the padding and minimum height are removed, the preview fills the inside of the border, and the default inner rounded corner is 1dp smaller than the card. The default is still square, MediaAspectRatio can change the ratio. After customizing the border/rounded corners of Root, you can use PartStyle(Media) to synchronize the inner rounded corners. When there is only an operation area, a normal flow layout is used to avoid operations hanging on an empty preview; when all three areas are hidden, the card shell and semantics are retained.

```go
photo.Vertical(true).ShowContent(false) // Pure picture card
file.ShowMedia(false)                  // Only meta information and operations
file.ShowMedia(false).ShowContent(false) // Only operating area
photo.ShowContent(true)                // Restore meta information
```

`MediaSource(source)` loads previews in the background from data URL, local path or HTTP(S) (requires introduction of `ui/netimage`), supports PNG, JPEG, WebP, GIF (shows the first frame) and SVG, the size limit is the same as core.DecodeImage. Repeatedly setting the same source will not request again; `RetryMedia()` is explicitly overloaded, and `MediaSource("")` restores the default icon. `Media(view)` cancels source loading and uses the given View. Switching sources cancels old requests, and version verification prevents late results from overwriting new previews; a single request has a 15-second deadline and is not cached across instances. When removing a card, the app can call MediaSource("") to cancel the pending request.

`MediaLoading()`, `MediaError()` query the loading results. Preview loading does not change the attachment upload status, nor does it call OnRetry; when loading fails, the retry button in the media only reloads the image. When the attachment itself is being uploaded, processed, or failed, the life cycle mask and upload operation will be displayed first; the application can still explicitly call RetryMedia to reload the image. Disabling a card or ancestor disables the image retry button.

The URL preview uses a fixed thumbnail corresponding to the size file in landscape orientation (or natural size mode), and the vertical format is filled according to MediaAspectRatio and cropped in the center; loading, failure, and success do not change the preview size. Loading reports image/loading, success reports image/loaded, and failed media groups report image-error. Decoding and actual transmission are in the background, and status updates are returned to the UI via core.Update; applications should also use core.Update when calling these configuration methods from the background.

```go
photo.MediaSource("https://example.com/photo.png")
if err := photo.MediaError(); err != nil { /* 显示错误详情 */ }
photo.RetryMedia()
```

`TitleShimmer(ShimmerStyle)` independently configures the sweep period, width, reverse and single playback of the default title, and can share the same configuration value with ShimmerText.Style. Pass ShimmerStyle{} to restore the default, and use the default value for illegal periods/widths. The title animation is restarted when the configuration changes. Setting the same configuration repeatedly in each frame will not restart; static text is displayed when the title is hidden, the upload is completed, or the animation is reduced. PartStatus(Title) continues to determine the valid status of the title, and custom Content combines ShimmerText by itself.

```go
a.TitleShimmer(kit.ShimmerStyle{Duration: 3*time.Second, Spread: .45, Reverse: true})
```

`RemoveOnHover(on)` controls the display of the built-in remove button: it is enabled by default on the desktop, displayed when the mouse enters an attachment or the keyboard focus enters it; it is hidden after leaving and losing focus. Hide only changes the drawing transparency, retains the layout, semantics, and tab stops, and does not affect cancellation, retry, and customization operations. Android/iOS always displays by default. Touch screen web pages or mixed input applications can explicitly call `RemoveOnHover(false)` to always display. `ShowActions(false)` still removes the layout and interaction of the entire action area. The built-in remove button uses the Surface background, a thin border, and a rounded outline, centered in the upper right corner of the card, leaving half the button's overhang space on the top and right sides; this space is still retained when hidden transparently. The button diameter varies with the size range to 20/22/24/28dp. The card width still refers to the surface width. The overall footprint is increased by an additional half button width. Narrow windows give priority to reducing the surface. Cancel/retry and custom operations continue to be located in Actions, PartStyle(Actions) no longer affects the removal of the badge; ShowActions(false) hides the action area and the badge at the same time. PartStyle(Root) still configures the card surface, Hidden will be hidden together with the corner marker, and surface color matching/rounding will not change the appearance of the corner marker itself.
