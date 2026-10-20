//go:build race

package kit

// raceDetector reports whether tests run under -race, which slows them
// several times; wall-clock budgets are checked only without it.
const raceDetector = true
