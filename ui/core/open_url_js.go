//go:build js && wasm

package core

import (
	"fmt"
	"syscall/js"
)

func openURL(raw string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("open URL: %v", r)
		}
	}()
	js.Global().Call("open", raw, "_blank", "noopener,noreferrer")
	return nil
}
