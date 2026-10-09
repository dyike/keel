// Package appkit calls AppKit through the Objective-C runtime with purego,
// without cgo. It is ui/window's only way to reach native macOS code.
//
// Rules: create Go callbacks (classes, blocks of one signature) once, since
// purego never frees them; route by object pointer, never by Go pointer.
// Objects from alloc/init/new/copy are owned and must be released. Work that
// creates temporary objects runs in Pool, MainAsync or MainSync, which wrap
// it in an autorelease pool.
package appkit
