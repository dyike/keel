package window

import "github.com/dyike/keel/ui/internal/loop"

// SetIdleMemoryReclaim sets process-wide reclamation of unused Go heap pages
// after all Keel windows have been quiet for two seconds. It defaults to true:
// startup alone leaves tens of megabytes of freed heap the runtime would keep.
// Reclamation requires at least 32 MiB of unused, unreturned heap pages and is
// limited to once per 30 seconds. It does not redraw windows or change GOGC.
//
// This runs a Go collection and returns free pages to the OS. It can briefly
// pause other goroutines, including non-UI workloads. Applications with latency
// sensitive background work can disable it. Disabling cancels pending
// work, but cannot undo a reclamation already in progress.
func SetIdleMemoryReclaim(enabled bool) { loop.SetIdleMemoryReclaim(enabled) }
