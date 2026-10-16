// Copyright (c) 2026 MyGo contributors
// SPDX-License-Identifier: MIT
// Font matching follows MyGo (bad73d119dc8); Gio adapter by Keel contributors.
// See third_party/mygo.LICENSE.
//go:build darwin && !ios

package text

import (
	"image"
	"math"
	"os"
	"runtime"
	"slices"
	"strings"
	"sync"
	"unicode/utf16"
	"unsafe"

	"github.com/dyike/keel/gio/f32"
	giofont "github.com/dyike/keel/gio/font"
	gioot "github.com/dyike/keel/gio/font/opentype"
	"github.com/dyike/keel/gio/io/system"
	"github.com/dyike/keel/gio/op"
	"github.com/dyike/keel/gio/op/clip"
	"github.com/dyike/keel/gio/op/paint"
	"github.com/dyike/keel/typesetting/font"
	ot "github.com/dyike/keel/typesetting/font/opentype"
	"github.com/dyike/keel/typesetting/font/opentype/tables"
	"github.com/dyike/keel/typesetting/shaping"
	"github.com/ebitengine/purego"
	"github.com/ebitengine/purego/objc"
	"golang.org/x/image/math/fixed"
)

var nativeInit struct {
	sync.Once
	err             error
	url             func(uintptr, *byte, int, bool) uintptr
	descriptorsURL  func(uintptr) uintptr
	fontDescriptor  func(uintptr) uintptr
	fontName        func(uintptr) uintptr
	urlPath         func(uintptr, int) uintptr
	nameKey, urlKey uintptr
}

func initNativeFonts() {
	nativeInit.err = loadCoreText()
	if nativeInit.err != nil {
		return
	}
	cf, _ := purego.Dlopen("/System/Library/Frameworks/CoreFoundation.framework/CoreFoundation", purego.RTLD_NOW|purego.RTLD_LOCAL)
	ctlib, _ := purego.Dlopen("/System/Library/Frameworks/CoreText.framework/CoreText", purego.RTLD_NOW|purego.RTLD_LOCAL)
	purego.RegisterLibFunc(&nativeInit.url, cf, "CFURLCreateFromFileSystemRepresentation")
	purego.RegisterLibFunc(&nativeInit.urlPath, cf, "CFURLCopyFileSystemPath")
	purego.RegisterLibFunc(&nativeInit.descriptorsURL, ctlib, "CTFontManagerCreateFontDescriptorsFromURL")
	purego.RegisterLibFunc(&nativeInit.fontDescriptor, ctlib, "CTFontCopyFontDescriptor")
	purego.RegisterLibFunc(&nativeInit.fontName, ctlib, "CTFontCopyPostScriptName")
	for name, dst := range map[string]*uintptr{"kCTFontNameAttribute": &nativeInit.nameKey, "kCTFontURLAttribute": &nativeInit.urlKey} {
		p, err := purego.Dlsym(ctlib, name)
		if err != nil {
			nativeInit.err = err
			return
		}
		*dst = **(**uintptr)(unsafe.Pointer(&p))
	}
}

// NativeFontData explicitly supplies an original TTC face for native shaping.
// This bypasses Face(): custom variation wrappers must retain the Go path
// unless they deliberately supply a corresponding native source identity.
// Theme's private prepared face implements it without parsing Go tables.
type nativeFontData interface{ NativeFontData() ([]byte, int) }
type rawFontData interface{ FontData() ([]byte, int) }

type ctRegistered struct {
	desc uintptr
	meta giofont.Font
}
type ctStyle struct {
	families string
	size     fixed.Int26_6
	weight   giofont.Weight
	style    giofont.Style
}
type ctNativeFace struct {
	font            uintptr
	face            *font.Face // identity only; never parsed or passed to HarfBuzz
	location        string
	index, instance int
	located         bool
}

// Cleanup owns only native handles, not the shaper. It therefore cannot keep
// its cleanup target alive. Font records keep exact sized/variable identities.
type ctResources struct{ owned map[uintptr]int }

func (r *ctResources) own(p uintptr) uintptr {
	if p != 0 {
		r.owned[p]++
	}
	return p
}
func (r *ctResources) drop(p uintptr) {
	if p != 0 {
		r.owned[p]--
		if r.owned[p] == 0 {
			delete(r.owned, p)
		}
		ct.release(p)
	}
}
func (r *ctResources) release() {
	for p, n := range r.owned {
		for range n {
			ct.release(p)
		}
	}
	clear(r.owned)
}

type ctShaper struct {
	resources          *ctResources
	sourceOwners       []*ctFontSource
	registered         map[string][]ctRegistered
	primary            map[ctStyle]uintptr
	faces              []*ctNativeFace
	byHash             map[uint][]int
	query              giofont.Font
	families, defaults []string
	styles             [2]uintptr
	u16                []uint16
	at                 []int
	active             uintptr
	direction          system.TextDirection
	indices            []int
	ids                []uint16
	points             []cgPoint
	advances           []cgSize
	clusters           []int
	rects              []cgRect
	bitmapsCache       bitmapCache
}

func newNativeFontShaper(collection []FontFace) nativeFontShaper {
	nativeInit.Do(initNativeFonts)
	if nativeInit.err != nil {
		return nil
	}
	// Memory sources explicitly opt in. Older file-backed faces can also
	// use CoreText, after checking that they don't carry custom variations.
	// Such faces are already parsed by their caller; do not retain that parsed
	// Go font after registering the native descriptor.
	for _, f := range collection {
		if _, ok := f.Face.(nativeFontData); ok {
			continue
		}
		if _, ok := f.Face.(gioot.Face); ok {
			continue
		}
		if _, ok := f.Face.(giofont.FaceSource); !ok {
			return nil
		}
		face := f.Face.Face()
		if face == nil {
			return nil
		}
		for _, coordinate := range face.Coords() {
			if coordinate != 0 {
				return nil
			}
		}
	}

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	pool := ct.poolPush()
	defer ct.poolPop(pool)
	e := &ctShaper{resources: &ctResources{owned: make(map[uintptr]int)}, registered: make(map[string][]ctRegistered), primary: make(map[ctStyle]uintptr), byHash: make(map[uint][]int)}
	// Share immutable source descriptors; each shaper owns distinct styles,
	// glyph IDs and scratch. Keeping the source owner alive prevents a native
	// source cache hit from outliving its copied/borrowed bytes.
	sources := make(map[ctFontSourceKey]*ctFontSource)

	for _, f := range collection {
		var data []byte
		var index int
		var path string
		if source, ok := f.Face.(giofont.FaceSource); ok {
			path, index = source.Source()
		}
		var loaders []*ot.Loader
		var err error
		var file *os.File
		key := ctFontSourceKey{path: path}
		if path != "" {
			file, err = os.Open(path)
			if err == nil {
				loaders, err = ot.NewLoaders(file)
			}
		} else {
			if source, ok := f.Face.(nativeFontData); ok {
				data, index = source.NativeFontData()
			} else if source, ok := f.Face.(rawFontData); ok {
				data, index = source.FontData()
			}
			if len(data) > 0 {
				loaders, err = ot.NewLoaders(ot.NewShared(data))
				key = ctFontSourceKey{data: uintptr(unsafe.Pointer(&data[0])), length: len(data)}
			}
		}
		if file != nil {
			defer file.Close()
		}
		if err != nil || index < 0 || index >= len(loaders) {
			e.resources.release()
			return nil
		}
		raw, err := loaders[index].RawTable(ot.MustNewTag("name"))
		if err != nil {
			e.resources.release()
			return nil
		}
		names, _, err := tables.ParseName(raw)
		if err != nil {
			e.resources.release()
			return nil
		}
		wanted := names.Name(6)
		if wanted == "" {
			e.resources.release()
			return nil
		}
		source := sources[key]
		if source == nil {
			source = sharedCTFontSource(key, data)
			if source != nil {
				sources[key] = source
				e.sourceOwners = append(e.sourceOwners, source)
			}
		}
		var array uintptr
		if source != nil {
			array = source.array
		}

		var matched uintptr
		if array != 0 {
			for i := 0; i < ct.arrayGetCount(array); i++ {
				d := ct.arrayGetValueAtIndex(array, i)
				n := ct.descriptorAttribute(d, nativeInit.nameKey)
				if n != 0 {
					same := ctString(n) == wanted
					ct.release(n)
					if same {
						matched = d
						break
					}
				}
			}
		}
		if matched == 0 {
			e.resources.release()
			return nil
		}
		family := strings.ToLower(string(f.Font.Typeface))
		e.registered[family] = append(e.registered[family], ctRegistered{e.resources.own(ct.retain(matched)), f.Font})
		e.defaults = append(e.defaults, string(f.Font.Typeface))
	}
	for i := range 2 {
		d := int8(i)
		setting := ctParagraphStyleSetting{spec: ctBaseWritingDirectionSpec, size: 1, value: unsafe.Pointer(&d)}
		e.styles[i] = e.resources.own(ct.paragraphStyleCreate(&setting, 1))
		runtime.KeepAlive(&d)
	}
	runtime.AddCleanup(e, func(r *ctResources) { r.release() }, e.resources)
	return e
}
func ctString(s uintptr) string {
	n := ct.stringGetLength(s)
	if n <= 0 {
		return ""
	}
	u := make([]uint16, n)
	ct.stringGetCharacters(s, cfRange{0, n}, &u[0])
	return string(utf16.Decode(u))
}
func ctStringNew(s string) uintptr {
	b := []byte(s)
	if len(b) == 0 {
		return ct.stringWithBytes(0, nil, 0, cfStringEncodingUTF8, false)
	}
	return ct.stringWithBytes(0, &b[0], len(b), cfStringEncodingUTF8, false)
}
func ctDict(keys, values []uintptr) uintptr {
	return ct.dictionaryCreate(0, &keys[0], &values[0], len(keys), ct.keyCallbacks, ct.valueCallbacks)
}
func ctFloat(v float64) uintptr       { return ct.numberCreate(0, cfNumberFloat64Type, unsafe.Pointer(&v)) }
func ctFixed(v float64) fixed.Int26_6 { return fixed.Int26_6(math.Round(v * 64)) }
func (e *ctShaper) setQuery(f giofont.Font, families []string) {
	e.query = f
	if len(families) == 0 {
		families = e.defaults
	}
	e.families = append(e.families[:0], families...)
}

// Same CSS-to-native weight scale and mandatory-family matching as MyGo.
func ctWeight(w giofont.Weight) float64 {
	steps := [...]float64{-.8, -.6, -.4, 0, .23, .3, .4, .56, .62}
	f := (float64(max(100, min(int(w)+400, 900))) - 100) / 100
	i := min(int(f), 7)
	return steps[i] + (steps[i+1]-steps[i])*(f-float64(i))
}
func (e *ctShaper) descriptor(family string) uintptr {
	if registered := e.registered[strings.ToLower(family)]; len(registered) > 0 {
		best := registered[0]
		score := int(^uint(0) >> 1)
		for _, r := range registered {
			d := absInt(int(r.meta.Weight - e.query.Weight))
			if r.meta.Style != e.query.Style {
				d += 10000
			}
			if d < score {
				best, score = r, d
			}
		}
		return ct.retain(best.desc)
	}
	switch strings.ToLower(family) {
	case "sans-serif", "system-ui":
		family = ".AppleSystemUIFont"
	case "monospace":
		family = "Menlo"
	case "serif":
		family = "Times New Roman"
	case "emoji":
		family = "Apple Color Emoji"
	}
	name := ctStringNew(family)
	defer ct.release(name)
	w := ctFloat(ctWeight(e.query.Weight))
	defer ct.release(w)
	symbolic := int32(0)
	if e.query.Style == giofont.Italic {
		symbolic = ctFontItalicTrait
	}
	sym := ct.numberCreate(0, cfNumberSInt32Type, unsafe.Pointer(&symbolic))
	defer ct.release(sym)
	traits := ctDict([]uintptr{ct.weightTrait, ct.symbolicTrait}, []uintptr{w, sym})
	defer ct.release(traits)
	attrs := ctDict([]uintptr{ct.familyNameAttribute, ct.traitsAttribute}, []uintptr{name, traits})
	defer ct.release(attrs)
	desc := ct.descriptorWithAttrs(attrs)
	if desc == 0 {
		return 0
	}
	defer ct.release(desc)
	mandatory := ct.familyNameAttribute
	set := ct.setCreate(0, &mandatory, 1, ct.setCallbacks)
	defer ct.release(set)
	return ct.descriptorMatching(desc, set)
}
func absInt(i int) int {
	if i < 0 {
		return -i
	}
	return i
}

// Match MyGo's NSFont system-font selection, including native weight/italic
// coordinates; a family-name guess cannot reproduce those variable instances.
func (e *ctShaper) systemFont(size fixed.Int26_6, mono bool) uintptr {
	cls := objc.ID(objc.GetClass("NSFont"))
	var native uintptr
	if cls != 0 {
		sel := "systemFontOfSize:weight:"
		if mono {
			sel = "monospacedSystemFontOfSize:weight:"
		}
		selector := objc.RegisterName(sel)
		if objc.Send[bool](cls, objc.RegisterName("respondsToSelector:"), selector) {
			native = uintptr(cls.Send(selector, float64(size)/64, ctWeight(e.query.Weight)))
		}
	}
	if native == 0 {
		native = ct.uiFontForLanguage(ctFontUIFontSystem, float64(size)/64, 0)
	} else {
		ct.retain(native)
	}
	if native != 0 && e.query.Style == giofont.Italic {
		if with := ct.fontWithTraits(native, 0, 0, ctFontItalicTrait, ctFontItalicTrait); with != 0 {
			ct.release(native)
			native = with
		}
	}
	return native
}

func (e *ctShaper) primaryFont(size fixed.Int26_6) uintptr {
	key := ctStyle{strings.Join(e.families, "\x00"), size, e.query.Weight, e.query.Style}
	if f := e.primary[key]; f != 0 {
		return f
	}
	if len(e.primary) >= 256 {
		for _, f := range e.primary {
			e.resources.drop(f)
		}
		clear(e.primary)
	}
	var font uintptr
	var cascade []uintptr
	for _, family := range e.families {
		if font == 0 {
			switch strings.ToLower(family) {
			case "system-ui", "sans-serif":
				font = e.systemFont(size, false)
				continue
			case "monospace":
				font = e.systemFont(size, true)
				continue
			}
		}
		if d := e.descriptor(family); d != 0 {
			if font == 0 {
				font = ct.fontWithDescriptor(d, float64(size)/64, 0)
				ct.release(d)
			} else {
				cascade = append(cascade, d)
			}
		}
	}
	if font == 0 {
		font = e.systemFont(size, false)
	}
	if font == 0 {
		return 0
	}
	if len(cascade) > 0 {
		array := ct.arrayCreate(0, &cascade[0], len(cascade), ct.arrayCallbacks)
		attrs := ctDict([]uintptr{ct.cascadeList}, []uintptr{array})
		desc := ct.descriptorWithAttrs(attrs)
		if with := ct.fontWithAttrs(font, 0, 0, desc); with != 0 {
			ct.release(font)
			font = with
		}
		ct.release(desc)
		ct.release(attrs)
		ct.release(array)
		for _, d := range cascade {
			ct.release(d)
		}
	}
	e.primary[key] = e.resources.own(font)
	return font
}
func (e *ctShaper) faceOf(s *shaperImpl, native uintptr) *ctNativeFace {
	h := ct.hash(native)
	for _, i := range e.byHash[h] {
		f := e.faces[i]
		if ct.equal(f.font, native) {
			return f
		}
	}
	f := &ctNativeFace{font: e.resources.own(ct.retain(native)), face: &font.Face{Font: &font.Font{}}}
	e.byHash[h] = append(e.byHash[h], len(e.faces))
	e.faces = append(e.faces, f)
	s.addFace(f.face, e.query)
	return f
}
func (e *ctShaper) shape(s *shaperImpl, size fixed.Int26_6, locale system.Locale, txt []rune) []shaping.Output {
	defer runtime.KeepAlive(e)
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	pool := ct.poolPush()
	defer ct.poolPop(pool)
	e.finishShape()
	e.direction = locale.Direction

	if s.faceToIndex == nil {
		s.faceToIndex = make(map[*font.Font]int)
	}
	// CoreText treats zero size as its default; Gio requires zero ink/metrics.
	scale := 1.
	nativeSize := size
	if size == 0 {
		scale = 0
		nativeSize = fixed.I(1)
	}
	toFixed := func(v float64) fixed.Int26_6 { return ctFixed(v * scale) }
	native := e.primaryFont(nativeSize)
	if native == 0 {
		return nil
	}
	e.u16 = e.u16[:0]
	e.indices = e.indices[:0]
	e.at = e.at[:0]
	for i, r := range txt {
		e.at = append(e.at, len(e.u16))
		if r >= 0x10000 {
			a, b := utf16.EncodeRune(r)
			e.u16 = append(e.u16, uint16(a), uint16(b))
			e.indices = append(e.indices, i, i)
		} else {
			e.u16 = append(e.u16, uint16(r))
			e.indices = append(e.indices, i)
		}
	}
	e.indices = append(e.indices, len(txt))
	e.at = append(e.at, len(e.u16))
	if len(txt) == 0 {
		f := e.faceOf(s, native)
		return []shaping.Output{{Face: f.face, Size: size, Direction: mapDirection(locale.Direction), LineBounds: shaping.Bounds{Ascent: toFixed(ct.fontGetAscent(native)), Descent: -toFixed(ct.fontGetDescent(native)), Gap: toFixed(ct.fontGetLeading(native))}}}
	}
	str := ct.stringWithCharacters(0, &e.u16[0], len(e.u16))
	defer ct.release(str)
	dir := 0
	if locale.Direction.Progression() == system.TowardOrigin {
		dir = 1
	}
	attrs := ctDict([]uintptr{ct.fontAttributeName, ct.paragraphStyleName}, []uintptr{native, e.styles[dir]})
	defer ct.release(attrs)
	attributed := ct.attributedString(0, str, attrs)
	defer ct.release(attributed)
	setter := ct.typesetterCreate(attributed)
	e.active = e.resources.own(setter)
	line := ct.typesetterCreateLine(setter, cfRange{0, len(e.u16)})
	defer ct.release(line)
	runs := ct.lineGetGlyphRuns(line)
	outputs := s.outScratchBuf[:0]
	for i := 0; i < ct.arrayGetCount(runs); i++ {
		run := ct.arrayGetValueAtIndex(runs, i)
		count := ct.runGetGlyphCount(run)
		if count == 0 {
			continue
		}
		f := e.faceOf(s, ct.dictionaryGetValue(ct.runGetAttributes(run), ct.fontAttributeName))
		rr := ct.runGetStringRange(run)
		start, end := e.indices[rr.location], e.indices[rr.location+rr.length]
		out := shaping.Output{Face: f.face, Size: size, Direction: mapDirection(system.LTR), Runes: shaping.Range{Offset: start, Count: end - start}, LineBounds: shaping.Bounds{Ascent: toFixed(ct.fontGetAscent(f.font)), Descent: -toFixed(ct.fontGetDescent(f.font)), Gap: toFixed(ct.fontGetLeading(f.font))}}
		if ct.runGetStatus(run)&ctRunStatusRightToLeft != 0 {
			out.Direction = mapDirection(system.RTL)
		}
		e.ids = slices.Grow(e.ids[:0], count)[:count]
		e.points = slices.Grow(e.points[:0], count)[:count]
		e.advances = slices.Grow(e.advances[:0], count)[:count]
		e.clusters = slices.Grow(e.clusters[:0], count)[:count]
		e.rects = slices.Grow(e.rects[:0], count)[:count]
		ct.runGetGlyphs(run, cfRange{}, &e.ids[0])
		ct.runGetPositions(run, cfRange{}, &e.points[0])
		ct.runGetAdvances(run, cfRange{}, &e.advances[0])
		ct.runGetStringIndices(run, cfRange{}, &e.clusters[0])
		ct.fontGetBoundingRects(f.font, 0, &e.ids[0], &e.rects[0], count)
		out.Glyphs = make([]shaping.Glyph, count)
		// CoreText runs are visually ordered. Translate whole-line positions back
		// to run-local pen offsets; Gio's wrapper retains its existing bidi order.
		origin, pen := e.points[0].x, 0.
		counts := make(map[int]int, count)
		for _, idx := range e.clusters {
			at := e.indices[max(0, min(idx, len(e.u16)))]
			counts[at]++
		}
		starts := make([]int, 0, len(counts))
		for at := range counts {
			starts = append(starts, at)
		}
		slices.Sort(starts)
		ends := make(map[int]int, len(starts))
		for j, at := range starts {
			next := end
			if j+1 < len(starts) {
				next = starts[j+1]
			}
			ends[at] = next
		}
		for j := range count {
			cluster := e.indices[max(0, min(e.clusters[j], len(e.u16)))]
			next, glyphCount := ends[cluster], counts[cluster]

			r := e.rects[j]
			advance := toFixed(pen+e.advances[j].w) - toFixed(pen)
			out.Glyphs[j] = shaping.Glyph{GlyphID: font.GID(e.ids[j]), ClusterIndex: cluster, RuneCount: next - cluster, GlyphCount: glyphCount, Advance: advance, XOffset: toFixed(origin + pen - e.points[j].x), YOffset: toFixed(e.points[j].y), XBearing: toFixed(r.x), YBearing: toFixed(r.y + r.h), Width: toFixed(r.w), Height: toFixed(-r.h)}
			pen += e.advances[j].w
			out.Advance += advance
		}
		out.GlyphBounds = out.LineBounds
		outputs = append(outputs, out)
	}
	slices.SortStableFunc(outputs, func(a, b shaping.Output) int { return a.Runes.Offset - b.Runes.Offset })
	s.outScratchBuf = outputs
	return outputs
}
func (e *ctShaper) glyphFont(id GlyphID) (uintptr, fixed.Int26_6, uint16, bool) {
	size, index, gid := splitGlyphID(id)
	if size <= 0 || index < 0 || index >= len(e.faces) || gid > 0xffff {
		return 0, 0, 0, false
	}
	return e.faces[index].font, size, uint16(gid), true
}
func (e *ctShaper) glyphFile(id GlyphID) (string, int, int, fixed.Int26_6, uint16, bool) {
	native, size, gid, ok := e.glyphFont(id)
	if !ok {
		return "", 0, 0, 0, 0, false
	}
	_, index, _ := splitGlyphID(id)
	f := e.faces[index]
	if !f.located {
		f.located = true
		d := nativeInit.fontDescriptor(native)
		defer ct.release(d)
		u := ct.descriptorAttribute(d, nativeInit.urlKey)
		if u != 0 {
			p := nativeInit.urlPath(u, 0)
			path := ctString(p)
			ct.release(p)
			ct.release(u)
			name := nativeInit.fontName(native)
			wanted := ctString(name)
			ct.release(name)
			file, err := os.Open(path)
			if err == nil {
				defer file.Close()
				loaders, err := ot.NewLoaders(file)
				if err == nil {
					for i, ld := range loaders {
						raw, err := ld.RawTable(ot.MustNewTag("name"))
						if err != nil {
							continue
						}
						names, _, err := tables.ParseName(raw)
						if err != nil {
							continue
						}
						if names.Name(6) == wanted {
							f.location, f.index = path, i
							break
						}
						raw, err = ld.RawTable(ot.MustNewTag("fvar"))
						if err == nil {
							fv, _, err := tables.ParseFvar(raw)
							if err == nil {
								for j, instance := range fv.Instances {
									if instance.PostScriptNameID != 0xffff && instance.PostScriptNameID != 0 && names.Name(tables.NameID(instance.PostScriptNameID)) == wanted {
										f.location, f.index, f.instance = path, i, j+1
										break
									}
								}
							}
						}
					}
				}
			}
		}
	}
	return f.location, f.index, f.instance, size, gid, f.location != ""
}

// A single process-wide callback is serialized while CGPathApply calls back
// synchronously. Native code sees an integer token, never a retained Go pointer.
var ctPaths struct {
	sync.Mutex
	once     sync.Once
	callback uintptr
	builder  *clip.Path
	origin   f32.Point
}

func (e *ctShaper) outline(ops *op.Ops, gs []Glyph) clip.PathSpec {
	defer runtime.KeepAlive(e)
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	pool := ct.poolPush()
	defer ct.poolPop(pool)
	var p clip.Path
	p.Begin(ops)
	ctPaths.Lock()
	defer ctPaths.Unlock()
	ctPaths.once.Do(func() {
		ctPaths.callback = purego.NewCallback(func(_ uintptr, element uintptr) {
			type pathElement struct {
				kind   int32
				points uintptr
			}
			el := *(**pathElement)(unsafe.Pointer(&element))
			n := [...]int{1, 1, 2, 3, 0}[el.kind]
			points := unsafe.Slice(*(**cgPoint)(unsafe.Pointer(&el.points)), n)
			at := func(i int) f32.Point { return f32.Pt(float32(points[i].x), -float32(points[i].y)).Add(ctPaths.origin) }
			switch el.kind {
			case 0:
				ctPaths.builder.MoveTo(at(0))
			case 1:
				ctPaths.builder.LineTo(at(0))
			case 2:
				ctPaths.builder.QuadTo(at(0), at(1))
			case 3:
				ctPaths.builder.CubeTo(at(0), at(1), at(2))
			case 4:
				ctPaths.builder.Close()
			}
		})
	})
	ctPaths.builder = &p
	defer func() { ctPaths.builder = nil }()
	for _, g := range gs {
		native, _, gid, ok := e.glyphFont(g.ID)
		if !ok || ct.fontGetSymbolicTraits(native)&ctFontColorGlyphsTrait != 0 {
			continue
		}
		path := ct.fontCreatePath(native, gid, 0)
		if path == 0 {
			continue
		}
		ctPaths.origin = f32.Pt(fixedToFloat(g.X-gs[0].X-g.Offset.X), -fixedToFloat(g.Offset.Y))
		ct.pathApply(path, 1, ctPaths.callback)
		ct.pathRelease(path)
	}
	return p.End()
}
func (e *ctShaper) bitmaps(ops *op.Ops, gs []Glyph) op.CallOp {
	defer runtime.KeepAlive(e)
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	pool := ct.poolPush()
	defer ct.poolPop(pool)
	record := op.Record(ops)
	for _, g := range gs {
		native, _, gid, ok := e.glyphFont(g.ID)
		if !ok || ct.fontGetSymbolicTraits(native)&ctFontColorGlyphsTrait == 0 {
			continue
		}
		var r cgRect
		ct.fontGetBoundingRects(native, 0, &gid, &r, 1)
		bounds := image.Rect(int(math.Floor(r.x))-1, int(math.Floor(-r.y-r.h))-1, int(math.Ceil(r.x+r.w))+1, int(math.Ceil(-r.y))+1)
		if bounds.Empty() {
			continue
		}
		cached, ok := e.bitmapsCache.Get(g.ID)
		if !ok {
			img := image.NewRGBA(image.Rectangle{Max: bounds.Size()})
			space := ct.colorSpaceDeviceRGB()
			ctx := ct.bitmapContextCreate(unsafe.Pointer(&img.Pix[0]), img.Rect.Dx(), img.Rect.Dy(), 8, img.Stride, space, cgImageAlphaPremultipliedLst|cgBitmapByteOrder32Big)
			ct.release(space)
			if ctx == 0 {
				continue
			}
			pos := cgPoint{float64(-bounds.Min.X), float64(bounds.Max.Y)}
			ct.fontDrawGlyphs(native, &gid, &pos, 1, ctx)
			ct.contextRelease(ctx)
			cached = bitmap{img: paint.NewImageOp(img), size: bounds.Size()}
			e.bitmapsCache.Put(g.ID, cached)
		}
		tr := op.Affine(f32.AffineId().Offset(f32.Pt(fixedToFloat(g.X-gs[0].X-g.Offset.X)+float32(bounds.Min.X), -fixedToFloat(g.Offset.Y)+float32(bounds.Min.Y)))).Push(ops)
		cl := clip.Rect(image.Rectangle{Max: cached.size}).Push(ops)
		cached.img.Add(ops)
		paint.PaintOp{}.Add(ops)
		cl.Pop()
		tr.Pop()
	}
	return record.Stop()
}

func (e *ctShaper) finishShape() {
	if e.active != 0 {
		e.resources.drop(e.active)
		e.active = 0
	}
}

// CoreText has already resolved nested bidi levels. Gio's legacy wrapper only
// has LTR/RTL flags, which cannot distinguish numbers nested in an RTL span.
// Resolve each wrapped line's visual ranks from that paragraph's typesetter.
func (e *ctShaper) orderLine(line shaping.Line, hasTruncator bool) {
	defer runtime.KeepAlive(e)
	count := len(line)
	if hasTruncator {
		count--
	}
	if count <= 0 {
		if len(line) > 0 {
			line[0].VisualIndex = 0
		}
		return
	}
	simple := e.direction == system.LTR
	for _, run := range line[:count] {
		simple = simple && unmapDirection(run.Direction) == system.LTR
	}
	if simple {
		for i := range line {
			line[i].VisualIndex = int32(i)
		}
		return
	}
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	pool := ct.poolPush()
	defer ct.poolPop(pool)
	for i := range line {
		line[i].VisualIndex = int32(i)
		if e.direction == system.RTL {
			line[i].VisualIndex = int32(len(line) - i - 1)
		}
	}
	start, end := line[0].Runes.Offset, line[count-1].Runes.Offset+line[count-1].Runes.Count
	if e.active == 0 || start < 0 || end >= len(e.at) {
		return
	}
	nativeLine := ct.typesetterCreateLine(e.active, cfRange{e.at[start], e.at[end] - e.at[start]})
	if nativeLine == 0 {
		return
	}
	defer ct.release(nativeLine)
	runs := ct.lineGetGlyphRuns(nativeLine)
	type nativeRun struct {
		start, end, visual int
		rtl                bool
	}
	nativeRuns := make([]nativeRun, 0, ct.arrayGetCount(runs))
	for i := 0; i < ct.arrayGetCount(runs); i++ {
		run := ct.arrayGetValueAtIndex(runs, i)
		rr := ct.runGetStringRange(run)
		nativeRuns = append(nativeRuns, nativeRun{rr.location, rr.location + rr.length, i, ct.runGetStatus(run)&ctRunStatusRightToLeft != 0})
	}
	type ranked struct {
		logical, rank int
		rtl           bool
	}
	order := make([]ranked, count)
	for i, run := range line[:count] {
		unit := e.at[run.Runes.Offset]
		order[i] = ranked{i, i, false}
		for _, r := range nativeRuns {
			if unit >= r.start && unit < r.end {
				order[i] = ranked{i, r.visual, r.rtl}
				break
			}
		}
	}
	slices.SortStableFunc(order, func(a, b ranked) int {
		if a.rank != b.rank {
			return a.rank - b.rank
		}
		if a.rtl {
			return b.logical - a.logical
		}
		return a.logical - b.logical
	})
	shift := 0
	if hasTruncator {
		if e.direction == system.RTL {
			line[count].VisualIndex = 0
			shift = 1
		} else {
			line[count].VisualIndex = int32(count)
		}
	}
	for i, r := range order {
		line[r.logical].VisualIndex = int32(i + shift)
	}
}
