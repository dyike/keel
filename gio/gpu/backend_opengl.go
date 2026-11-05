// SPDX-License-Identifier: Unlicense OR MIT

//go:build !noopengl

package gpu

// Keel patch: the existing noopengl option must exclude registration too;
// otherwise its init function retains the backend and native GL bindings.
import _ "github.com/dyike/keel/gio/gpu/internal/opengl"
