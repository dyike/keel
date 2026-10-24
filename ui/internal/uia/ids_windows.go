//go:build windows

package uia

import "golang.org/x/sys/windows"

// Interface IDs, from UIAutomationCore.h; the test checks them, and each
// interface's method order, against the Windows SDK where it is installed.
var (
	iidUnknown            = windows.GUID{Data1: 0x00000000, Data2: 0x0000, Data3: 0x0000, Data4: [8]byte{0xc0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46}}
	iidSimple             = windows.GUID{Data1: 0xd6dd68d1, Data2: 0x86fd, Data3: 0x4332, Data4: [8]byte{0x86, 0x66, 0x9a, 0xbe, 0xde, 0xa2, 0xd2, 0x4c}}
	iidFragment           = windows.GUID{Data1: 0xf7063da8, Data2: 0x8359, Data3: 0x439c, Data4: [8]byte{0x92, 0x97, 0xbb, 0xc5, 0x29, 0x9a, 0x7d, 0x87}}
	iidFragmentRoot       = windows.GUID{Data1: 0x620ce2a5, Data2: 0xab8f, Data3: 0x40a9, Data4: [8]byte{0x86, 0xcb, 0xde, 0x3c, 0x75, 0x59, 0x9b, 0x58}}
	iidInvokeProvider     = windows.GUID{Data1: 0x54fcb24b, Data2: 0xe18e, Data3: 0x47a2, Data4: [8]byte{0xb4, 0xd3, 0xec, 0xcb, 0xe7, 0x75, 0x99, 0xa2}}
	iidToggleProvider     = windows.GUID{Data1: 0x56d00bd0, Data2: 0xc4f4, Data3: 0x433c, Data4: [8]byte{0xa8, 0x36, 0x1a, 0x52, 0xa5, 0x7e, 0x08, 0x92}}
	iidScrollItemProvider = windows.GUID{Data1: 0x2360c714, Data2: 0x4bf1, Data3: 0x4b26, Data4: [8]byte{0xba, 0x65, 0x9b, 0x21, 0x31, 0x61, 0x27, 0xeb}}
	iidValueProvider      = windows.GUID{Data1: 0xc7935180, Data2: 0x6fb3, Data3: 0x4201, Data4: [8]byte{0xb1, 0x74, 0x7d, 0xf7, 0x3a, 0xdb, 0xf6, 0x4a}}
)

// Pattern, property, control type and event IDs, from UIAutomationClient.h.
const (
	patternInvoke     = 10000
	patternValue      = 10002
	patternToggle     = 10015
	patternScrollItem = 10017

	propRuntimeID            = 30000
	propControlType          = 30003
	propLocalizedControlType = 30004
	propName                 = 30005
	propHasKeyboardFocus     = 30008
	propIsKeyboardFocusable  = 30009
	propIsEnabled            = 30010
	propAutomationID         = 30011
	propClassName            = 30012
	propIsControlElement     = 30016
	propIsContentElement     = 30017
	propIsOffscreen          = 30022
	propFrameworkID          = 30024
	propItemStatus           = 30026
	propValueValue           = 30045
	propValueIsReadOnly      = 30046
	propToggleState          = 30086

	ctButton      = 50000
	ctCheckBox    = 50002
	ctComboBox    = 50003
	ctEdit        = 50004
	ctHyperlink   = 50005
	ctImage       = 50006
	ctListItem    = 50007
	ctList        = 50008
	ctMenu        = 50009
	ctMenuItem    = 50011
	ctProgressBar = 50012
	ctRadioButton = 50013
	ctSlider      = 50015
	ctStatusBar   = 50017
	ctTab         = 50018
	ctTabItem     = 50019
	ctText        = 50020
	ctToolBar     = 50021
	ctToolTip     = 50022
	ctTree        = 50023
	ctTreeItem    = 50024
	ctGroup       = 50026
	ctDataItem    = 50029
	ctPane        = 50033
	ctHeaderItem  = 50035
	ctTable       = 50036
	ctSeparator   = 50038

	// UiaRootObjectId, the object WM_GETOBJECT asks for.
	rootObjectID = -25
	// UiaAppendRuntimeId: the runtime ID continues the host's.
	appendRuntimeID = 3

	providerServerSide = 2 // ProviderOptions_ServerSideProvider

	navParent          = 0
	navNextSibling     = 1
	navPreviousSibling = 2
	navFirstChild      = 3
	navLastChild       = 4

	toggleOff           = 0
	toggleOn            = 1
	toggleIndeterminate = 2

	structureChildrenInvalidated = 2

	eventFocusChanged = 20005 // UIA_AutomationFocusChangedEventId
)

// HRESULTs.
const (
	sOK                    = 0
	eNoInterface           = 0x80004002
	eInvalidArg            = 0x80070057
	eAccessDenied          = 0x80070005
	uiaElementNotAvailable = 0x80040201
)

// VARIANT types.
const (
	vtEmpty = 0
	vtI4    = 3
	vtBSTR  = 8
	vtBool  = 11
)
