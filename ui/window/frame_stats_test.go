package window

import (
	"strings"
	"testing"
	"time"
)

func TestFrameStatsSetting(t *testing.T) {
	for _, test := range []struct {
		value     string
		threshold time.Duration
		enabled   bool
	}{
		{"", 0, false},
		{"off", 0, false},
		{"1", 8 * time.Millisecond, true},
		{"true", 8 * time.Millisecond, true},
		{"all", 0, true},
		{"4.5", 4500 * time.Microsecond, true},
		{"invalid", 8 * time.Millisecond, true},
	} {
		threshold, enabled := frameStatsSetting(test.value)
		if threshold != test.threshold || enabled != test.enabled {
			t.Errorf("KEEL_FRAME_STATS=%q: got %v, %v; want %v, %v", test.value, threshold, enabled, test.threshold, test.enabled)
		}
	}
}

func TestFrameStatsLog(t *testing.T) {
	var line string
	old, oldThreshold := frameStatsLog, frameStatsThreshold
	frameStatsLog = func(values ...any) { line = values[0].(string) }
	frameStatsThreshold = 0
	defer func() { frameStatsLog, frameStatsThreshold = old, oldThreshold }()
	now := time.Now()
	stats := frameStats{
		enabled:        true,
		start:          now.Add(-2 * time.Millisecond),
		layoutFinished: now.Add(-time.Millisecond),
		before:         readFrameStats(),
	}
	stats.end()
	for _, want := range []string{"layout", "submit", "objects", "KiB", "GC cycles"} {
		if !strings.Contains(line, want) {
			t.Errorf("frame log %q lacks %q", line, want)
		}
	}
}
