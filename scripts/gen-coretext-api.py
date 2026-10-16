#!/usr/bin/env python3
"""Generate Keel's CoreText ABI calls from public macOS SDK declarations.

The table records ABI facts from CoreFoundation/CoreText/CoreGraphics headers.
Integer and CFRange calls use SyscallN; floating-point ABI calls use purego's
signature-aware trampoline. This generator needs only Python and gofmt.
"""
from pathlib import Path
import subprocess

# Go name | framework | SDK symbol | argument names/types | result
FUNCTIONS = '''
release|cf|CFRelease|obj:uintptr|
retain|cf|CFRetain|obj:uintptr|uintptr
hash|cf|CFHash|obj:uintptr|uint
equal|cf|CFEqual|a:uintptr b:uintptr|bool
stringWithCharacters|cf|CFStringCreateWithCharacters|allocator:uintptr chars:*uint16 count:int|uintptr
stringWithBytes|cf|CFStringCreateWithBytes|allocator:uintptr bytes:*byte count:int encoding:uint32 external:bool|uintptr
stringGetLength|cf|CFStringGetLength|value:uintptr|int
stringGetCharacters|cf|CFStringGetCharacters|value:uintptr span:cfRange out:*uint16|
dictionaryCreate|cf|CFDictionaryCreate|allocator:uintptr keys:*uintptr values:*uintptr count:int keyCB:uintptr valueCB:uintptr|uintptr
dictionaryGetValue|cf|CFDictionaryGetValue|dictionary:uintptr key:uintptr|uintptr
arrayGetCount|cf|CFArrayGetCount|array:uintptr|int
arrayGetValueAtIndex|cf|CFArrayGetValueAtIndex|array:uintptr index:int|uintptr
setCreate|cf|CFSetCreate|allocator:uintptr values:*uintptr count:int callbacks:uintptr|uintptr
arrayCreate|cf|CFArrayCreate|allocator:uintptr values:*uintptr count:int callbacks:uintptr|uintptr
numberCreate|cf|CFNumberCreate|allocator:uintptr kind:int value:unsafe.Pointer|uintptr
attributedString|cf|CFAttributedStringCreate|allocator:uintptr text:uintptr attributes:uintptr|uintptr
dataCreate|cf|CFDataCreate|allocator:uintptr bytes:*byte count:int|uintptr
fontWithDescriptor|text|CTFontCreateWithFontDescriptor|descriptor:uintptr size:float64 matrix:uintptr|uintptr
uiFontForLanguage|text|CTFontCreateUIFontForLanguage|kind:uint32 size:float64 language:uintptr|uintptr
fontWithTraits|text|CTFontCreateCopyWithSymbolicTraits|font:uintptr size:float64 matrix:uintptr value:uint32 mask:uint32|uintptr
fontWithAttrs|text|CTFontCreateCopyWithAttributes|font:uintptr size:float64 matrix:uintptr descriptor:uintptr|uintptr
descriptorWithAttrs|text|CTFontDescriptorCreateWithAttributes|attributes:uintptr|uintptr
descriptorMatching|text|CTFontDescriptorCreateMatchingFontDescriptor|descriptor:uintptr mandatory:uintptr|uintptr
descriptorAttribute|text|CTFontDescriptorCopyAttribute|descriptor:uintptr key:uintptr|uintptr
descriptorsFromData|text|CTFontManagerCreateFontDescriptorsFromData|data:uintptr|uintptr
fontGetAscent|text|CTFontGetAscent|font:uintptr|float64
fontGetDescent|text|CTFontGetDescent|font:uintptr|float64
fontGetLeading|text|CTFontGetLeading|font:uintptr|float64
fontGetSymbolicTraits|text|CTFontGetSymbolicTraits|font:uintptr|uint32
fontCreatePath|text|CTFontCreatePathForGlyph|font:uintptr glyph:uint16 matrix:uintptr|uintptr
fontGetBoundingRects|text|CTFontGetBoundingRectsForGlyphs|font:uintptr orientation:uint32 glyphs:*uint16 rects:*cgRect count:int|cgRect
fontDrawGlyphs|text|CTFontDrawGlyphs|font:uintptr glyphs:*uint16 positions:*cgPoint count:int context:uintptr|
paragraphStyleCreate|text|CTParagraphStyleCreate|settings:*ctParagraphStyleSetting count:int|uintptr
typesetterCreate|text|CTTypesetterCreateWithAttributedString|text:uintptr|uintptr
typesetterCreateLine|text|CTTypesetterCreateLine|typesetter:uintptr span:cfRange|uintptr
lineGetGlyphRuns|text|CTLineGetGlyphRuns|line:uintptr|uintptr
runGetGlyphCount|text|CTRunGetGlyphCount|run:uintptr|int
runGetStringRange|text|CTRunGetStringRange|run:uintptr|cfRange
runGetStatus|text|CTRunGetStatus|run:uintptr|uint32
runGetAttributes|text|CTRunGetAttributes|run:uintptr|uintptr
runGetGlyphs|text|CTRunGetGlyphs|run:uintptr span:cfRange out:*uint16|
runGetPositions|text|CTRunGetPositions|run:uintptr span:cfRange out:*cgPoint|
runGetAdvances|text|CTRunGetAdvances|run:uintptr span:cfRange out:*cgSize|
runGetStringIndices|text|CTRunGetStringIndices|run:uintptr span:cfRange out:*int|
colorSpaceDeviceRGB|graphics|CGColorSpaceCreateDeviceRGB||uintptr
bitmapContextCreate|graphics|CGBitmapContextCreate|data:unsafe.Pointer width:int height:int bits:int stride:int space:uintptr flags:uint32|uintptr
contextRelease|graphics|CGContextRelease|context:uintptr|
pathApply|graphics|CGPathApply|path:uintptr info:uintptr callback:uintptr|
pathRelease|graphics|CGPathRelease|path:uintptr|
poolPush|objc|objc_autoreleasePoolPush||uintptr
poolPop|objc|objc_autoreleasePoolPop|pool:uintptr|
'''
# Go name | framework | exported symbol | indirect (object variable) or address
DATA = '''
keyCallbacks|cf|kCFTypeDictionaryKeyCallBacks|address
valueCallbacks|cf|kCFTypeDictionaryValueCallBacks|address
setCallbacks|cf|kCFTypeSetCallBacks|address
arrayCallbacks|cf|kCFTypeArrayCallBacks|address
fontAttributeName|text|kCTFontAttributeName|indirect
paragraphStyleName|text|kCTParagraphStyleAttributeName|indirect
familyNameAttribute|text|kCTFontFamilyNameAttribute|indirect
traitsAttribute|text|kCTFontTraitsAttribute|indirect
weightTrait|text|kCTFontWeightTrait|indirect
symbolicTrait|text|kCTFontSymbolicTrait|indirect
cascadeList|text|kCTFontCascadeListAttribute|indirect
'''

def records(table):
    return [line.split('|') for line in table.strip().splitlines()]

functions = []
for name, library, symbol, arguments, result in records(FUNCTIONS):
    args = [part.split(':') for part in arguments.split()]
    signature = '(' + ', '.join(n + ' ' + t for n, t in args) + ')'
    if result:
        signature += ' ' + result
    # Pointers to floating-point structures remain integer ABI arguments.
    typed = result in ('float64', 'cgRect') or any(t == 'float64' for _, t in args)
    functions.append((name, library, symbol, args, result, signature, typed))

lines = ['''// Code generated by scripts/gen-coretext-api.py; DO NOT EDIT.
// SPDX-License-Identifier: Unlicense OR MIT
//go:build darwin && !ios

package text

import (
    "unsafe"
    "github.com/ebitengine/purego"
)

// Public ABI layouts: CFRange, CGPoint, CGSize, CGRect and
// CTParagraphStyleSetting from the macOS SDK.
type cfRange struct { location, length int }
type cgPoint struct { x, y float64 }
type cgSize struct { w, h float64 }
type cgRect struct { x, y, w, h float64 }
type ctParagraphStyleSetting struct { spec uint32; size uintptr; value unsafe.Pointer }

const (
    cfStringEncodingUTF8 = 0x08000100
    cfNumberSInt32Type = 3
    cfNumberSInt64Type = 4
    cfNumberFloat64Type = 6
    ctFontUIFontSystem = 2
    ctFontItalicTrait = 1 << 0
    ctFontColorGlyphsTrait = 1 << 13
    ctFontOrientationDefault = 0
    ctRunStatusRightToLeft = 1 << 0
    ctBaseWritingDirectionSpec = 13
    cgImageAlphaOnly = 7
    cgImageAlphaPremultipliedLst = 1
    cgBitmapByteOrder32Big = 4 << 12
)

// The owning frameworks stay loaded for the process lifetime.
var ct coreTextAPI

type coreTextAPI struct {''']
for name, _, _, _, _, signature, typed in functions:
    lines.append(f'    {name} func{signature}' if typed else f'    {name}PC uintptr')
for name, *_ in records(DATA):
    lines.append(f'    {name} uintptr')
lines += ['    weights [9]float64', '}']
for name, _, _, args, result, signature, typed in functions:
    if typed:
        continue
    words = []
    for argument, kind in args:
        if kind == 'cfRange':
            words += [f'uintptr({argument}.location)', f'uintptr({argument}.length)']
        elif kind.startswith('*'):
            words.append(f'uintptr(unsafe.Pointer({argument}))')
        elif kind == 'bool':
            words.append(f'coreTextBoolWord({argument})')
        else:
            words.append(f'uintptr({argument})')
    call = f'purego.SyscallN(api.{name}PC' + ''.join(', ' + word for word in words) + ')'
    lines.append(f'func (api *coreTextAPI) {name}{signature} {{')
    if not result:
        lines.append('    ' + call)
    elif result == 'cfRange':
        lines += ['    location, length, _ := ' + call, '    return cfRange{int(location), int(length)}']
    else:
        lines.append('    nativeResult, _, _ := ' + call)
        lines.append('    return uint8(nativeResult) != 0' if result == 'bool' else f'    return {result}(nativeResult)')
    lines.append('}')
lines += ['''func coreTextBoolWord(value bool) uintptr {
    if value { return 1 }; return 0
}

//go:generate python3 ../../scripts/gen-coretext-api.py

func loadCoreText() error {
    loader := newCoreTextLoader()
    defer loader.closeOnError()
''']
for name, library, symbol, _, _, _, typed in functions:
    if typed:
        lines.append(f'    loader.function("{library}", "{symbol}", &ct.{name})')
    else:
        lines.append(f'    ct.{name}PC = loader.symbol("{library}", "{symbol}")')
for name, library, symbol, mode in records(DATA):
    lines.append(f'    ct.{name} = loader.{mode}("{library}", "{symbol}")')
for i, weight in enumerate(['UltraLight', 'Thin', 'Light', 'Regular', 'Medium', 'Semibold', 'Bold', 'Heavy', 'Black']):
    lines.append(f'    ct.weights[{i}] = loader.scalar("appkit", "NSFontWeight{weight}")')
lines += ['    return loader.err', '}']
root = Path(__file__).resolve().parents[1]
output = root / 'gio/text/coretext_api_darwin.go'
output.write_text('\n'.join(lines) + '\n')
subprocess.run(['gofmt', '-w', str(output)], check=True)
