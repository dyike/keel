package kit

import (
	"strconv"

	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/locale"
	"github.com/dyike/keel/ui/theme"
)

// PaginationView pages through total items, pageSize at a time. Pages are
// numbered from 1; long ranges show the first, the last and the pages around
// the current one, with … between.
type PaginationView struct {
	total, size, page int
	onChange          func(page int)
	compact, disabled bool
	visible           int
	buttonSize        float32
}

func Pagination(total, pageSize int) *PaginationView {
	return &PaginationView{total: max(total, 0), size: max(pageSize, 1), page: 1}
}
func (v *PaginationView) OnChange(fn func(page int)) *PaginationView { v.onChange = fn; return v }

// Compact shows only previous/next controls, retaining page semantics.
func (v *PaginationView) Compact(on bool) *PaginationView { v.compact = on; return v }

// VisiblePages limits numbered buttons, excluding gaps and previous/next.
// Positive limits are clamped to 3–101 to retain first/current/last pages.
// Zero restores the original page window; negative values are ignored.
func (v *PaginationView) VisiblePages(n int) *PaginationView {
	if n >= 0 {
		v.visible = n
		if n > 0 {
			v.visible = min(101, max(3, n))
		}
	}
	return v
}

// Size sets button height in dp (16–128); zero restores the default 28dp.
func (v *PaginationView) Size(dp float32) *PaginationView {
	if dp >= 0 && finiteNumber(float64(dp)) {
		v.buttonSize = dp
		if dp > 0 {
			v.buttonSize = min(128, max(16, dp))
		}
	}
	return v
}

func (v *PaginationView) SetDisabled(on bool) { v.disabled = on }

// Value is the current page, from 1.
func (v *PaginationView) Value() int { return v.page }

// SetValue goes to page p (clamped) without calling OnChange.
func (v *PaginationView) SetValue(p int) { v.page = min(max(p, 1), v.Pages()) }

// SetTotal changes the item count, keeping the page in range.
func (v *PaginationView) SetTotal(n int) { v.total = max(n, 0); v.SetValue(v.page) }

// Pages is the number of pages, at least 1.
func (v *PaginationView) Pages() int {
	if v.total == 0 {
		return 1
	}
	return (v.total-1)/v.size + 1
}

// Bounds returns the item range [start, end) of the current page.
func (v *PaginationView) Bounds() (start, end int) {
	start = (v.page - 1) * v.size
	return start, start + min(v.size, v.total-start)
}

func (v *PaginationView) goTo(p int) {
	p = min(max(p, 1), v.Pages())
	if v.disabled || p == v.page {
		return
	}
	v.page = p
	if v.onChange != nil {
		v.onChange(p)
	}
}

// numbers lists the pages to show; 0 stands for a gap.
func (v *PaginationView) numbers() []int {
	n := v.Pages()
	if v.visible > 0 && n > v.visible {
		count := v.visible - 2
		lo := max(2, min(v.page-count/2, n-count))
		hi := lo + count - 1
		out := []int{1}
		if lo > 2 {
			out = append(out, 0)
		}
		for p := lo; p <= hi; p++ {
			out = append(out, p)
		}
		if hi < n-1 {
			out = append(out, 0)
		}
		return append(out, n)
	}
	if v.visible > 0 || n <= 7 {
		out := make([]int, n)
		for i := range out {
			out[i] = i + 1
		}
		return out
	}
	lo, hi := max(2, v.page-1), min(n-1, v.page+min(1, n-v.page))
	if v.page <= 3 {
		lo, hi = 2, 4
	}
	if v.page >= n-2 {
		lo, hi = n-3, n-1
	}
	out := []int{1}
	if lo > 2 {
		out = append(out, 0)
	}
	for p := lo; p <= hi; p++ {
		out = append(out, p)
	}
	if hi < n-1 {
		out = append(out, 0)
	}
	return append(out, n)
}

func (v *PaginationView) Render(cx *el.Context) el.Element {
	text := locale.Current()
	size := v.buttonSize
	if size == 0 {
		size = 28
	}
	prev := Button("", func() { v.goTo(v.page - 1) }).Name(text.PrevPage).Icon(IconChevronLeft).Variant(ButtonGhost).Size(size)
	next := Button("", func() { v.goTo(v.page + min(1, v.Pages()-v.page)) }).Name(text.NextPage).Icon(IconChevronRight).Variant(ButtonGhost).Size(size)
	prev.ID("prev")
	next.ID("next")
	prev.SetDisabled(v.page <= 1)
	next.SetDisabled(v.page >= v.Pages())
	row := el.Div().Disabled(v.disabled).Role("navigation").Value(strconv.Itoa(v.page) + "/" + strconv.Itoa(v.Pages())).
		Wrap().Items(el.Center).Gap(theme.SpaceXs)
	if v.compact {
		return row.Child(prev.Render(cx), next.Render(cx))
	}
	row.Child(el.Text(text.Total(v.total)).ID("total").TextColor(theme.Muted).TextSize(theme.TextMd), el.Div().ID("spacer").W(el.Dp(4)), prev.Render(cx))
	for i, p := range v.numbers() {
		if p == 0 {
			row.Child(el.Text("…").ID("gap-" + strconv.Itoa(i)).TextColor(theme.Muted).Px(theme.SpaceXs))
			continue
		}
		p := p
		variant := ButtonGhost
		if p == v.page {
			variant = ButtonPrimary
		}
		row.Child(Button(strconv.Itoa(p), func() { v.goTo(p) }).ID("page-" + strconv.Itoa(p)).Variant(variant).Size(size).Render(cx))
	}
	return row.Child(next.Render(cx))
}
