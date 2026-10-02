package kit

import "testing"

func TestListOwnsItemsAndProgrammaticSelection(t *testing.T) {
	source := []string{"a", "b"}
	list := List(source...)
	source[0] = "changed"
	items := list.Items()
	items[1] = "changed"
	if got := list.Items(); got[0] != "a" || got[1] != "b" {
		t.Fatal("list retained/exposed item aliases")
	}
	calls := 0
	list.OnChange(func(int) { calls++ })
	list.SetValue(1)
	replacement := []string{"c"}
	list.SetItems(replacement...)
	replacement[0] = "changed"
	if list.Items()[0] != "c" || list.Value() != -1 || calls != 0 {
		t.Fatal("replacement alias or callback")
	}
	list.SetItems()
	list.SetValue(0)
	if list.Value() != -1 || len(list.Items()) != 0 || calls != 0 {
		t.Fatal("empty list selection")
	}
}
