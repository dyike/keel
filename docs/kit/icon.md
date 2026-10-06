# Icon

English | [简体中文](icon.zh-CN.md)

`kit.Icon(kit.IconInfo).Size(24).Color(theme.Info)` Create vector icons. Render reads theme.Text each time Color is omitted. Size accepts float32 dp. The zero value `IconNone` does not draw anything and does not take up space. If the optional icon field is not set, there will be no icon. Built-in icons from Material Design: Check, Done, Close, Plus, Minus, Search, Copy, ChevronLeft, ChevronRight, ChevronDown, Info, Warning, Error, User, Inbox, Star, StarOutline, Calendar, Clock, Settings, Bell, Lock, Folder, File, Archive, Receipt, Home, Trash, Edit. Choose icons based on their meaning and don’t borrow icons with similar shapes (for example, use Info to represent settings). The parsing results will be cached, and calling `Icon` for each frame will not decode repeatedly.

The icon respects the parent size constraints and provides no semantics or keyboard operation itself; the control that contains it provides the name. Run `go run ./examples/components -section icon -theme dark` to check the dark theme and omit theme to check for light colors.

`Rotate(degrees)` Rotates clockwise around the center of the layout, negative numbers are counterclockwise, and the angle is normalized to 360°; non-finite values are ignored. Rotation does not change the layout occupancy. When the drawing exceeds the original frame, it will be cropped by the ancestors. Supports built-in Icon and VectorIcon; `Rotate(0)` restored. Size also ignores non-finite values. Added 0/45/90/180/270° examples to the component library; automatic testing verifies double-rate rotation, centering, and positioning. See below for SVG path/byte entry.

## SVG files and bytes

`SVGIcon(data)` and `SVGIconFile(path)` return `(*IconView, error)`, receiving up to 1MiB of SVG. File reading and parsing are executed synchronously, and the return value should be created and retained during initialization to avoid repeated parsing for each frame. The file entry only reads local files and does not listen for changes or download URLs.

Supports shapes, paths, groups, transforms, gradients, etc. [oksvg supported SVG subset](https://github.com/srwiley/oksvg). Use strict error mode; return an error if parsing fails, there is no valid viewBox/size, or if an unsupported element is identified. It is not a browser SVG engine and does not support complete web capabilities such as scripts, animations, fonts, filters, etc. It is suitable for the icon resources that come with the application.

By default, Color (theme.Text if not set) is used to color according to the alpha outline of the graphic, consistent with the built-in icon. `OriginalColors(true)` retains the source color, at which point currentColor resolves to black. The icon maintains the aspect ratio and is centered, reusing Size/Rotate. Rasterize by physical pixels, each instance only caches the image of the closest size/color; the longest side of the raster is limited to 2048 pixels, larger display sizes will enlarge the cached image.

Automated tests cover local files, input caps, invalid dimensions, and unsupported elements, as well as double-ratio non-zero viewBoxes, gradients, group transforms, rotations, primary/theme tints, and transparency pixels. Cache images are converted to the color space used by Gio to avoid darkening of translucent colors; longest side limit and theme/size switching are also verified. The native window is not visually accepted.
