package notification

import (
	"errors"
	"github.com/dyike/keel/native"
	"testing"
	"time"
)

func result(t *testing.T, call func(func(error)), want error) {
	t.Helper()
	ch := make(chan error, 2)
	call(func(err error) { ch <- err })
	select {
	case err := <-ch:
		if !errors.Is(err, want) {
			t.Fatalf("got %v want %v", err, want)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("completion not delivered")
	}
}

func TestInvalidMessages(t *testing.T) {
	for _, m := range []Message{{}, {ID: "x"}, {ID: "x\x00y", Title: "a"}, {ID: "x", Title: "a\x00b"}, {ID: "x", Body: string([]byte{255})}} {
		result(t, func(done func(error)) { Post(m, done) }, native.ErrInvalidArgument)
	}
	for _, id := range []string{"", "a\x00b", string([]byte{255})} {
		result(t, func(done func(error)) { Remove(id, done) }, native.ErrInvalidArgument)
	}
	Post(Message{}, nil)
	Remove("", nil)
}

func TestUnsupportedProcess(t *testing.T) {
	// Never ask for permission or post notifications from automated tests.
	if Available() {
		t.Skip("bundled application: use the manual notification example")
	}
	result(t, RequestPermission, native.ErrUnsupported)
	result(t, func(done func(error)) { Post(Message{ID: "test", Title: "Test"}, done) }, native.ErrUnsupported)
	result(t, func(done func(error)) { Remove("test", done) }, native.ErrUnsupported)
	result(t, func(done func(error)) {
		Post(Message{ID: "click-test", Title: "Test", OnClick: func() { t.Error("unsupported post activated") }}, done)
	}, native.ErrUnsupported)
	result(t, func(done func(error)) {
		Post(Message{ID: "activation-test", Title: "Test", OnActivate: func(Activation) { t.Error("unsupported post activated") }}, done)
	}, native.ErrUnsupported)
}
