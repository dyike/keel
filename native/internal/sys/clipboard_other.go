//go:build !windows && (!linux || android) && (!darwin || ios)

package sys

import "github.com/dyike/keel/native"

func ClipboardRead(done func([]byte, error)) { go done(nil, native.ErrUnsupported) }
