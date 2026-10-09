package harfbuzz

import (
	"fmt"

	"github.com/go-text/typesetting/font/opentype/tables"
)

// ported from harfbuzz/src/hb-shape.cc, harfbuzz/src/hb-shape-plan.cc Copyright © 2009, 2012 Behdad Esfahbod

/**
 * Shaping is the central operation of HarfBuzz. Shaping operates on buffers,
 * which are sequences of Unicode characters that use the same font and have
 * the same text direction, script, and language. After shaping the buffer
 * contains the output glyphs and their positions.
 **/

// Shape shapes the buffer using `font`, turning its Unicode characters content to
// positioned glyphs. If `features` is not empty, it will be used to control the
// features applied during shaping. If two features have the same tag but
// overlapping ranges the value of the feature with the higher index takes
// precedence.
//
// The shapping plan depends on the font capabilities. See `NewFont` and `Face` and
// its extension interfaces for more details.
//
// It also depends on the properties of the segment of text : the `Props`
// field of the buffer must be set before calling `Shape`.
func (b *Buffer) Shape(font *Font, features []Feature) {
	shapePlan := b.newShapePlanCached(font, b.Props, features, font.varCoords())
	shapePlan.execute(font, b, features)
}

// Shape plans are an internal mechanism. Each plan contains state
// describing how HarfBuzz will shape a particular text segment, based on
// the combination of segment properties and the capabilities in the
// font face in use.
//
// Shape plans are not used for shaping directly, but can be queried to
// access certain information about how shaping will perform, given a set
// of specific input parameters (script, language, direction, features,
// etc.).
//
// Most client programs will not need to deal with shape plans directly.
type shapePlan struct {
	shaper       shaperOpentype
	props        SegmentProperties
	userFeatures []Feature
}

func (plan *shapePlan) init(copy bool, font *Font, props SegmentProperties,
	userFeatures []Feature, coords []tables.Coord,
) {
	plan.props = props
	if !copy {
		plan.userFeatures = userFeatures
	} else {
		plan.userFeatures = append([]Feature(nil), userFeatures...)
		/* Make start/end uniform to easier catch bugs. */
		for i := range plan.userFeatures {
			if plan.userFeatures[i].Start != FeatureGlobalStart {
				plan.userFeatures[i].Start = 1
			}
			if plan.userFeatures[i].End != FeatureGlobalEnd {
				plan.userFeatures[i].End = 2
			}
		}
	}

	// init shaper
	plan.shaper.init(font.face.Font, coords)
}

func userFeaturesMatch(features, other []Feature) bool {
	if len(features) != len(other) {
		return false
	}
	for i, feat := range features {
		if feat.Tag != other[i].Tag || feat.Value != other[i].Value ||
			(feat.Start == FeatureGlobalStart && feat.End == FeatureGlobalEnd) !=
				(other[i].Start == FeatureGlobalStart && other[i].End == FeatureGlobalEnd) {
			return false
		}
	}
	return true
}

// matches reports whether plan suits props and userFeatures.
//
// Keel patch: upstream built a whole plan as the key, initializing its
// shaper for the font, and compared plans by value, copying each.
func (plan *shapePlan) matches(props SegmentProperties, userFeatures []Feature) bool {
	return plan.props == props && userFeaturesMatch(plan.userFeatures, userFeatures)
}

// Constructs a shaping plan for a combination of @face, @userFeatures, @props,
// plus the variation-space coordinates @coords.
// See newShapePlanCached for caching support.
func newShapePlan(font *Font, props SegmentProperties,
	userFeatures []Feature, coords []tables.Coord,
) *shapePlan {
	if debugMode {
		fmt.Printf("NEW SHAPE PLAN: face:%p features:%v coords:%v\n", &font.face, userFeatures, coords)
	}

	var sp shapePlan

	sp.init(true, font, props, userFeatures, coords)

	if debugMode {
		fmt.Println("NEW SHAPE PLAN - compiling shaper plan for script", props.Script)
	}
	sp.shaper.compile(props, userFeatures)

	return &sp
}

// Executes the given shaping plan on the specified `buffer`, using
// the given `font` and `features`.
func (sp *shapePlan) execute(font *Font, buffer *Buffer, features []Feature) {
	if debugMode {
		fmt.Printf("EXECUTE shape plan %p features:%v shaper:%T\n", sp, features, sp.shaper.plan.shaper)
	}

	sp.shaper.shape(font, buffer, features)
}

/*
 * Caching
 */

// creates (or returns) a cached shaping plan suitable for reuse, for a combination
// of `face`, `userFeatures`, `props`, plus the variation-space coordinates `coords`.
func (b *Buffer) newShapePlanCached(font *Font, props SegmentProperties,
	userFeatures []Feature, coords []tables.Coord,
) *shapePlan {
	plans := b.planCache[font.face]

	for _, plan := range plans {
		if plan.matches(props, userFeatures) {
			if debugMode {
				fmt.Printf("\tPLAN %p fulfilled from cache\n", plan)
			}
			return plan
		}
	}
	plan := newShapePlan(font, props, userFeatures, coords)

	plans = append(plans, plan)
	b.planCache[font.face] = plans

	if debugMode {
		fmt.Printf("\tPLAN %p inserted into cache\n", plan)
	}

	return plan
}
