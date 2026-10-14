// SPDX-License-Identifier: Unlicense OR MIT

package app

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	"testing"

	"github.com/dyike/keel/gio/gpu"
	"github.com/dyike/keel/gio/op"
)

type recoveryContext struct {
	t                *testing.T
	err              error
	locked, released bool
	presents         int
}

func (c *recoveryContext) API() gpu.API                            { return nil }
func (c *recoveryContext) RenderTarget() (gpu.RenderTarget, error) { return nil, nil }
func (c *recoveryContext) Refresh() error                          { return nil }
func (c *recoveryContext) Lock() error {
	if c.locked {
		c.t.Fatal("recursive context lock")
	}
	c.locked = true
	return nil
}
func (c *recoveryContext) Unlock() {
	if !c.locked {
		c.t.Fatal("unlock without lock")
	}
	c.locked = false
}
func (c *recoveryContext) Present() error { c.presents++; return c.err }
func (c *recoveryContext) Release() {
	if c.locked {
		c.t.Fatal("context released while locked")
	}
	c.released = true
}

type recoveryGPU struct {
	ctx           *recoveryContext
	frames, idles int
	released      bool
}

func (g *recoveryGPU) Clear(color.NRGBA)                                  {}
func (g *recoveryGPU) Frame(*op.Ops, gpu.RenderTarget, image.Point) error { g.frames++; return nil }
func (g *recoveryGPU) Idle() {
	g.idles++
	if g.ctx.err != nil {
		g.ctx.t.Fatal("idle reclamation used a failed context")
	}
}
func (g *recoveryGPU) Release() {
	if !g.ctx.locked {
		g.ctx.t.Fatal("GPU released without context lock")
	}
	g.released = true
}

func TestPresentDeviceLostSchedulesRecovery(t *testing.T) {
	other := errors.New("presentation failed")
	for _, tc := range []struct {
		name string
		err  error
		lost bool
	}{
		{"success", nil, false}, {"lost", gpu.ErrDeviceLost, true}, {"wrapped", fmt.Errorf("present: %w", gpu.ErrDeviceLost), true}, {"other", other, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx := &recoveryContext{t: t, err: tc.err}
			g := &recoveryGPU{ctx: ctx}
			w := &Window{ctx: ctx, gpu: g}
			ack := make(chan struct{}, 2)
			var ops op.Ops
			got := w.validateAndProcess(image.Pt(80, 60), false, &ops, ack)
			if tc.lost {
				if got != nil || w.ctx != nil || w.gpu != nil || !ctx.released || !g.released || !w.hasNextFrame || !w.nextFrame.IsZero() {
					t.Fatalf("did not release and schedule recovery: err=%v window=%+v", got, w)
				}
			} else {
				if got != tc.err || ctx.released || g.released || w.hasNextFrame {
					t.Fatalf("unexpected recovery: err=%v", got)
				}
				defer w.destroyGPU()
			}
			if len(ack) != 1 || ctx.presents != 1 || g.frames != 1 || ctx.locked {
				t.Fatalf("frame retried, signalled twice, or remained locked: ack=%d presents=%d frames=%d locked=%v", len(ack), ctx.presents, g.frames, ctx.locked)
			}
			wantIdle := 0
			if tc.err == nil {
				wantIdle = 1
			}
			if g.idles != wantIdle {
				t.Fatalf("idle calls=%d want %d", g.idles, wantIdle)
			}
		})
	}
}
