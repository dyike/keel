//go:build !js

package atspi

import (
	"os"
	"regexp"
	"strings"
	"testing"
)

// The role and state numbers match at-spi2-core's header, where it is
// installed (libatspi2.0-dev).
func TestConstantsMatchHeader(t *testing.T) {
	header := os.Getenv("ATSPI_CONSTANTS_H")
	if header == "" {
		header = "/usr/include/at-spi-2.0/atspi/atspi-constants.h"
	}
	src, err := os.ReadFile(header)
	if err != nil {
		t.Skip("atspi-constants.h not found:", err)
	}
	values := func(enum, prefix string) map[string]uint32 {
		// typedef enum { ATSPI_ROLE_INVALID, ... } AtspiRole;
		re := regexp.MustCompile(`(?s)typedef enum\s*\{([^}]*)\}\s*` + enum + `;`)
		m := re.FindSubmatch(src)
		if m == nil {
			t.Fatalf("%s not found in %s", enum, header)
		}
		body := regexp.MustCompile(`(?s)/\*.*?\*/`).ReplaceAllString(string(m[1]), "")
		out := map[string]uint32{}
		next := uint32(0)
		for _, item := range strings.Split(body, ",") {
			name, alias, isAlias := strings.Cut(strings.TrimSpace(item), "=")
			name = strings.TrimSpace(name)
			if !strings.HasPrefix(name, prefix) {
				continue
			}
			if isAlias { // ATSPI_ROLE_PUSH_BUTTON = ATSPI_ROLE_BUTTON
				out[strings.TrimPrefix(name, prefix)] = out[strings.TrimPrefix(strings.TrimSpace(alias), prefix)]
				continue
			}
			out[strings.TrimPrefix(name, prefix)] = next
			next++
		}
		return out
	}
	roles := values("AtspiRole", "ATSPI_ROLE_")
	for value, name := range roleNames {
		key := strings.ToUpper(strings.ReplaceAll(name, " ", "_"))
		if got, ok := roles[key]; !ok || got != value {
			t.Errorf("role %s = %d, header has %d (%v)", key, value, got, ok)
		}
	}
	states := values("AtspiStateType", "ATSPI_STATE_")
	for name, value := range map[string]uint32{
		"ACTIVE": stateActive, "CHECKABLE": stateCheckable, "CHECKED": stateChecked,
		"EDITABLE": stateEditable, "ENABLED": stateEnabled, "EXPANDABLE": stateExpandable,
		"EXPANDED": stateExpanded, "FOCUSABLE": stateFocusable, "INDETERMINATE": stateIndeterminate,
		"RESIZABLE": stateResizable, "SELECTABLE": stateSelectable, "SELECTED": stateSelected,
		"SENSITIVE": stateSensitive, "SHOWING": stateShowing, "SINGLE_LINE": stateSingleLine,
		"VISIBLE": stateVisible,
	} {
		if got, ok := states[name]; !ok || got != value {
			t.Errorf("state %s = %d, header has %d (%v)", name, value, got, ok)
		}
	}
}
