//go:build !windows && (!darwin || ios || !cgo)

package sys

import "github.com/dyike/keel/native"

func ClipboardRead(done func([]byte, error)) { go done(nil, native.ErrUnsupported) }
