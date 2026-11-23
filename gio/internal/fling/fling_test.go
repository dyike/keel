package fling

import (
	"math"
	"testing"
	"time"

	"github.com/dyike/keel/gio/unit"
)

func TestExtrapolationTwoSampleQuickFlick(t *testing.T) {
	var e Extrapolation
	e.Sample(0, 500)
	e.Sample(16*time.Millisecond, 420)
	est := e.Estimate()
	if est.Velocity == 0 {
		t.Fatal("expected non-zero velocity for 2-sample quick flick")
	}
	// Moving finger up from 500 to 420 in 16ms -> positive scroll velocity (~5000 px/s).
	if est.Velocity < 4500 || est.Velocity > 5500 {
		t.Fatalf("velocity = %v, want ~5000", est.Velocity)
	}
	if math.Abs(float64(est.Distance)) < 75 {
		t.Fatalf("distance = %v, want magnitude ~80", est.Distance)
	}
}

func TestAnimationFrictionAndSmoothRounding(t *testing.T) {
	if got := frictionForOS("ios"); got != -2.0 {
		t.Fatalf("ios friction = %v, want -2.0", got)
	}
	if got := frictionForOS("android"); got > -1.5 || got < -2.6 {
		t.Fatalf("android friction = %v, expected smooth range [-2.6, -1.5]", got)
	}
	var a Animation
	m := unit.Metric{PxPerDp: 3, PxPerSp: 3} // typical high-DPI Android phone
	t0 := time.Unix(1000, 0)
	if !a.Start(m, t0, 3000) {
		t.Fatal("Start returned false for 3000 px/s")
	}
	total := 0
	frames := 0
	for now := t0.Add(8 * time.Millisecond); a.Active() && frames < 400; now = now.Add(8 * time.Millisecond) {
		d := a.Tick(now)
		if d < 0 {
			t.Fatalf("negative step %d during positive fling", d)
		}
		total += d
		frames++
	}
	if a.Active() {
		t.Fatal("fling did not settle within 400 frames")
	}
	// Expected total distance ≈ v0 / |k| = 3000 / 2.2 ≈ 1363px (for non-darwin).
	if total < 1000 {
		t.Fatalf("total fling distance = %d, too stiff", total)
	}
}
