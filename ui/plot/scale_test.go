package plot

import (
	"math"
	"testing"
)

func near(a, b float64) bool { return math.Abs(a-b) < 1e-9*max(1, math.Abs(a), math.Abs(b)) }
func TestLinearDomainsAndExtremes(t *testing.T) {
	for _, s := range []ScaleLinear{{Domain: [2]float64{0, 100}, Range: [2]float64{0, 500}}, {Domain: [2]float64{100, 0}, Range: [2]float64{0, 500}}, {Domain: [2]float64{-math.MaxFloat64, math.MaxFloat64}, Range: [2]float64{500, 0}}} {
		ticks := s.Ticks(5)
		for i, x := range ticks {
			y, ok := s.Map(x)
			if !ok || !near(y, interpolate(s.Range[0], s.Range[1], float64(i)/4)) {
				t.Fatal("map", x, y)
			}
			back, ok := s.Invert(y)
			if !ok || !near(x, back) {
				t.Fatal("inverse", x, back)
			}
		}
	}
	s := ScaleLinear{Domain: [2]float64{2, 2}, Range: [2]float64{0, 10}}
	if x, ok := s.Map(99); !ok || x != 5 {
		t.Fatal("constant")
	}
	s = ScaleLinear{Domain: [2]float64{0, 1}, Range: [2]float64{0, 10}, Clamp: true}
	if x, _ := s.Map(-9); x != 0 {
		t.Fatal("clamp")
	}
	if _, ok := s.Map(math.NaN()); ok {
		t.Fatal("NaN")
	}
}
func TestCategoricalScales(t *testing.T) {
	input := []string{"a", "b", "a", "c"}
	s := NewBand(input, [2]float64{0, 300}).Padding(.1, .1)
	input[0] = "changed"
	if len(s.Domain()) != 3 {
		t.Fatal("unique owned domain")
	}
	d := s.Domain()
	d[0] = "changed"
	if _, ok := s.Map("a"); !ok {
		t.Fatal("domain escaped")
	}
	previous := -1.0
	for _, key := range s.Domain() {
		p, ok := s.Map(key)
		if !ok || p <= previous || p+s.Bandwidth() > 300 {
			t.Fatal("band", p)
		}
		previous = p + s.Bandwidth()
	}
	reverse := NewBand([]int{1, 2, 3}, [2]float64{300, 0})
	a, _ := reverse.Map(1)
	b, _ := reverse.Map(3)
	if a != 200 || b != 0 {
		t.Fatal("reversed", a, b)
	}
	singleton := NewPoint([]string{"only"}, [2]float64{0, 300})
	if p, _ := singleton.Map("only"); p != 150 {
		t.Fatal("singleton", p)
	}
	pts := NewPoint([]int{1, 2, 3}, [2]float64{0, 300})
	for i := 1; i <= 3; i++ {
		if p, _ := pts.Map(i); p != float64(i-1)*150 {
			t.Fatal("point", p)
		}
	}
	if _, ok := pts.Map(9); ok {
		t.Fatal("missing")
	}
	values := []string{"red", "blue"}
	ord := NewOrdinal([]int{1, 2, 3}, values)
	values[0] = "changed"
	if c, ok := ord.Map(3); !ok || c != "red" {
		t.Fatal("ordinal", c)
	}
	if _, ok := NewOrdinal([]int{1}, []string{}).Map(1); ok {
		t.Fatal("empty ordinal")
	}
}
func TestStackAndPieConservation(t *testing.T) {
	stacks, err := Stack([][]float64{{3, -2, math.NaN()}, {4, -5}, {-1, 8, 6}})
	if err != nil {
		t.Fatal(err)
	}
	if stacks[1][0].Low != 3 || stacks[1][0].High != 7 || stacks[1][1].High != -7 || stacks[2][0].Low != 0 || stacks[2][1].High != 8 || stacks[1][2].Valid {
		t.Fatal(stacks)
	}
	if out, err := Stack([][]float64{{math.MaxFloat64}, {math.MaxFloat64}}); err == nil || out != nil {
		t.Fatal("overflow atomicity")
	}
	for _, sign := range []float64{-1, 1} {
		arcs, err := Pie([]float64{math.MaxFloat64, 0, math.MaxFloat64}, 0, sign*2*math.Pi, .1)
		if err != nil || len(arcs) != 2 {
			t.Fatal(arcs, err)
		}
		total := 0.0
		for _, a := range arcs {
			total += math.Abs(a.End - a.Start)
		}
		if !near(total+.2, 2*math.Pi) || arcs[1].Index != 2 {
			t.Fatal("pie conservation", arcs)
		}
	}
	if _, err := Pie([]float64{-1}, 0, 1, 0); err == nil {
		t.Fatal("negative pie")
	}
}
