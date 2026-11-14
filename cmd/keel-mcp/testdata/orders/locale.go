package main

import (
	"log"
	"strings"

	"github.com/dyike/keel/ui/locale"
)

func applyLanguage(language string) {
	switch language {
	case "en":
		locale.Apply(locale.English())
	case "zh-CN":
		locale.Apply(locale.Chinese())
	default:
		log.Fatalf("unknown language %q", language)
	}
}

func text(english, chinese string) string {
	if strings.HasPrefix(locale.Current().Lang, "zh") {
		return chinese
	}
	return english
}
