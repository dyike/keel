// SPDX-License-Identifier: Unlicense OR MIT

package app

import (
	"sync"
	"testing"
)

func TestEventFuncsOnlyRunWhenDrained(t *testing.T) {
	var q eventFuncs
	calls, wakes := 0, 0
	var wg sync.WaitGroup
	for range 20 {
		wg.Add(1)
		go func() { defer wg.Done(); q.add(func() { calls++ }, func() { wakes++ }) }()
	}
	wg.Wait()
	if calls != 0 || wakes != 20 {
		t.Fatalf("calls=%d wakes=%d", calls, wakes)
	}
	q.drain()
	if calls != 20 {
		t.Fatalf("calls=%d", calls)
	}
	q.drain()
	if calls != 20 {
		t.Fatal("work repeated")
	}
	q.close()
	q.add(func() { t.Fatal("work ran after shutdown") }, func() { t.Fatal("woke destroyed native pipe") })
	q.drain()
}
