//go:build darwin && !ios && cgo

package sys

import (
	"errors"
	"github.com/dyike/keel/native"
	"strings"
	"testing"
)

func TestNotificationFailureClassification(t *testing.T) {
	for _, tc := range []struct {
		name, domain string
		code         int64
		want         error
	}{
		{"permission", "UNErrorDomain", 1, native.ErrPermissionDenied},
		{"same code other domain", "NSCocoaErrorDomain", 1, native.ErrFailed},
		{"unknown notification error", "UNErrorDomain", 9999, native.ErrFailed},
		{"signed system code", "NSOSStatusErrorDomain", -600, native.ErrFailed},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := notificationFailure(tc.domain, tc.code, "通知不可用 %s")
			if !errors.Is(err, tc.want) {
				t.Fatalf("got %v, want wrapped %v", err, tc.want)
			}
			other := native.ErrFailed
			if tc.want == native.ErrFailed {
				other = native.ErrPermissionDenied
			}
			if errors.Is(err, other) {
				t.Fatalf("misclassified: %v", err)
			}
			if !strings.Contains(err.Error(), tc.domain) || !strings.Contains(err.Error(), "通知不可用 %s") {
				t.Fatalf("lost system details: %v", err)
			}
		})
	}
}
