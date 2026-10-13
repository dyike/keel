// Command gogio runs Keel's copy of Gio's packager directly, for builds
// that have no keel.json, such as the site's component gallery:
//
//	go run ./cmd/keel/internal/gogio/cmd -target js -tags osusergo -o demo ./examples/components
package main

import (
	"fmt"
	"os"

	"github.com/dyike/keel/cmd/keel/internal/gogio"
)

func main() {
	if err := gogio.Run(os.Args[1:], gogio.Options{}); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
