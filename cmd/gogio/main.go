// Command gogio runs Keel's copy of Gio's packager directly, for builds
// that have no keel.json, such as the site's component gallery:
//
//	go run ./cmd/gogio -target js -tags osusergo -o demo ./examples/components
package main

import (
	"fmt"
	"os"

	"github.com/dyike/keel/internal/gogio"
)

func main() {
	if err := gogio.Run(os.Args[1:], gogio.Options{}); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
