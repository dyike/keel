package window

import (
	"strconv"
	"strings"
)

// startupTime is the X server timestamp a startup notification ID carries
// after "_TIME", the time of the user action that launched or activated
// us, or 0 when it has none. Window managers compare it with the user's
// last interaction to decide whether the activation may take focus.
func startupTime(id string) uint32 {
	i := strings.LastIndex(id, "_TIME")
	if i < 0 {
		return 0
	}
	digits := id[i+len("_TIME"):]
	if end := strings.IndexFunc(digits, func(r rune) bool { return r < '0' || r > '9' }); end >= 0 {
		digits = digits[:end]
	}
	t, err := strconv.ParseUint(digits, 10, 32)
	if err != nil {
		return 0
	}
	return uint32(t)
}

// activeWindowData is the _NET_ACTIVE_WINDOW request body: source 1 marks
// an application acting on a user's request, then the timestamp and the
// currently active window (none).
func activeWindowData(token string) [5]uint32 {
	return [5]uint32{1, startupTime(token), 0, 0, 0}
}
