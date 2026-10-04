//go:build linux && !android

package sys

import (
	"errors"
	"fmt"
	"github.com/dyike/keel/native"
)

func ClipboardRead(done func([]byte, error)) {
	go func() {
		raw, err := readX11Clipboard()
		if err != nil && !errors.Is(err, native.ErrUnsupported) && !errors.Is(err, native.ErrTimeout) && !errors.Is(err, native.ErrFailed) {
			err = fmt.Errorf("%w: X11 clipboard: %v", native.ErrFailed, err)
		}
		done(raw, err)
	}()
}
