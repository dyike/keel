package kit

import (
	"gioui.org/io/key"
	"github.com/dyike/keel/ui/el"
	"math"
	"testing"
)

func TestPaginationVisiblePageBudget(t *testing.T) {
	for _, total := range []int{1, 2, 3, 7, 20, 200, int(^uint(0) >> 1)} {
		for _, limit := range []int{1, 3, 4, 5, 10, 101} {
			p := Pagination(total, 1).VisiblePages(limit)
			for _, current := range []int{1, 2, total / 2, total - 1, total} {
				p.SetValue(current)
				nums := p.numbers()
				count, last := 0, 0
				found := false
				if nums[0] != 1 || nums[len(nums)-1] != total {
					t.Fatal("lost endpoints", nums)
				}
				for i, n := range nums {
					if n == 0 {
						if i == 0 || i == len(nums)-1 || nums[i-1] == 0 || nums[i+1]-nums[i-1] <= 1 {
							t.Fatal("invalid gap", nums)
						}
						continue
					}
					if n <= last || n > total {
						t.Fatal("invalid ordering", nums)
					}
					if i > 0 && nums[i-1] != 0 && n != last+1 {
						t.Fatal("missing gap", nums)
					}
					count++
					last = n
					found = found || n == p.Value()
				}
				if count > max(3, limit) || !found {
					t.Fatal("budget or current page", nums, current)
				}
			}
		}
	}
	p := Pagination(100, 1).VisiblePages(5)
	p.VisiblePages(-1)
	if p.visible != 5 {
		t.Fatal("negative accepted")
	}
	p.VisiblePages(0)
	if p.visible != 0 {
		t.Fatal("default not restored")
	}
	p.Size(40).Size(float32(math.NaN())).Size(float32(math.Inf(1)))
	if p.buttonSize != 40 {
		t.Fatal("invalid size accepted")
	}
}

func TestPaginationCompactSizeDisabledAndState(t *testing.T) {
	for _, scale := range []int{1, 2} {
		calls := 0
		p := Pagination(100, 10).VisiblePages(5).OnChange(func(int) { calls++ })
		disabled := false
		h := renderView(viewFunc(func(cx *el.Context) el.Element { return el.Div().Disabled(disabled).Child(p.Render(cx)) }), 224, scale)
		small := bounds(h, "下一页").Dy()
		p.Size(40).Compact(true)
		h.Frame()
		if shown(h, "1") || bounds(h, "下一页").Dy() <= small {
			t.Fatal("compact/size ignored")
		}
		click(t, h, "下一页")
		if p.Value() != 2 || calls != 1 {
			t.Fatal("compact next")
		}
		p.Compact(false)
		h.Frame()
		h.Key(key.NameSpace, 0)
		h.Frame()
		if p.Value() != 3 || calls != 2 {
			t.Fatal("mode switch lost next focus", p.Value(), calls)
		}
		p.SetDisabled(true)
		h.Frame()
		click(t, h, "下一页")
		if calls != 2 {
			t.Fatal("disabled action")
		}
		p.SetDisabled(false)
		disabled = true
		h.Frame()
		click(t, h, "下一页")
		if calls != 2 {
			t.Fatal("ancestor disabled action")
		}
		disabled = false
		p.SetTotal(1)
		h.Frame()
		if p.Value() != 1 || calls != 2 {
			t.Fatal("total clamp callback")
		}
		p.Size(0)
		h.Frame()
		if bounds(h, "下一页").Dy() != small {
			t.Fatal("default size not restored")
		}
		n, ok := semanticNode(h, "navigation:1/1")
		if !ok {
			t.Fatal("page semantics", n)
		}
	}
}
