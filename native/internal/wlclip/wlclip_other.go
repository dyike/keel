//go:build !linux || android || nowayland

package wlclip

import (
	"time"
	"unsafe"

	"github.com/dyike/keel/native"
)

type Selection struct{}

func Open(display unsafe.Pointer) (*Selection, error) { return nil, native.ErrUnsupported }
func (s *Selection) MIMEs() []string                  { return nil }
func (s *Selection) Read(mime string, limit int, deadline time.Time) ([]byte, error) {
	return nil, native.ErrUnsupported
}
func (s *Selection) Close() {}
