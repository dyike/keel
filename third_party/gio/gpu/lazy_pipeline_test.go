// SPDX-License-Identifier: Unlicense OR MIT

package gpu

import (
	"testing"

	"gioui.org/gpu/internal/driver"
)

type lazyTestPipeline struct{ releases int }

func (p *lazyTestPipeline) Release() { p.releases++ }

func TestLazyPipelineLifecycle(t *testing.T) {
	for _, used := range []bool{false, true} {
		loads := 0
		native := new(lazyTestPipeline)
		p := &pipeline{load: func() (driver.Pipeline, *uniformBuffer) {
			loads++
			return native, nil
		}}
		if used {
			p.ensure()
			p.ensure()
			if loads != 1 {
				t.Fatalf("loaded %d times", loads)
			}
		}
		p.Release()
		p.Release()
		want := 0
		if used {
			want = 1
		}
		if loads != want || native.releases != want || p.load != nil {
			t.Fatalf("used=%v loads=%d releases=%d", used, loads, native.releases)
		}
	}
}
