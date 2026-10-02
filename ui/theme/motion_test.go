package theme

import "testing"

func TestSystemMotionAndApplicationOverride(t *testing.T) {
	old, system, override := ReducedMotion, systemReducedMotion, motionOverride
	defer func() { ReducedMotion, systemReducedMotion, motionOverride = old, system, override }()
	FollowSystemMotion()
	SetSystemReducedMotion(true)
	if !ReducedMotion {
		t.Fatal("system not applied")
	}
	SetReducedMotion(false)
	SetSystemReducedMotion(false)
	SetSystemReducedMotion(true)
	if ReducedMotion {
		t.Fatal("system replaced explicit preference")
	}
	FollowSystemMotion()
	if !ReducedMotion {
		t.Fatal("latest preference not remembered")
	}
	SetSystemReducedMotion(false)
	if ReducedMotion {
		t.Fatal("follow mode not restored")
	}
}
