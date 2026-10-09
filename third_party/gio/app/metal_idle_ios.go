// SPDX-License-Identifier: Unlicense OR MIT

//go:build darwin && ios && !nometal

package app

type metalIdleState struct{}

func (c *mtlContext) stopIdle()   {}
func (c *mtlContext) resumeIdle() {}
