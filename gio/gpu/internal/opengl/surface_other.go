// SPDX-License-Identifier: Unlicense OR MIT

//go:build !linux || android

package opengl

import "github.com/dyike/keel/gio/gpu/internal/driver"

type surfaceState struct{}

func (*Backend) initSurfaces()                                            {}
func (*Backend) collectSurfaces(bool)                                     {}
func (*Backend) releaseSurfaces()                                         {}
func (*Backend) retireSurface(*texture)                                   {}
func (*Backend) externalTexture() bool                                    { return false }
func (*Backend) bindSurfaceTexture(int, *texture) bool                    { return false }
func (*Backend) newExternalProgram(driver.PipelineDesc) (*program, error) { return nil, nil }
