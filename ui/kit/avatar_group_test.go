package kit

import (
	"github.com/dyike/keel/third_party/gio/f32"
	"github.com/dyike/keel/third_party/gio/io/pointer"
	"math"
	"testing"
)

func TestAvatarGroupLimitOverlapAndOwnership(t *testing.T) {
	for _, scale := range []int{1, 2} {
		a, b, c := Avatar("Ada"), Avatar("Bob"), Avatar("Carol")
		avatars := []*AvatarView{a, nil, b, c}
		group := AvatarGroup(avatars...).Size(32).Limit(2)
		avatars[0] = nil
		group.Avatars()[0] = nil
		h := renderView(group, 200, scale)
		first, second := bounds(h, "Ada"), bounds(h, "Bob")
		if first.Dx() != 32*scale || second.Min.X-first.Min.X != 24*scale || second.Min.Y != first.Min.Y {
			t.Fatalf("overlap: %v %v", first, second)
		}
		if !shown(h, "+1") || shown(h, "Carol") || a.size != 40 {
			t.Fatal("overflow or shared avatar size changed")
		}
		group.Ellipsis(true)
		h.Frame()
		if !shown(h, "…") || !shown(h, "更多 1") {
			t.Fatal("ellipsis or count semantics missing")
		}
		group.Limit(0).Ellipsis(false)
		h.Frame()
		if shown(h, "Ada") || !shown(h, "+3") {
			t.Fatal("zero limit")
		}
		group.Limit(-1)
		h.Frame()
		if !shown(h, "Carol") || shown(h, "+1") {
			t.Fatal("unlimited")
		}
		group.SetAvatars()
		h.Frame()
		if shown(h, "Ada") {
			t.Fatal("empty group")
		}
	}
}

func TestAvatarGroupNarrowScrollAndInvalidSize(t *testing.T) {
	group := AvatarGroup(Avatar("A"), Avatar("B"), Avatar("C")).Size(32)
	for _, invalid := range []float32{0, -1, float32(math.NaN()), float32(math.Inf(1))} {
		group.Size(invalid)
	}
	h := renderView(group, 48, 1)
	h.Router.Queue(pointer.Event{Kind: pointer.Scroll, Source: pointer.Mouse, Position: f32.Pt(20, 15), Scroll: f32.Pt(200, 0)})
	h.Frame()
	h.Frame()
	b := bounds(h, "C")
	if b.Empty() || b.Min.X < 0 || b.Max.X > 48 {
		t.Fatalf("last avatar unreachable: %v", b)
	}
}
