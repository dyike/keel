//go:build darwin && !ios

/*
MIT License

Copyright (c) 2026 MyGo contributors

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

package text

import (
	"fmt"
	"strings"
	"unsafe"

	"github.com/ebitengine/purego"
)

// Core Foundation, Core Text and Core Graphics structs passed by value.
type (
	cfRange struct{ location, length int }
	cgPoint struct{ x, y float64 }
	cgSize  struct{ w, h float64 }
	cgRect  struct{ x, y, w, h float64 }

	ctParagraphStyleSetting struct {
		spec  uint32
		size  uintptr
		value unsafe.Pointer
	}
)

const (
	cfStringEncodingUTF8 = 0x08000100
	cfNumberSInt32Type   = 3
	cfNumberSInt64Type   = 4
	cfNumberFloat64Type  = 6

	ctFontUIFontSystem         = 2
	ctFontItalicTrait          = 1 << 0
	ctFontColorGlyphsTrait     = 1 << 13
	ctFontOrientationDefault   = 0
	ctRunStatusRightToLeft     = 1 << 0
	ctBaseWritingDirectionSpec = 13

	cgImageAlphaOnly             = 7
	cgImageAlphaPremultipliedLst = 1
	cgBitmapByteOrder32Big       = 4 << 12
)

var ct struct {
	// Core Foundation
	release               func(obj uintptr)
	retain                func(obj uintptr) uintptr
	hash                  func(obj uintptr) uint
	equal                 func(a, b uintptr) bool
	stringWithCharacters  func(alloc uintptr, chars *uint16, n int) uintptr
	stringWithBytes       func(alloc uintptr, bytes *byte, n int, encoding uint32, external bool) uintptr
	stringGetLength       func(s uintptr) int
	stringGetCharacters   func(s uintptr, r cfRange, out *uint16)
	dictionaryCreate      func(alloc uintptr, keys, values *uintptr, n int, keyCallbacks, valueCallbacks uintptr) uintptr
	dictionaryGetValue    func(d, key uintptr) uintptr
	arrayGetCount         func(a uintptr) int
	arrayGetValueAtIndex  func(a uintptr, i int) uintptr
	setCreate             func(alloc uintptr, values *uintptr, n int, callbacks uintptr) uintptr
	arrayCreate           func(alloc uintptr, values *uintptr, n int, callbacks uintptr) uintptr
	numberCreate          func(alloc uintptr, typ int, value unsafe.Pointer) uintptr
	attributedString      func(alloc, str, attrs uintptr) uintptr
	dataCreate            func(alloc uintptr, bytes *byte, n int) uintptr
	keyCallbacks          uintptr
	valueCallbacks        uintptr
	setCallbacks          uintptr
	arrayCallbacks        uintptr
	fontAttributeName     uintptr
	paragraphStyleName    uintptr
	familyNameAttribute   uintptr
	traitsAttribute       uintptr
	weightTrait           uintptr
	symbolicTrait         uintptr
	fontWithDescriptor    func(desc uintptr, size float64, matrix uintptr) uintptr
	uiFontForLanguage     func(typ uint32, size float64, language uintptr) uintptr
	fontWithTraits        func(font uintptr, size float64, matrix uintptr, value, mask uint32) uintptr
	fontWithAttrs         func(font uintptr, size float64, matrix, desc uintptr) uintptr
	descriptorWithAttrs   func(attrs uintptr) uintptr
	descriptorMatching    func(desc, mandatory uintptr) uintptr
	descriptorAttribute   func(desc, name uintptr) uintptr
	descriptorsFromData   func(data uintptr) uintptr
	fontGetAscent         func(font uintptr) float64
	fontGetDescent        func(font uintptr) float64
	fontGetLeading        func(font uintptr) float64
	fontGetSymbolicTraits func(font uintptr) uint32
	fontCreatePath        func(font uintptr, glyph uint16, matrix uintptr) uintptr
	fontGetBoundingRects  func(font uintptr, orientation uint32, glyphs *uint16, rects *cgRect, n int) cgRect
	fontDrawGlyphs        func(font uintptr, glyphs *uint16, positions *cgPoint, n int, context uintptr)
	paragraphStyleCreate  func(settings *ctParagraphStyleSetting, n int) uintptr
	typesetterCreate      func(str uintptr) uintptr
	typesetterCreateLine  func(typesetter uintptr, r cfRange) uintptr
	lineGetGlyphRuns      func(line uintptr) uintptr
	runGetGlyphCount      func(run uintptr) int
	runGetStringRange     func(run uintptr) cfRange
	runGetStatus          func(run uintptr) uint32
	runGetAttributes      func(run uintptr) uintptr
	runGetGlyphs          func(run uintptr, r cfRange, out *uint16)
	runGetPositions       func(run uintptr, r cfRange, out *cgPoint)
	runGetAdvances        func(run uintptr, r cfRange, out *cgSize)
	runGetStringIndices   func(run uintptr, r cfRange, out *int)

	// Font fallback families before the system cascade.
	cascadeList uintptr

	// Core Graphics
	colorSpaceDeviceRGB func() uintptr
	bitmapContextCreate func(data unsafe.Pointer, w, h, bitsPerComponent, bytesPerRow int, space uintptr, info uint32) uintptr
	contextRelease      func(ctx uintptr)
	pathApply           func(path, info, fn uintptr)
	pathRelease         func(path uintptr)

	// Objective-C
	poolPush func() uintptr
	poolPop  func(pool uintptr)
}

func loadCoreText() error {
	var missing []string
	lib := func(path string) uintptr {
		h, err := purego.Dlopen(path, purego.RTLD_NOW|purego.RTLD_GLOBAL)
		if err != nil {
			missing = append(missing, path)
		}
		return h
	}
	bind := func(lib uintptr, fn any, name string) {
		sym, err := purego.Dlsym(lib, name)
		if err != nil || sym == 0 {
			missing = append(missing, name)
			return
		}
		purego.RegisterFunc(fn, sym)
	}
	// Constants are pointers to CFStringRefs, or callback structs.
	addr := func(lib uintptr, name string) uintptr {
		sym, err := purego.Dlsym(lib, name)
		if err != nil || sym == 0 {
			missing = append(missing, name)
		}
		return sym
	}
	value := func(lib uintptr, name string) uintptr {
		p := addr(lib, name)
		if p == 0 {
			return 0
		}
		return **(**uintptr)(unsafe.Pointer(&p))
	}
	cf := lib("/System/Library/Frameworks/CoreFoundation.framework/CoreFoundation")
	text := lib("/System/Library/Frameworks/CoreText.framework/CoreText")
	cg := lib("/System/Library/Frameworks/CoreGraphics.framework/CoreGraphics")
	objcLib := lib("/usr/lib/libobjc.A.dylib")
	lib("/System/Library/Frameworks/AppKit.framework/AppKit")
	if len(missing) > 0 {
		return fmt.Errorf("cannot load %s", strings.Join(missing, ", "))
	}
	bindDirect(cf, text, cg, addr)
	bind(cf, &ct.stringWithBytes, "CFStringCreateWithBytes")
	bind(cf, &ct.stringGetLength, "CFStringGetLength")
	bind(cf, &ct.stringGetCharacters, "CFStringGetCharacters")
	bind(cf, &ct.setCreate, "CFSetCreate")
	bind(cf, &ct.arrayCreate, "CFArrayCreate")
	bind(cf, &ct.dataCreate, "CFDataCreate")
	ct.keyCallbacks = addr(cf, "kCFTypeDictionaryKeyCallBacks")
	ct.valueCallbacks = addr(cf, "kCFTypeDictionaryValueCallBacks")
	ct.setCallbacks = addr(cf, "kCFTypeSetCallBacks")
	ct.arrayCallbacks = addr(cf, "kCFTypeArrayCallBacks")
	ct.fontAttributeName = value(text, "kCTFontAttributeName")
	ct.paragraphStyleName = value(text, "kCTParagraphStyleAttributeName")
	ct.familyNameAttribute = value(text, "kCTFontFamilyNameAttribute")
	ct.traitsAttribute = value(text, "kCTFontTraitsAttribute")
	ct.weightTrait = value(text, "kCTFontWeightTrait")
	ct.symbolicTrait = value(text, "kCTFontSymbolicTrait")
	bind(text, &ct.fontWithDescriptor, "CTFontCreateWithFontDescriptor")
	bind(text, &ct.uiFontForLanguage, "CTFontCreateUIFontForLanguage")
	bind(text, &ct.fontWithAttrs, "CTFontCreateCopyWithAttributes")
	bind(text, &ct.fontWithTraits, "CTFontCreateCopyWithSymbolicTraits")
	bind(text, &ct.descriptorWithAttrs, "CTFontDescriptorCreateWithAttributes")
	bind(text, &ct.descriptorMatching, "CTFontDescriptorCreateMatchingFontDescriptor")
	bind(text, &ct.descriptorAttribute, "CTFontDescriptorCopyAttribute")
	bind(text, &ct.fontGetAscent, "CTFontGetAscent")
	bind(text, &ct.fontGetDescent, "CTFontGetDescent")
	bind(text, &ct.fontGetLeading, "CTFontGetLeading")
	bind(text, &ct.fontGetBoundingRects, "CTFontGetBoundingRectsForGlyphs")
	bind(text, &ct.fontDrawGlyphs, "CTFontDrawGlyphs")
	bind(text, &ct.paragraphStyleCreate, "CTParagraphStyleCreate")
	bind(cg, &ct.colorSpaceDeviceRGB, "CGColorSpaceCreateDeviceRGB")
	bind(cg, &ct.bitmapContextCreate, "CGBitmapContextCreate")
	bind(cg, &ct.contextRelease, "CGContextRelease")
	bind(objcLib, &ct.poolPush, "objc_autoreleasePoolPush")
	bind(objcLib, &ct.poolPop, "objc_autoreleasePoolPop")
	if len(missing) > 0 {
		return fmt.Errorf("missing %s", strings.Join(missing, ", "))
	}
	optional := func(lib uintptr, name string) uintptr {
		if p, err := purego.Dlsym(lib, name); err == nil && p != 0 {
			return **(**uintptr)(unsafe.Pointer(&p))
		}
		return 0
	}
	ct.cascadeList = optional(text, "kCTFontCascadeListAttribute")
	// Optional: adding fonts from memory (macOS 10.13).
	if sym, err := purego.Dlsym(text, "CTFontManagerCreateFontDescriptorsFromData"); err == nil && sym != 0 {
		purego.RegisterFunc(&ct.descriptorsFromData, sym)
	}
	return nil
}

// bindDirect binds the functions of Core Foundation and Core Text that
// each layout calls to call them through purego.SyscallN, which allocates
// once a call where RegisterFunc's reflection allocates three or more
// times. They take integers, pointers and CFRanges, which C passes in two
// integer registers and returns in the first two on arm64 and amd64, as
// SyscallN's r1 and r2.
func bindDirect(cf, text, cg uintptr, addr func(lib uintptr, name string) uintptr) {
	createPath := addr(text, "CTFontCreatePathForGlyph")
	ct.fontCreatePath = func(font uintptr, glyph uint16, matrix uintptr) uintptr {
		r, _, _ := purego.SyscallN(createPath, font, uintptr(glyph), matrix)
		return r
	}
	traits := addr(text, "CTFontGetSymbolicTraits")
	ct.fontGetSymbolicTraits = func(font uintptr) uint32 { r, _, _ := purego.SyscallN(traits, font); return uint32(r) }
	apply, releasePath := addr(cg, "CGPathApply"), addr(cg, "CGPathRelease")
	ct.pathApply = func(path, info, fn uintptr) { purego.SyscallN(apply, path, info, fn) }
	ct.pathRelease = func(path uintptr) { purego.SyscallN(releasePath, path) }

	release, retain := addr(cf, "CFRelease"), addr(cf, "CFRetain")
	ct.release = func(obj uintptr) { purego.SyscallN(release, obj) }
	ct.retain = func(obj uintptr) uintptr { r, _, _ := purego.SyscallN(retain, obj); return r }
	hash, equal := addr(cf, "CFHash"), addr(cf, "CFEqual")
	ct.hash = func(obj uintptr) uint { r, _, _ := purego.SyscallN(hash, obj); return uint(r) }
	ct.equal = func(a, b uintptr) bool { r, _, _ := purego.SyscallN(equal, a, b); return uint8(r) != 0 }
	stringWithCharacters := addr(cf, "CFStringCreateWithCharacters")
	ct.stringWithCharacters = func(alloc uintptr, chars *uint16, n int) uintptr {
		r, _, _ := purego.SyscallN(stringWithCharacters, alloc, uintptr(unsafe.Pointer(chars)), uintptr(n))
		return r
	}
	dictionaryCreate, dictionaryGetValue := addr(cf, "CFDictionaryCreate"), addr(cf, "CFDictionaryGetValue")
	ct.dictionaryCreate = func(alloc uintptr, keys, values *uintptr, n int, keyCallbacks, valueCallbacks uintptr) uintptr {
		r, _, _ := purego.SyscallN(dictionaryCreate, alloc, uintptr(unsafe.Pointer(keys)), uintptr(unsafe.Pointer(values)), uintptr(n), keyCallbacks, valueCallbacks)
		return r
	}
	ct.dictionaryGetValue = func(d, key uintptr) uintptr { r, _, _ := purego.SyscallN(dictionaryGetValue, d, key); return r }
	arrayGetCount, arrayGetValueAtIndex := addr(cf, "CFArrayGetCount"), addr(cf, "CFArrayGetValueAtIndex")
	ct.arrayGetCount = func(a uintptr) int { r, _, _ := purego.SyscallN(arrayGetCount, a); return int(r) }
	ct.arrayGetValueAtIndex = func(a uintptr, i int) uintptr {
		r, _, _ := purego.SyscallN(arrayGetValueAtIndex, a, uintptr(i))
		return r
	}
	numberCreate := addr(cf, "CFNumberCreate")
	ct.numberCreate = func(alloc uintptr, typ int, value unsafe.Pointer) uintptr {
		r, _, _ := purego.SyscallN(numberCreate, alloc, uintptr(typ), uintptr(value))
		return r
	}
	attributedString := addr(cf, "CFAttributedStringCreate")
	ct.attributedString = func(alloc, str, attrs uintptr) uintptr {
		r, _, _ := purego.SyscallN(attributedString, alloc, str, attrs)
		return r
	}

	typesetterCreate, typesetterCreateLine := addr(text, "CTTypesetterCreateWithAttributedString"), addr(text, "CTTypesetterCreateLine")
	ct.typesetterCreate = func(str uintptr) uintptr { r, _, _ := purego.SyscallN(typesetterCreate, str); return r }
	ct.typesetterCreateLine = func(typesetter uintptr, r cfRange) uintptr {
		line, _, _ := purego.SyscallN(typesetterCreateLine, typesetter, uintptr(r.location), uintptr(r.length))
		return line
	}
	lineGetGlyphRuns := addr(text, "CTLineGetGlyphRuns")
	ct.lineGetGlyphRuns = func(line uintptr) uintptr { r, _, _ := purego.SyscallN(lineGetGlyphRuns, line); return r }

	runGetGlyphCount, runGetStringRange := addr(text, "CTRunGetGlyphCount"), addr(text, "CTRunGetStringRange")
	runGetStatus, runGetAttributes := addr(text, "CTRunGetStatus"), addr(text, "CTRunGetAttributes")
	ct.runGetGlyphCount = func(run uintptr) int { r, _, _ := purego.SyscallN(runGetGlyphCount, run); return int(r) }
	ct.runGetStringRange = func(run uintptr) cfRange {
		loc, n, _ := purego.SyscallN(runGetStringRange, run)
		return cfRange{int(loc), int(n)}
	}
	ct.runGetStatus = func(run uintptr) uint32 { r, _, _ := purego.SyscallN(runGetStatus, run); return uint32(r) }
	ct.runGetAttributes = func(run uintptr) uintptr { r, _, _ := purego.SyscallN(runGetAttributes, run); return r }
	runGetGlyphs, runGetPositions := addr(text, "CTRunGetGlyphs"), addr(text, "CTRunGetPositions")
	runGetAdvances, runGetStringIndices := addr(text, "CTRunGetAdvances"), addr(text, "CTRunGetStringIndices")
	ct.runGetGlyphs = func(run uintptr, r cfRange, out *uint16) {
		purego.SyscallN(runGetGlyphs, run, uintptr(r.location), uintptr(r.length), uintptr(unsafe.Pointer(out)))
	}
	ct.runGetPositions = func(run uintptr, r cfRange, out *cgPoint) {
		purego.SyscallN(runGetPositions, run, uintptr(r.location), uintptr(r.length), uintptr(unsafe.Pointer(out)))
	}
	ct.runGetAdvances = func(run uintptr, r cfRange, out *cgSize) {
		purego.SyscallN(runGetAdvances, run, uintptr(r.location), uintptr(r.length), uintptr(unsafe.Pointer(out)))
	}
	ct.runGetStringIndices = func(run uintptr, r cfRange, out *int) {
		purego.SyscallN(runGetStringIndices, run, uintptr(r.location), uintptr(r.length), uintptr(unsafe.Pointer(out)))
	}
}
