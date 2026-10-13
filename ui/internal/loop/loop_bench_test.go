package loop

import "testing"

func BenchmarkPostAndDrain(b *testing.B) {
	SetIdleMemoryReclaim(false)
	defer SetIdleMemoryReclaim(true)
	fn := func() {}
	b.ReportAllocs()
	for b.Loop() {
		Post(fn)
		Lock()
		Drain()
		Unlock()
	}
}
