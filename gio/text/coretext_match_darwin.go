// SPDX-License-Identifier: Unlicense OR MIT
//go:build darwin && !ios

package text

import (
	"strings"
	"unsafe"

	giofont "github.com/dyike/keel/gio/font"
	"github.com/ebitengine/purego/objc"
	"golang.org/x/image/math/fixed"
)

// CSS weights correspond to AppKit's nine named weights. Use the framework's
// exported values, including interpolation for the intermediate Gio weights.
func ctWeight(weight giofont.Weight) float64 {
	css := min(900, max(100, int(weight)+400)) - 100
	lower := css / 100
	if lower == len(ct.weights)-1 {
		return ct.weights[lower]
	}
	fraction := float64(css%100) / 100
	return (1-fraction)*ct.weights[lower] + fraction*ct.weights[lower+1]
}

func (e *ctShaper) descriptor(family string) uintptr {
	faces := e.registered[strings.ToLower(family)]
	if len(faces) != 0 {
		// Registered sources take precedence over the system database. Select
		// the requested style before comparing weight, retaining stable ties.
		best := 0
		for index := 1; index < len(faces); index++ {
			a, b := faces[index].meta, faces[best].meta
			aStyle, bStyle := a.Style == e.query.Style, b.Style == e.query.Style
			if aStyle && !bStyle || aStyle == bStyle && fontWeightDistance(a.Weight, e.query.Weight) < fontWeightDistance(b.Weight, e.query.Weight) {
				best = index
			}
		}
		return ct.retain(faces[best].desc)
	}
	if alias, ok := nativeFamilyAliases[strings.ToLower(family)]; ok {
		family = alias
	}
	name := ctStringNew(family)
	weight := ctFloat(ctWeight(e.query.Weight))
	italic := int32(0)
	if e.query.Style == giofont.Italic {
		italic = ctFontItalicTrait
	}
	style := ct.numberCreate(0, cfNumberSInt32Type, unsafe.Pointer(&italic))
	traits := ctDict([]uintptr{ct.symbolicTrait, ct.weightTrait}, []uintptr{style, weight})
	attributes := ctDict([]uintptr{ct.traitsAttribute, ct.familyNameAttribute}, []uintptr{traits, name})
	request := ct.descriptorWithAttrs(attributes)
	ct.release(attributes)
	ct.release(traits)
	ct.release(style)
	ct.release(weight)
	ct.release(name)
	if request == 0 {
		return 0
	}
	// Requiring the family prevents an unknown first choice from silently
	// selecting a default font before the next explicit family is tried.
	key := ct.familyNameAttribute
	required := ct.setCreate(0, &key, 1, ct.setCallbacks)
	matched := ct.descriptorMatching(request, required)
	ct.release(required)
	ct.release(request)
	return matched
}

var nativeFamilyAliases = map[string]string{
	"system-ui":  ".AppleSystemUIFont",
	"sans-serif": ".AppleSystemUIFont",
	"monospace":  "Menlo",
	"serif":      "Times New Roman",
	"emoji":      "Apple Color Emoji",
}

func fontWeightDistance(a, b giofont.Weight) int {
	if a > b {
		return int(a) - int(b)
	}
	return int(b) - int(a)
}

// AppKit creates the actual system variable-font instance. Retain its result
// before leaving the autorelease pool; symbolic copying adds an italic request.
func (e *ctShaper) systemFont(size fixed.Int26_6, mono bool) uintptr {
	method := "systemFontOfSize:weight:"
	if mono {
		method = "monospacedSystemFontOfSize:weight:"
	}
	class := objc.ID(objc.GetClass("NSFont"))
	selector := objc.RegisterName(method)
	var owned uintptr
	if class != 0 && objc.Send[bool](class, objc.RegisterName("respondsToSelector:"), selector) {
		if borrowed := class.Send(selector, float64(size)/64, ctWeight(e.query.Weight)); borrowed != 0 {
			owned = ct.retain(uintptr(borrowed))
		}
	}
	if owned == 0 {
		owned = ct.uiFontForLanguage(ctFontUIFontSystem, float64(size)/64, 0)
	}
	if owned == 0 || e.query.Style != giofont.Italic {
		return owned
	}
	italic := ct.fontWithTraits(owned, float64(size)/64, 0, ctFontItalicTrait, ctFontItalicTrait)
	if italic == 0 {
		return owned
	}
	ct.release(owned)
	return italic
}

func (e *ctShaper) primaryFont(size fixed.Int26_6) uintptr {
	style := ctStyle{families: strings.Join(e.families, "\x00"), size: size, weight: e.query.Weight, style: e.query.Style}
	if cached := e.primary[style]; cached != 0 {
		return cached
	}
	if len(e.primary) == 256 {
		for key, value := range e.primary {
			e.resources.drop(value)
			delete(e.primary, key)
		}
	}

	var selected uintptr
	cascade := make([]uintptr, 0, len(e.families))
	for _, family := range e.families {
		if selected == 0 {
			switch strings.ToLower(family) {
			case "system-ui", "sans-serif", "monospace":
				selected = e.systemFont(size, strings.EqualFold(family, "monospace"))
				if selected != 0 {
					continue
				}
			}
		}
		descriptor := e.descriptor(family)
		if descriptor == 0 {
			continue
		}
		if selected != 0 {
			cascade = append(cascade, descriptor)
		} else {
			selected = ct.fontWithDescriptor(descriptor, float64(size)/64, 0)
			ct.release(descriptor)
		}
	}
	if selected == 0 {
		selected = e.systemFont(size, false)
	}
	if len(cascade) > 0 {
		if selected != 0 {
			selected = attachFontCascade(selected, cascade)
		}
		for _, descriptor := range cascade {
			ct.release(descriptor)
		}
	}
	if selected != 0 {
		e.primary[style] = e.resources.own(selected)
	}
	return selected
}

func attachFontCascade(font uintptr, descriptors []uintptr) uintptr {
	array := ct.arrayCreate(0, &descriptors[0], len(descriptors), ct.arrayCallbacks)
	attributes := ctDict([]uintptr{ct.cascadeList}, []uintptr{array})
	cascade := ct.descriptorWithAttrs(attributes)
	updated := ct.fontWithAttrs(font, 0, 0, cascade)
	ct.release(cascade)
	ct.release(attributes)
	ct.release(array)
	if updated == 0 {
		return font
	}
	ct.release(font)
	return updated
}
