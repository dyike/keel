# NavStack

English | [简体中文](nav_stack.zh-CN.md)

A full-height navigation stack for list/detail flows. Only the current page is rendered; earlier entries retain their controls and scroll position until removed. Place it in `el.Root` or a container with a bounded height.

```go
nav := kit.NavStack()
nav.Page(kit.NavPage{ID: "home", Title: "Home", Create: func() el.View {
    return kit.Button("Details", func() { nav.Push("detail") })
}})
nav.Page(kit.NavPage{ID: "detail", Title: "Details", Create: func() el.View {
    return kit.Input("Notes")
}})
// window.Open(window.Options{Content: el.Root(nav)})
```

## API and state

- `Page(NavPage)` registers an ID, title and optional `Create func() el.View`. The first enabled page is the initial root. Empty/duplicate IDs are ignored. A disabled page cannot be pushed or restored.
- `Push(id) bool` and `Pop() bool` are navigation commands. Successful commands notify `OnChange(func([]string))`; rejected commands return false. Pop cannot remove the root. The header's previous-page button invokes Pop.
- `Value() []string` returns a copied root-to-current route path. The callback also receives a copy.
- `SetValue([]string)` restores a path silently and retains the existing common prefix. It rejects the entire path if any ID is unknown/disabled. An empty path clears it. Repeated IDs are allowed as separate entries.
- `SetDisabled(bool)` disables navigation and descendant interaction. Silent restoration remains possible while disabled.
- `NavPage.Fill` lets the page own scrolling. Otherwise the stack adds a padded scroll container. `Create` runs once on first display of each entry. Pop releases that entry; a later Push creates new state. Applications retaining external references still own those references.

A stack is not an unlimited cache. Pages below the current one consume memory; remove unused entries with Pop or SetValue. Repeated calls to Render do not create new pages or animate inactive pages. Create may return nil for an empty destination.

## Interaction and semantics

The title names a `group`, with the current page's semantics beneath it. Hidden pages do not appear in automation or receive input. The localized previous-page button has at least a 44 dp target with default theme metrics; Tab reaches it and Enter/Space activates it. Parent disabling propagates through the retained page root.

This component does not intercept Escape, browser history or OS back events. That leaves modal dismissal and application routing under the host's control; the host may call Pop or restore a path through SetValue. The host also supplies safe-area insets. Modal presentations and swipe-to-go-back are not implemented by this initial component.

## Validation

`ui/kit/nav_stack_test.go` covers input, keyboard activation, retained scroll, common-prefix restoration, copied paths, invalid/disabled navigation and ancestor disabling. `ui/window/kit_nav_stack_test.go` checks real semantic snapshots and interaction at 320/600 dp. The `nav_stack` gallery section demonstrates three levels and retained draft input.
