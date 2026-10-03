package kit

import "testing"

func TestToggleItemsInheritAndIsolate(t *testing.T) {
	options := []string{"A", "B"}
	group := ToggleGroup(options...).Size(ToggleSizeLarge)
	options[0] = "mutated"
	group.Item("unknown", Toggle("Ignored", false))
	if len(group.items) != 0 {
		t.Fatal("unknown option retained")
	}
	group.Item("A", Toggle("Display", false).Icon(IconStar))
	for _, scale := range []int{1, 2} {
		h := renderView(group, 300, scale)
		if bounds(h, "Display").Dy() != 40*scale {
			t.Fatal("size inheritance failed")
		}
		group.Item("A", Toggle("Display", false).Size(ToggleSizeMedium))
		h.Frame()
		if bounds(h, "Display").Dy() != 32*scale || bounds(h, "B").Dy() != 40*scale {
			t.Fatal("explicit medium not isolated")
		}
		group.Item("A", Toggle("Display", false).Icon(IconStar))
	}
}
