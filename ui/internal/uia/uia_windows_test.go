//go:build windows

package uia

import (
	"fmt"
	"image"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"unsafe"

	"golang.org/x/sys/windows"
)

// sdkHeader finds a Windows SDK header, newest first.
func sdkHeader(t *testing.T, name string) []byte {
	t.Helper()
	dirs, _ := filepath.Glob(`C:\Program Files (x86)\Windows Kits\10\Include\*\um`)
	sort.Sort(sort.Reverse(sort.StringSlice(dirs)))
	for _, dir := range dirs {
		if b, err := os.ReadFile(filepath.Join(dir, name)); err == nil {
			return b
		}
	}
	t.Skip(name + " not found")
	return nil
}

// The IIDs, method order and IDs match the Windows SDK.
func TestMatchesSDK(t *testing.T) {
	core := string(sdkHeader(t, "UIAutomationCore.h"))
	methods := map[string][]string{
		"IRawElementProviderSimple":       {"get_ProviderOptions", "GetPatternProvider", "GetPropertyValue", "get_HostRawElementProvider"},
		"IRawElementProviderFragment":     {"Navigate", "GetRuntimeId", "get_BoundingRectangle", "GetEmbeddedFragmentRoots", "SetFocus", "get_FragmentRoot"},
		"IRawElementProviderFragmentRoot": {"ElementProviderFromPoint", "GetFocus"},
		"IInvokeProvider":                 {"Invoke"},
		"IToggleProvider":                 {"Toggle", "get_ToggleState"},
		"IScrollItemProvider":             {"ScrollIntoView"},
		"IValueProvider":                  {"SetValue", "get_Value", "get_IsReadOnly"},
	}
	iids := map[string]windows.GUID{
		"IRawElementProviderSimple": iidSimple, "IRawElementProviderFragment": iidFragment,
		"IRawElementProviderFragmentRoot": iidFragmentRoot, "IInvokeProvider": iidInvokeProvider,
		"IToggleProvider": iidToggleProvider, "IScrollItemProvider": iidScrollItemProvider,
		"IValueProvider": iidValueProvider,
	}
	for name, want := range methods {
		re := regexp.MustCompile(`MIDL_INTERFACE\("([0-9a-fA-F-]+)"\)\s*` + name + `\s*:\s*public IUnknown\s*\{(?s:(.*?))\};`)
		m := re.FindStringSubmatch(core)
		if m == nil {
			t.Errorf("%s not found", name)
			continue
		}
		if got := strings.ToLower(m[1]); got != strings.ToLower(guidString(iids[name])) {
			t.Errorf("%s IID %s, SDK %s", name, guidString(iids[name]), got)
		}
		var got []string
		for _, mm := range regexp.MustCompile(`STDMETHODCALLTYPE\s+(\w+)\s*\(`).FindAllStringSubmatch(m[2], -1) {
			got = append(got, mm[1])
		}
		if strings.Join(got, ",") != strings.Join(want, ",") {
			t.Errorf("%s methods %v, SDK %v", name, want, got)
		}
	}
	client := string(sdkHeader(t, "UIAutomationClient.h"))
	defines := map[string]int{}
	// "#define UIA_X ( 1 )" in older SDKs, "const long UIA_X = 1;" in newer.
	for _, m := range regexp.MustCompile(`(?:#define\s+(UIA_\w+)\s+\(\s*(\d+)\s*\))|(?:const\s+\w+\s+(UIA_\w+)\s*=\s*(\d+)\s*;)`).FindAllStringSubmatch(client, -1) {
		name, value := m[1], m[2]
		if name == "" {
			name, value = m[3], m[4]
		}
		v, _ := strconv.Atoi(value)
		defines[name] = v
	}
	for name, value := range map[string]int{
		"UIA_InvokePatternId": patternInvoke, "UIA_ValuePatternId": patternValue,
		"UIA_TogglePatternId": patternToggle, "UIA_ScrollItemPatternId": patternScrollItem,
		"UIA_RuntimeIdPropertyId": propRuntimeID, "UIA_ControlTypePropertyId": propControlType,
		"UIA_LocalizedControlTypePropertyId": propLocalizedControlType, "UIA_NamePropertyId": propName,
		"UIA_HasKeyboardFocusPropertyId": propHasKeyboardFocus, "UIA_IsKeyboardFocusablePropertyId": propIsKeyboardFocusable,
		"UIA_IsEnabledPropertyId": propIsEnabled, "UIA_AutomationIdPropertyId": propAutomationID,
		"UIA_ClassNamePropertyId": propClassName, "UIA_IsControlElementPropertyId": propIsControlElement,
		"UIA_IsContentElementPropertyId": propIsContentElement, "UIA_IsOffscreenPropertyId": propIsOffscreen,
		"UIA_FrameworkIdPropertyId": propFrameworkID, "UIA_ItemStatusPropertyId": propItemStatus,
		"UIA_ValueValuePropertyId": propValueValue, "UIA_ValueIsReadOnlyPropertyId": propValueIsReadOnly,
		"UIA_ToggleToggleStatePropertyId": propToggleState,
		"UIA_ButtonControlTypeId":         ctButton, "UIA_CheckBoxControlTypeId": ctCheckBox,
		"UIA_ComboBoxControlTypeId": ctComboBox, "UIA_EditControlTypeId": ctEdit,
		"UIA_HyperlinkControlTypeId": ctHyperlink, "UIA_ImageControlTypeId": ctImage,
		"UIA_ListItemControlTypeId": ctListItem, "UIA_ListControlTypeId": ctList,
		"UIA_MenuControlTypeId": ctMenu, "UIA_MenuItemControlTypeId": ctMenuItem,
		"UIA_ProgressBarControlTypeId": ctProgressBar, "UIA_RadioButtonControlTypeId": ctRadioButton,
		"UIA_SliderControlTypeId": ctSlider, "UIA_StatusBarControlTypeId": ctStatusBar,
		"UIA_TabControlTypeId": ctTab, "UIA_TabItemControlTypeId": ctTabItem,
		"UIA_TextControlTypeId": ctText, "UIA_ToolBarControlTypeId": ctToolBar,
		"UIA_ToolTipControlTypeId": ctToolTip, "UIA_TreeControlTypeId": ctTree,
		"UIA_TreeItemControlTypeId": ctTreeItem, "UIA_GroupControlTypeId": ctGroup,
		"UIA_DataItemControlTypeId": ctDataItem, "UIA_PaneControlTypeId": ctPane,
		"UIA_HeaderItemControlTypeId": ctHeaderItem, "UIA_TableControlTypeId": ctTable,
		"UIA_SeparatorControlTypeId": ctSeparator,
	} {
		if got, ok := defines[name]; !ok || got != value {
			t.Errorf("%s = %d, SDK %d (%v)", name, value, got, ok)
		}
	}
}

func guidString(g windows.GUID) string {
	return fmt.Sprintf("%08x-%04x-%04x-%02x%02x-%x", g.Data1, g.Data2, g.Data3, g.Data4[0], g.Data4[1], g.Data4[2:])
}

// vcall calls method m (counting IUnknown's three) of interface pointer p.
func vcall(p uintptr, m int, args ...uintptr) uintptr {
	vtbl := *(*uintptr)(unsafe.Pointer(p))
	fn := *(*uintptr)(unsafe.Pointer(vtbl + uintptr(m)*ptrSize))
	r, _, _ := syscall.SyscallN(fn, append([]uintptr{p}, args...)...)
	return r
}

type recorder struct{ pressed, scrolled []string }

func (r *recorder) Press(key string) bool    { r.pressed = append(r.pressed, key); return true }
func (r *recorder) ScrollTo(key string) bool { r.scrolled = append(r.scrolled, key); return true }

var procSysFree = oleaut32.NewProc("SysFreeString")

func readVariant(t *testing.T, p uintptr, prop int) (uint16, uint64) {
	t.Helper()
	var v [2]uint64
	if hr := vcall(p, 5, uintptr(prop), uintptr(unsafe.Pointer(&v))); hr != sOK {
		t.Fatalf("GetPropertyValue(%d): %#x", prop, hr)
	}
	return uint16(v[0]), v[1]
}

func readString(t *testing.T, p uintptr, prop int) string {
	t.Helper()
	vt, v := readVariant(t, p, prop)
	if vt != vtBSTR {
		return ""
	}
	defer procSysFree.Call(uintptr(v))
	return windows.UTF16PtrToString((*uint16)(unsafe.Pointer(uintptr(v))))
}

func ptr(b bool) *bool { return &b }

// UI Automation's view of a window, through the providers' vtables:
// navigation, properties, patterns, actions and reference counts.
func TestProviders(t *testing.T) {
	acts := &recorder{}
	w := NewWindow(0, acts)
	view, area := image.Rect(0, 0, 400, 300), image.Rect(0, 40, 400, 300)
	w.Update([]Node{
		{Key: "1", Role: "checkbox", Name: "Paid", Checked: ptr(true), Bounds: image.Rect(10, 10, 200, 30), Viewport: view, Parent: -1},
		{Key: "2", Role: "scrollarea", Bounds: area, Viewport: view, Parent: -1},
		{Key: "3", Role: "button", Name: "Order 1", Bounds: image.Rect(0, 60, 400, 90), Viewport: area, Parent: 1},
		{Key: "4", Role: "button", Name: "Order 99", Bounds: image.Rect(0, 3000, 400, 3030), Viewport: area, Parent: 1},
	})
	mu.Lock()
	root := w.root.as(ifFragment)
	mu.Unlock()
	navigate := func(p uintptr, dir int) uintptr {
		t.Helper()
		var out uintptr
		if hr := vcall(p, 3, uintptr(dir), uintptr(unsafe.Pointer(&out))); hr != sOK {
			t.Fatalf("Navigate: %#x", hr)
		}
		return out
	}
	box := navigate(root, navFirstChild)
	pane := navigate(box, navNextSibling)
	if box == 0 || pane == 0 || navigate(pane, navNextSibling) != 0 {
		t.Fatal("top level is not checkbox, scroll pane")
	}
	first := navigate(pane, navFirstChild)
	last := navigate(pane, navLastChild)
	// Fragment pointers; GetPropertyValue lives on the simple interface.
	simple := func(fragment uintptr) uintptr { return fragment - (ifFragment-ifSimple)*ptrSize }
	if got := readString(t, simple(box), propName); got != "Paid" {
		t.Fatalf("checkbox name %q", got)
	}
	if vt, v := readVariant(t, simple(box), propControlType); vt != vtI4 || int32(v) != ctCheckBox {
		t.Fatalf("checkbox control type %d %d", vt, v)
	}
	if vt, v := readVariant(t, simple(last), propIsOffscreen); vt != vtBool || v == 0 {
		t.Fatalf("row out of view: offscreen %d %#x", vt, v)
	}
	if vt, v := readVariant(t, simple(first), propIsOffscreen); vt != vtBool || v != 0 {
		t.Fatalf("visible row: offscreen %d %#x", vt, v)
	}
	parent := navigate(first, navParent)
	if parent != pane {
		t.Fatalf("row's parent %#x, want the pane %#x", parent, pane)
	}
	pattern := func(fragment uintptr, id int) uintptr {
		t.Helper()
		var out uintptr
		if hr := vcall(simple(fragment), 4, uintptr(id), uintptr(unsafe.Pointer(&out))); hr != sOK {
			t.Fatalf("GetPatternProvider: %#x", hr)
		}
		return out
	}
	if pattern(box, patternInvoke) != 0 {
		t.Fatal("a checkbox has the Invoke pattern; it toggles")
	}
	toggle := pattern(box, patternToggle)
	var state int32
	if toggle == 0 || vcall(toggle, 4, uintptr(unsafe.Pointer(&state))) != sOK || state != toggleOn {
		t.Fatalf("toggle state %d", state)
	}
	vcall(toggle, 3) // Toggle
	invoke := pattern(first, patternInvoke)
	if invoke == 0 || vcall(invoke, 3) != sOK {
		t.Fatal("Invoke failed")
	}
	scroll := pattern(last, patternScrollItem)
	if scroll == 0 || vcall(scroll, 3) != sOK {
		t.Fatal("ScrollIntoView failed")
	}
	if strings.Join(acts.pressed, ",") != "1,3" || strings.Join(acts.scrolled, ",") != "4" {
		t.Fatalf("pressed %v, scrolled %v", acts.pressed, acts.scrolled)
	}
	// An element that leaves is unavailable, and freed with its last
	// reference.
	w.Update([]Node{{Key: "1", Role: "checkbox", Name: "Paid", Checked: ptr(false), Bounds: image.Rect(10, 10, 200, 30), Viewport: view, Parent: -1}})
	var out uintptr
	if hr := vcall(first, 3, navParent, uintptr(unsafe.Pointer(&out))); hr != uiaElementNotAvailable {
		t.Fatalf("navigating from a removed element: %#x", hr)
	}
	for _, p := range []uintptr{toggle, invoke, scroll, box, pane, parent, first, last, root} {
		vcall(p, 2) // Release
	}
	mu.Lock()
	n := len(objects)
	mu.Unlock()
	if n != 1 { // the root, held until Disconnect
		t.Fatalf("%d objects alive after every release", n)
	}
}
