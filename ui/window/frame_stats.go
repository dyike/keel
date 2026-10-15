package window

import (
	"fmt"
	"log"
	"os"
	"runtime/metrics"
	"strconv"
	"strings"
	"time"
)

var frameStatsThreshold, frameStatsEnabled = frameStatsSetting(os.Getenv("KEEL_FRAME_STATS"))

var frameStatsLog = log.Print

func frameStatsSetting(value string) (time.Duration, bool) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", "0", "false", "off", "no":
		return 0, false
	case "1", "true", "on", "yes":
		return 8 * time.Millisecond, true
	case "all":
		return 0, true
	}
	ms, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
	if err != nil || ms < 0 {
		log.Printf("window: KEEL_FRAME_STATS=%q is not a non-negative millisecond threshold; using 8ms", value)
		return 8 * time.Millisecond, true
	}
	return time.Duration(ms * float64(time.Millisecond)), true
}

const (
	frameStatObjects = iota
	frameStatTiny
	frameStatBytes
	frameStatGC
	frameStatCount
)

type frameStats struct {
	enabled        bool
	start          time.Time
	layoutFinished time.Time
	before         [frameStatCount]float64
}

func beginFrameStats() frameStats {
	if !frameStatsEnabled {
		return frameStats{}
	}
	return frameStats{enabled: true, start: time.Now(), before: readFrameStats()}
}

func (f *frameStats) layoutDone() {
	if f.enabled {
		f.layoutFinished = time.Now()
	}
}

func (f *frameStats) end() {
	if !f.enabled {
		return
	}
	finished := time.Now()
	total := finished.Sub(f.start)
	if total < frameStatsThreshold {
		return
	}
	after := readFrameStats()
	allocs := after[frameStatObjects] + after[frameStatTiny] - f.before[frameStatObjects] - f.before[frameStatTiny]
	bytes := after[frameStatBytes] - f.before[frameStatBytes]
	gc := after[frameStatGC] - f.before[frameStatGC]
	layout := f.layoutFinished.Sub(f.start)
	submit := finished.Sub(f.layoutFinished)
	frameStatsLog(fmt.Sprintf("window: frame took %.2fms: layout %.2fms, submit %.2fms; process allocated about %.0f objects, %.1f KiB; %.0f GC cycles ended",
		total.Seconds()*1000, layout.Seconds()*1000, submit.Seconds()*1000, allocs, bytes/1024, gc))
}

func readFrameStats() [frameStatCount]float64 {
	samples := [frameStatCount]metrics.Sample{
		{Name: "/gc/heap/allocs:objects"},
		{Name: "/gc/heap/tiny/allocs:objects"},
		{Name: "/gc/heap/allocs:bytes"},
		{Name: "/gc/cycles/total:gc-cycles"},
	}
	metrics.Read(samples[:])
	var values [frameStatCount]float64
	for i, sample := range samples {
		switch sample.Value.Kind() {
		case metrics.KindUint64:
			values[i] = float64(sample.Value.Uint64())
		case metrics.KindFloat64:
			values[i] = sample.Value.Float64()
		}
	}
	return values
}
