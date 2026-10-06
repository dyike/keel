# Automation

English | [简体中文](automation.zh-CN.md)

`cmd/keel-mcp` is an MCP server. After being connected, the Agent can start the Keel application, read out what is in the window, click, enter, press keys, scroll, take screenshots, and walk through the application like a human.

Two modes:

- **Visible Mode**: The window is displayed on the screen normally. You can watch the Agent operate it, or you can do it yourself and then let the Agent do it.
- **Interfaceless mode**: The window does not appear on the screen and is only rendered in memory. Suitable for running tests in batches without disturbing you.

Both modes do not use a real mouse and keyboard, nor do they require system permissions such as accessibility functions. A complete "input → click → open new window → save → close" process takes about 1.5 seconds (excluding compilation).

## Access

```sh
go install github.com/dyike/keel/cmd/keel-mcp@latest
claude mcp add keel -- keel-mcp
```

When modifying keel-mcp itself, use `go install ./cmd/keel-mcp` to install the local version in the repository.

`claude mcp add` will write Claude Code's configuration. For other clients that support MCP, just configure `keel-mcp` as a stdio type server. `keel-mcp` does not require parameters; which application to test is specified in the `launch` tool.

## What does a test look like?

Agent calls `launch` with parameter `{"command": "go run ./examples/multiwindow"}` and gets:

```
Window w1 "Keel 主窗口" 640×420
e1 text "Keel · Gio" @24,24 582×34
e2 text "你的名字" @42,88 546×21
e3 textbox "你的名字" value="" @42,115 546×40
e4 button "打招呼" @42,167 74×38
e5 button "打开设置窗口" @124,167 116×38
e6 text "主窗口和设置窗口拥有各自的状态。" @42,217 546×24
e7 text "⌘ + , 打开设置。" @42,253 546×21
```

Each line is an element: reference (`e3`), role, name, status, position (`@x,y width×height`, unit dp). Then:

```
type       {"ref": "e3", "text": "小明"}
click      {"text": "打招呼"}          → e6 text "你好，小明！"
press_key  {"key": "mod+,"}           → Open windows: w1 "Keel 主窗口" (shown above); w2 "设置" (new)
click      {"window": "w2", "text": "保存"}
screenshot {"window": "w2"}           → PNG
```

Each operation returns the element list after the operation, and the Agent does not need to call `snapshot` separately to confirm the result. When an action opens a new window, `(new)` is marked at the end of the list.

## Operate applications you launch yourself

The application is displayed on the screen normally. You first manually operate it to a certain state, and then let the Agent continue to do it. You watch it operate:

```sh
KEEL_AUTOMATION=1 go run ./examples/multiwindow
```

After the application starts, print the connection address, such as `keel automation: listening on /var/folders/…/T/keel/multiwindow-4821.sock`. Then have the Agent call `attach` (without parameters):

- When only one application is running, it connects directly and returns its current window and elements;
- If there are multiple addresses, list all addresses and let the Agent select one and pass the `socket` parameter;
- When not found, confirm that the application was started with `KEEL_AUTOMATION=1` and did not exit.

Differences from `launch`:

| | `launch` | `attach` |
| --- | --- | --- |
| Who starts the app | `keel-mcp` | You |
| `stop` | End the application | Just disconnect, the application continues to run |
| Connect again | Restart, the status is cleared | `attach` again, the status is retained |
| `logs` | Can see the output | Can't, the output is in the terminal you started it from |

An application can only accept one connection at a time. When another Agent session is already connected, `attach` will report "Another client is connected" after 3 seconds and will not stay stuck.

Each operation of the Agent will be immediately reflected in the window on the screen: the entered text, the result of the click, and the newly opened window. Your operations in the window can also be seen when the Agent reads it next time.

If you do not want to display the window, add `KEEL_HEADLESS=1`.

`KEEL_AUTOMATION` can also directly write the socket path (`KEEL_AUTOMATION=/tmp/my.sock`). In this case, `attach` will pass the same `socket`.

## Tool

| Tools | Parameters | Function |
| --- | --- | --- |
| `launch` | `command`, `dir?`, `timeout_seconds?`, `visible?` | Start the application with a shell, wait for it to be ready, and return to the first window. The window is not displayed by default; `visible: true` is displayed on the screen. The previously launched application will be stopped first. Compilation time is included in the timeout, default is 120 seconds |
| `attach` | `socket?` | Connect an app you launched with `KEEL_AUTOMATION=1`, preserving its current state |
| `snapshot` | `window?` | Elements in the window |
| `click` | `ref` or `text` or `x`+`y` | Left-click the center of the element or specify the coordinates |
| `type` | `text`, `ref?`, `clear?` | Enter at the cursor. Give `ref` click to focus first; `clear` first select all and replace with new text |
| `press_key` | `key` | Press a key combination: `enter` `esc` `tab` `shift+tab` `space` `backspace` `up` `mod+a` `ctrl+shift+s`. Window shortcut keys will trigger and Tab will move focus |
| `scroll` | `dy`, `ref?` or `x`+`y` | Scroll wheel, positive number downward, unit dp. Defaults to window center |
| `wait_for` | `text`, `timeout_ms?` | Wait until an element's name or value contains `text`. Used for background goroutine update interface scenarios, default 5 seconds |
| `screenshot` | `window?` | Window screenshot, 1 pixel = 1 dp |
| `windows` | | List open windows |
| `close_window` | `window?` | Close the window like clicking the close button. Close the last one and the application exits |
| `logs` | `lines?` | Applies the latest standard output and error output: panic, log, compile error. Only valid for applications launched by `launch` |
| `stop` | | Application of `launch`: End it; Application of `attach`: Disconnect only |

Except for `launch`, `attach`, `logs` and `stop`, `window` (such as `w2`) can be used. When omitted, it will act on the current window: the window that was last operated; if it has not been operated yet, it is the most recently opened window.

### Element

| role | from | additional information |
| --- | --- | --- |
| `text` | `el.Text`、`kit.Kbd` | |
| `button` | `kit.Button`, `el.Div` with `OnClick` | `disabled`; `value: loading` when the kit is loaded; the selected button (`Selected(true)`) has `selected: true`, and is not listed when it is not selected |
| `link` | `kit.Link` | |
| `textbox` | `kit.Input`, `kit.TextArea`, `el.Input` | `value`; the values of the password boxes are equal length `•` |
| `checkbox` | `kit.Checkbox` | `checked` / `unchecked`; when half selected, `value` is mixed |
| `radio` | Every option for `kit.RadioGroup` | `checked` / `unchecked` |
| `switch` | `kit.Switch` | `checked` / `unchecked` |
| `select` | `kit.Select` | `value` is the currently selected item; `option` will appear after clicking it |
| `option` | Expanded drop-down options | `selected` |
| `tab` | `kit.Tabs`'s tags | `selected` |
| `table` | `kit.Table` | `value` is the total number of rows, such as `36 行` |
| `columnheader` | Header, click to sort | |
| `row` | The rows visible in the table are named with vertical lines connecting the columns | `selected` |
| `slider` | `kit.Slider` | `value` is the current value; click the track or focus and press the direction keys to adjust |
| `accordion` | `kit.Accordion` | Title and expanded content listed separately |
| `tag` | kit tag container | value is neutral/info/success/warning/danger, selected is the selected state; the select and remove buttons are listed separately |
| `group` | kit component grouping | The name is the group title, retaining the semantics of sub-components |
| `status` | kit status bar | The left and right content are listed separately as sub-elements |
| `alert` | kit inline prompt container | The name is the title, the value is neutral/info/success/warning/danger; the text and close button are listed separately |
| `badge` | Numbers, dots, icons (not clickable) | The name is the original count, `value` is the displayed value, `dot` or `icon` |
| `toggle` | Status button | `selected` means selected, supports `disabled` |
| `disclosure` | Collapse panel title | `value` is expanded / collapsed, supports `disabled` |
| `avatar` | kit avatar | The name is the person's name, the value is online/busy/offline, and the status is empty |
| `image` | `kit.Image`, Markdown image | `value` is loading / loaded / error, the name is the alternative text |
| `footnotes` | Markdown footnotes | Quotes and backlinks listed separately |
| `progressbar` | `kit.Progress` | `value` is a percentage or indeterminate |
| `dialog` | The open panel of `kit.Dialog`, `kit.Sheet`, `kit.Popover`, `kit.HoverCard` | The text and buttons in it are listed separately |
| `alertdialog` | `ConfirmDanger` and `Persistent()` of `kit.Dialog` | The point mask is not closed, Esc equals to cancel; the elements inside are listed separately |
| `radiogroup` | `kit.RadioGroup` | Each option is `radio`, listed separately |
| `listbox` | Open list of `kit.Select` / `kit.Combobox` | Option is `option`, listed separately |
| `combobox` | `kit.Combobox` | `value` is the current selection; the text box and expand button inside are listed separately |
| `grid` / `gridcell` | `kit.Calendar` | Every day is `gridcell`, the name is the date, `selected` means selected or within the range |
| `list` / `step` | `kit.Stepper` | Each step of `value` is done / current / upcoming |
| `form` | `kit.Form` | Row labels and controls are listed separately, and controls are named after the row label |
| `tree` / `treeitem` | `kit.Tree` | The node `value` is expanded / collapsed, `selected` means selected |
| `tablist` / `tabpanel` | `kit.Tabs` | The label is `tab`; the panel is named after the label title |
| `navigation` | `kit.Pagination` | `value` is "current page/total number of pages" |
| `log` / `article` | `kit.MessageScroller` / `kit.Message` | The message is named after the author and the content is listed separately |
| `attachment` | `kit.Attachment` | The name is the file name, `value` is the upload progress or error |
| `separator` | Separator bars of `kit.Resizable`, `kit.Dock` | `value` is the size of the first panel (Resizable) |
| Dock area for `region` | `kit.Dock` | Name is the current panel title; labels, menu buttons, and content are listed separately |
| `toolbar` | `kit.Toolbar` | Buttons listed separately |
| `banner` | `kit.TitleBar` | The name is the title; window buttons and application content are listed separately |
| `figure` | `kit.LineChart`, `kit.BarChart`, `kit.Plot` | The name is the title, and the `value` of the chart is "number of items x number of series"; after switching to the data table, each row is listed with `row` |
| `tooltip` | Tip for `kit.WithTooltip` | The name is the tip text, the rich content and the action keys are listed as child elements |
| `menu` | Open `kit.Menu` | Menu items listed separately |
| `menuitem` | Menu item | When there is a submenu, `value` is submenu; supported `disabled` |
| `link` | Link in Markdown paragraph | `value` is the URL |
| `search` | Find panel for `kit.CodeEditor` | Input boxes and buttons listed separately |
| `code` | Markdown code block | The name is the language; the code text and "Copy" button inside are listed separately |

Other roles declared by the component through `core.Role` or el's `Role` will be listed as they are, and do not need to be registered in the automation code. You only need to supplement the above table. By default, an element will absorb the text inside it; if the child elements need to be listed separately (containers such as dialog boxes and menus), add the role of `ui/window/automation.go` to `containerRoles`.

In automation mode (`KEEL_AUTOMATION` is set), "reduced motion" is turned on by default when the application starts: overlays such as Sheet that slide in directly appear at the final position, and the coordinates read by the Agent are the clicked coordinates. The application can change it back by itself using `theme.SetReducedMotion(false)`.

The element list contains only the visible part: table rows and page content that scroll out of view are not listed, and partially visible elements report their positions as visible parts. To see more lines, first `scroll`.

`ref` is only valid until the next operation because it is renumbered for each operation. Positioning by text (`text`) is more stable: complete matching takes precedence over partial matching, and controls take precedence over ordinary text; at the same time, the one drawn later is taken first, so buttons in dialog boxes and drop-down boxes take precedence over elements of the same name covered by them.

Coordinates are only used when there is no better way. Screenshots and coordinates use the same set of units: one pixel on the screenshot is one dp.

## Principle

```
Agent ──MCP(stdio)──► keel-mcp ──JSON Line (unix socket)──► application process
                         │                                 │
               launch：启动应用并设置             ui/window 自动化模式：
               KEEL_AUTOMATION=<socket>          窗口不上屏，按请求渲染
               attach：连接你启动的应用           一个连接断开后等下一个
```

In the application process:

- `ui/window` reads the `KEEL_AUTOMATION` environment variable when starting, and assigns a **shadow window** to each window: the same set of components, the same size, but has its own Gio input route. Agent operations are sent to the shadow window.
- In visible mode, the real window is displayed as usual, and your mouse and keyboard follow the original path of the real window without being affected at all. Both sides share the same batch of component objects, so the Agent clicks the button through the shadow window, the callback is executed, the state changes, and the real window is immediately redrawn; the content you enter in the real window can also be seen the next time the shadow window is rendered. The size of the shadow window follows the real window. If you drag the window to change the size, the coordinates read by the Agent will also change.
- There is no real window in interfaceless mode (`KEEL_HEADLESS=1`). `window.Main` does not enter the system event loop and only processes requests on the socket.
- Components, callbacks, window shortcut keys, `core.Update`, and frame locks all use the same code as the real window.
- "What's in the page" comes from the semantic tree that Gio generates every frame: each component declares its role, name, and state, and the router calculates its absolute position in the window. The el element declares this information when drawing, and the Gio code written by yourself is declared with `core.Semantic`.
- Clicks, inputs, and scrolling are converted into Gio pointer and keyboard events, sent to the router of this window, and then rendered until the picture is stable: if the callback changes the state, you have to draw another frame to see it, and a maximum of 10 frames can be drawn.

The application side protocol is written in the file comments of `ui/window/automation_server.go`. `keel-mcp` does not reference any Keel package, only this protocol. If you want to write a test client in other languages, just send JSON according to the protocol.

## Limitations

- **The focus is on its own.** After the Agent clicks into an input box, the focus is in the shadow window. The input box on the screen does not display the cursor and blue border, but the entered text will be displayed. On the other hand, if you click on the input box on the screen and the Agent calls `type` without `ref`, it will report "no focus". You must bring `ref`.
- **When you and the Agent operate at the same time, they will be executed in order** without conflict. But when you hold down the mouse and drag, the Agent inserts and clicks, and the dragging state may be messed up.
- **Closing the window will wait until it is actually closed before returning.** In visible mode, `close_window` turns off the real window on the screen. It will be restored after the window is destroyed, so the window list returned is what it looks like on the screen.
- **Agent does not operate on the real window itself.** Issues at the system window level cannot be detected: window position and size, system menu, input method candidate box, and main thread-related deadlocks between real windows. The latter is covered with a real window test of `KEEL_DESKTOP=1`, see [Testing](testing.md#real-window-test).
- **`native/*` is not in range. ** Permissions, screenshots, and global shortcut keys adjust the system API, and the automation mode will not simulate them.
- **Time only advances when requested.** App only renders when requested. The background goroutine's `core.Update` takes effect on the next request; to wait for it, use `wait_for`.
- **Hover, drag, right-click, and double-click are not supported.** Existing components are not used and will be added to `ui/window/automation.go` if needed.
- **Does not report focus position.** `type` without `ref` is entered into the input box that currently has focus; an error will be reported if there is no focus.
- **Layouts written with `core.Func` are not visible to Agent** unless declared with `core.Semantic`. It is not necessary to write in el, see [Extension Guide](extending.md#add-new-components).

## Used in Go tests

`cmd/keel-mcp/main_test.go` Use the client of MCP official SDK to start `keel-mcp`, walk through the multiwindow example completely (`TestEndToEnd`), and test the connection, disconnection, reconnection and occupation prompts of `attach` (`TestAttach`), which is also a template for writing such tests. `ui/window/automation_test.go` Test automation modes (scrolling, tab focus, callback closing window) directly in the process. Both run with `go test ./...` without pop-up windows.

The floating content is listed after the main content in the order of declaration. When there is a modal overlay, the snapshot omits the main content and the overlay below it; it will be restored after closing. The bottom layer skips the entire subtree via the internal `el-inert` semantic tag, which is not output as a component role.
