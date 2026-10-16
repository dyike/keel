//go:build darwin && !ios

// SPDX-License-Identifier: Unlicense OR MIT
package text

import (
	"runtime"
	"sync"
	"weak"
)

type ctFontSourceKey struct {
	path   string
	data   uintptr
	length int
}
type ctFontSource struct {
	array uintptr
	// The native CFData owns a copy. Keep the immutable Go source alive as
	// well so its address cannot be reused while this weak cache entry lives.
	data []byte
}

var ctFontSources struct {
	sync.Mutex
	entries map[ctFontSourceKey]weak.Pointer[ctFontSource]
}

// The caller holds a native autorelease pool on its locked OS thread.
func sharedCTFontSource(key ctFontSourceKey, data []byte) *ctFontSource {
	ctFontSources.Lock()
	defer ctFontSources.Unlock()
	if p, ok := ctFontSources.entries[key]; ok {
		if source := p.Value(); source != nil {
			return source
		}
	}
	var array uintptr
	if key.path != "" {
		b := []byte(key.path)
		u := nativeInit.url(0, &b[0], len(b), false)
		if u != 0 {
			array = nativeInit.descriptorsURL(u)
			ct.release(u)
		}
	} else if ct.descriptorsFromData != nil && len(data) > 0 {
		d := ct.dataCreate(0, &data[0], len(data))
		if d != 0 {
			array = ct.descriptorsFromData(d)
			ct.release(d)
		}
	}
	if array == 0 {
		return nil
	}
	source := &ctFontSource{array: array, data: data}
	runtime.AddCleanup(source, func(array uintptr) { ct.release(array) }, array)
	if ctFontSources.entries == nil {
		ctFontSources.entries = make(map[ctFontSourceKey]weak.Pointer[ctFontSource])
	}
	if len(ctFontSources.entries) >= 512 {
		for key, p := range ctFontSources.entries {
			if p.Value() == nil {
				delete(ctFontSources.entries, key)
			}
		}
	}
	if len(ctFontSources.entries) < 512 {
		ctFontSources.entries[key] = weak.Make(source)
	}
	return source
}
