package theme

// The size scales. Components pick a step instead of a number of their own,
// so the same role looks the same everywhere and a design change is made
// here. The values are plain numbers for el, which takes dp and sp as float32.

// Radius steps, in dp. RadiusFull makes a pill or a circle: el clamps a
// radius to half the shorter side.
const (
	RadiusSm   = 4  // small marks: links, keycaps, compact chips
	RadiusMd   = 6  // controls: buttons, fields, menu items
	RadiusLg   = 8  // cards, popovers, menus, navigation rows
	RadiusXl   = 12 // dialogs, sheets, chat bubbles
	RadiusFull = 9999
)

// Text size steps, in sp. BodySize, SmallSize and HeadingSize are TextBody,
// TextMd and TextHeading.
const (
	TextXs      = 11 // axis labels, badges
	TextSm      = 12 // captions, errors, group titles
	TextMd      = 13 // secondary text
	TextControl = 14 // buttons, tabs, toggles
	TextBody    = 15 // body text
	TextLg      = 17 // panel and dialog titles
	TextXl      = 20 // page titles
	TextHeading = 22
)

// Spacing steps, in dp, for gaps, padding and margins. Most layouts need
// only Xs to Xl; off-scale values are for optical adjustments such as
// centering an icon.
const (
	SpaceXxs = 2  // hairline gaps: stacked labels, tight icon pairs
	SpaceXs  = 4  // inside compact controls, between a label and its hint
	SpaceSm  = 6  // between an icon and its text
	SpaceMd  = 8  // between controls in a row, list rows
	SpaceLg  = 12 // control padding, between form fields
	SpaceXl  = 16 // card padding, between groups
	Space2xl = 24 // dialog and page padding, between sections
	Space3xl = 32 // between page regions
)

// Elevation is a soft shadow under a raised surface, in dp: Offset moves it
// down, Blur is how far it fades out. Its color is Shadow.
type Elevation struct{ Offset, Blur float32 }

// Elevation steps: menus and popovers float a little, dialogs more.
var (
	ElevationSm = Elevation{Offset: 1, Blur: 3}   // cards that lift on hover
	ElevationMd = Elevation{Offset: 4, Blur: 12}  // popovers, menus, dropdowns, toasts
	ElevationLg = Elevation{Offset: 12, Blur: 32} // dialogs, sheets, command palette
)
