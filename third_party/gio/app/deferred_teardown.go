// SPDX-License-Identifier: Unlicense OR MIT

package app

// deferredTeardown keeps native handles alive while a caller processes an
// invalid view event. The next Event call acknowledges that notification.
// It is owned by the event-loop goroutine, not called concurrently.
type deferredTeardown struct {
	finish    func()
	delivered bool
}

func (d *deferredTeardown) begin(notify, finish func()) {
	if d.finish != nil {
		return
	}
	d.finish = finish
	notify()
}

func (d *deferredTeardown) viewReturned() {
	d.delivered = d.finish != nil
}

func (d *deferredTeardown) complete() {
	if !d.delivered {
		return
	}
	if finish := d.finish; finish != nil {
		d.finish = nil
		d.delivered = false
		finish()
	}
}
