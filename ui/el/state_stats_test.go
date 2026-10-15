package el

import "testing"

func TestStateCreationCounterDistinguishesReuseAndNewKeys(t *testing.T) {
	s := newStore()
	s.frame = 1
	a := s.get(7)
	if s.created != 1 || s.get(7) != a || s.created != 1 {
		t.Fatal("stable key allocated new state")
	}
	s.frame++
	s.get(7)
	s.get(8)
	if s.created != 2 {
		t.Fatal("new key was not counted")
	}
	s.sweep()
	s.frame++
	s.get(8)
	s.sweep()
	if len(s.states) != 1 || s.states[8] == nil {
		t.Fatal("counter changed state sweeping")
	}
}

func TestStateStatsLabelsAndReferences(t *testing.T) {
	s := newStore()
	s.stats = &stateStats{id: 1, labels: make(map[stateKey]stateLabel)}
	s.beginStateStats()
	n := Div().ID("save").Role("button").Name("Save").node()
	s.assignKeys(n, 1)
	s.get(1)
	if s.stats.new != 1 || len(s.stats.first) != 1 || s.stats.first[0].label.name != "Save" {
		t.Fatal("missing state label")
	}
	old := stateStatsLog
	defer func() { stateStatsLog = old }()
	calls := 0
	stateStatsLog = func(string, ...any) { calls++ }
	s.endStateStats()
	if calls != 2 || len(s.stats.labels) != 0 || len(s.stats.first) != 0 {
		t.Fatal("diagnostics retained frame references")
	}
	for _, entry := range s.stats.first[:cap(s.stats.first)] {
		if entry != (stateCreation{}) {
			t.Fatal("diagnostic buffer retained strings")
		}
	}
}
