package kit

import (
	"crypto/rand"
	"errors"
	"github.com/dyike/keel/ui/core"
	"strconv"
	"sync"
)

// NoticeDelivery chooses where a notice is sent. Default follows the host.
type NoticeDelivery uint8

const (
	NoticeDeliveryDefault NoticeDelivery = iota
	NoticeInApp
	NoticeSystemOnly
	NoticeInAppAndSystem
)

// NoticeSystemBackend bridges kit to an application's platform integration.
// Each method must call done exactly once, including failures. Methods run on a
// worker goroutine; they must not directly mutate UI state. Requests are serialized
// per Notifier, including asynchronous completion, to prevent post/remove races.
type NoticeSystemBackend interface {
	Post(id, title, body string, done func(error))
	Remove(id string, done func(error))
}

// NoticeSystemInteractiveBackend optionally adds system click delivery. activated
// may run on any goroutine; kit forwards it to the UI frame lock.
type NoticeSystemInteractiveBackend interface {
	NoticeSystemBackend
	PostInteractive(id, title, body string, activated func(), done func(error))
}

// NoticeActivation carries optional platform window-activation data.
// Token is opaque; it may be empty and is consumed by the application's backend.
type NoticeActivation struct{ Token string }

// NoticeSystemActivationBackend adds activation data to system click delivery.
// Notifier prefers it when a backend implements both interactive interfaces.
type NoticeSystemActivationBackend interface {
	NoticeSystemBackend
	PostActivated(id, title, body string, activated func(NoticeActivation), done func(error))
}

// OnSystemActivation receives platform activation data on the UI frame lock,
// before OnSystemActivate and the notice's close/click callbacks.
func (v *NotifierView) OnSystemActivation(fn func(NoticeActivation)) *NotifierView {
	v.systemActivation = fn
	return v
}

// OnSystemActivate sets the application's window activation hook (e.g. Raise).
// It runs before the notice's close and click callbacks, on the UI frame lock.
func (v *NotifierView) OnSystemActivate(fn func()) *NotifierView { v.systemActivate = fn; return v }

func (v *NotifierView) activateSystem(id int) { v.activateSystemWith(id, NoticeActivation{}) }
func (v *NotifierView) activateSystemWith(id int, activation NoticeActivation) {
	for _, current := range v.items {
		if current.id != id || !current.systemPosted || current.Delivery == NoticeInApp {
			continue
		}
		n, _ := v.removeNotice(id) // detach before any user callback, preventing reentrancy
		v.systemRequest(n, true)
		if v.systemActivation != nil {
			v.systemActivation(activation)
		}
		if v.systemActivate != nil {
			v.systemActivate()
		}
		if n.inApp() && n.OnClose != nil {
			n.OnClose()
		}
		if n.OnClick != nil {
			n.OnClick()
		}
		return
	}
}

// NoticeSystemResult reports a completed backend request on the UI frame lock.
type NoticeSystemResult struct {
	ID       int
	Removing bool
	Err      error
}

var ErrNoticeSystemUnavailable = errors.New("kit: no system notification backend")

// SystemBackend configures future notices. Existing notices retain their backend
// so later updates and removal reach the service that originally received them.
// onResult may be nil; delivery errors are also available from SystemError.
func (v *NotifierView) SystemBackend(backend NoticeSystemBackend, onResult func(NoticeSystemResult)) *NotifierView {
	v.systemBackend, v.systemResult = backend, onResult
	return v
}

// Delivery sets the default for subsequent Notify and Update calls. It does not
// migrate existing notices. Default restores in-app delivery; invalid values are ignored.
func (v *NotifierView) Delivery(mode NoticeDelivery) *NotifierView {
	if mode <= NoticeInAppAndSystem {
		v.delivery = mode
	}
	return v
}

// SystemError returns the most recent completed backend request's error.
func (v *NotifierView) SystemError() error { return v.systemError }
func (v *NotifierView) resolveDelivery(mode NoticeDelivery) NoticeDelivery {
	if mode == NoticeDeliveryDefault || mode > NoticeInAppAndSystem {
		mode = v.delivery
	}
	if mode == NoticeDeliveryDefault {
		mode = NoticeInApp
	}
	return mode
}
func (n notice) inApp() bool { return n.Delivery != NoticeSystemOnly && !n.hidden }
func (v *NotifierView) postSystem(n *notice) {
	if n.Delivery == NoticeInApp || n.Title == "" && n.Body == "" {
		return
	}
	n.systemPosted = true
	v.systemRequest(*n, false)
}
func (v *NotifierView) systemRequest(n notice, remove bool) {
	backend := n.backend
	if v.systemPrefix == "" {
		v.systemPrefix = "keel-" + rand.Text()
	}
	id := v.systemPrefix + "/" + strconv.Itoa(n.id)
	report := v.systemResult
	v.outbox.enqueue(func(done func()) {
		var once sync.Once
		complete := func(err error) {
			once.Do(func() {
				core.Update(func() {
					v.systemError = err
					if report != nil {
						report(NoticeSystemResult{ID: n.id, Removing: remove, Err: err})
					}
				})
				done()
			})
		}
		if backend == nil {
			complete(ErrNoticeSystemUnavailable)
			return
		}
		if remove {
			backend.Remove(id, complete)
		} else if interactive, ok := backend.(NoticeSystemActivationBackend); ok {
			interactive.PostActivated(id, n.Title, n.Body, func(activation NoticeActivation) { core.Update(func() { v.activateSystemWith(n.id, activation) }) }, complete)
		} else if interactive, ok := backend.(NoticeSystemInteractiveBackend); ok {
			interactive.PostInteractive(id, n.Title, n.Body, func() { core.Update(func() { v.activateSystem(n.id) }) }, complete)
		} else {
			backend.Post(id, n.Title, n.Body, complete)
		}
	})
}
func (v *NotifierView) expire(id int) {
	for i := range v.items {
		n := &v.items[i]
		if n.id != id || !n.inApp() {
			continue
		}
		if !n.systemPosted {
			v.Dismiss(id)
			return
		}
		n.hidden = true
		closed := n.OnClose
		if closed != nil {
			closed()
		}
		return
	}
}

type noticeOutbox struct {
	mu      sync.Mutex
	jobs    []func(func())
	running bool
}

func (q *noticeOutbox) enqueue(job func(func())) {
	q.mu.Lock()
	q.jobs = append(q.jobs, job)
	if q.running {
		q.mu.Unlock()
		return
	}
	q.running = true
	q.mu.Unlock()
	go func() {
		for {
			q.mu.Lock()
			if len(q.jobs) == 0 {
				q.running = false
				q.mu.Unlock()
				return
			}
			job := q.jobs[0]
			q.jobs[0] = nil
			q.jobs = q.jobs[1:]
			q.mu.Unlock()
			finished := make(chan struct{})
			job(func() { close(finished) })
			<-finished
		}
	}()
}
