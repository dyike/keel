package main

import (
	"strings"

	"github.com/dyike/keel/ui/locale"
)

// demoText selects application-owned example text. Framework labels are
// translated separately by ui/locale. English comes first, like the docs.
func demoText(english, chinese string) string {
	if strings.HasPrefix(locale.Current().Lang, "zh") {
		return chinese
	}
	return english
}
