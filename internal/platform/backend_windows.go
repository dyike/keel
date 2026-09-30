//go:build windows

package platform

import "github.com/dyike/keel/internal/driver"

func New() driver.Backend { return unsupported{} }
