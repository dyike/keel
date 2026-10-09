//go:build !(unix && !js && !wasip1) && !windows

package fontscan

import (
	"os"

	"github.com/dyike/keel/third_party/typesetting/font"
)

func openFont(path string) (font.Resource, func(), error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}
	return f, func() { f.Close() }, nil
}
