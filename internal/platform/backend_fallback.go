//go:build (!darwin && !windows && !linux) || (darwin && !cgo)

package platform

import "github.com/dyike/keel/internal/driver"

func New() driver.Backend { return unsupported{} }
