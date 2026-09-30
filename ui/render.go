package ui

import (
	"fmt"
	"reflect"
)

func compile(children []Component) (map[string]*element, error) {
	index := make(map[string]*element)
	var visit func(Component, int) error
	visit = func(c Component, depth int) error {
		if c == nil || (reflect.ValueOf(c).Kind() == reflect.Pointer && reflect.ValueOf(c).IsNil()) {
			return fmt.Errorf("ui: nil component")
		}
		if depth > 64 || len(index) >= 1024 {
			return fmt.Errorf("ui: page exceeds component limit")
		}
		e := c.node()
		if _, ok := index[e.id]; ok {
			return fmt.Errorf("ui: component reused in the same page: %s", e.id)
		}
		e.mu.RLock()
		kind, inputType, variant, maxLength := e.kind, e.inputType, e.variant, e.maxLength
		e.mu.RUnlock()
		if kind == "input" {
			switch InputType(inputType) {
			case TextInput, PasswordInput, EmailInput, NumberInput, SearchInput:
			default:
				return fmt.Errorf("ui: unsupported input type %q", inputType)
			}
		}
		if kind == "button" {
			switch ButtonVariant(variant) {
			case Primary, Secondary, Danger:
			default:
				return fmt.Errorf("ui: unsupported button variant %q", variant)
			}
		}
		if maxLength < 0 {
			return fmt.Errorf("ui: negative MaxLength")
		}
		index[e.id] = e
		for _, child := range e.children {
			if err := visit(child, depth+1); err != nil {
				return err
			}
		}
		return nil
	}
	for _, child := range children {
		if err := visit(child, 0); err != nil {
			return nil, err
		}
	}
	return index, nil
}
