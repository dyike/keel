//go:build linux

package platform

import "github.com/dyike/keel/internal/driver"

func New() driver.Backend { return unsupported{} }
