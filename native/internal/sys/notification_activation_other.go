//go:build !linux || android

package sys

// Other platforms retain their interactive behavior without a Linux token.
func NotificationPostActivated(id, title, body string, onActivate func(string), done func(error)) {
	var click func()
	if onActivate != nil {
		click = func() { onActivate("") }
	}
	NotificationPostInteractive(id, title, body, click, done)
}
