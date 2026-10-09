package window

import "github.com/dyike/keel/ui/internal/loop"

// SetIdleMemoryReclaim sets process-wide reclamation of unused Go heap pages
// after all Keel windows have been quiet for two seconds. It defaults to true:
// startup alone leaves tens of megabytes of freed heap the runtime would keep.
// Reclamation requires 8 MiB of unused, unreturned pages, or at least 8 MiB
// of heap objects and 8 MiB allocated since the last idle collection. It is
// limited to once per 5 seconds and does not redraw windows or change GOGC.
//
// This runs a Go collection and returns free pages to the OS. It can briefly
// pause other goroutines, including non-UI workloads. Applications with latency
// sensitive background work can disable it. Disabling cancels pending
// work, but cannot undo a reclamation already in progress.
func SetIdleMemoryReclaim(enabled bool) { loop.SetIdleMemoryReclaim(enabled) }
