//go:build (!darwin && !linux && !windows) || ios || android

package sys

import "github.com/dyike/keel/native"

func NotificationAvailable() bool                               { return false }
func NotificationPermission(done func(error))                   { done(native.ErrUnsupported) }
func NotificationPost(id, title, body string, done func(error)) { done(native.ErrUnsupported) }
func NotificationRemove(id string, done func(error))            { done(native.ErrUnsupported) }
