//go:build (!darwin && !windows && !linux) || ios || android

package sys

import "github.com/dyike/keel/native"

func TrayOpen(TrayOptions) (TrayBackend, error) { return nil, native.ErrUnsupported }
func TrayMain()                                 { panic(native.ErrUnsupported) }
