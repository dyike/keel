package kit

import (
	"errors"
	"github.com/dyike/keel/ui/el"
)

type InputToken = el.InputToken
type InputTokenSpan = el.InputTokenSpan
type InputContent = el.InputContent
type InputRange = el.InputRange

func NewInputContent(text string, tokens ...InputTokenSpan) (InputContent, error) {
	return el.NewInputContent(text, tokens...)
}

// Content returns an immutable draft, including atomic reference metadata.
func (v *InputView) Content() InputContent {
	if v.document != nil {
		return v.document.Content()
	}
	c, _ := NewInputContent(v.value)
	return c
}

// SetContent restores a draft and clears undo. Atomic references are supported
// by Input and TextArea; masks, password fields and text filters cannot be combined with references.
func (v *InputView) SetContent(c InputContent) error {
	if v.password || v.mask != nil || v.filter != "" || v.maxLen > 0 {
		return errors.New("input tokens cannot be combined with masks, passwords, filters or maximum length")
	}
	if v.document == nil {
		v.document = &el.InputDocument{}
	}
	v.document.SetContent(c)
	v.value = c.Text()
	return nil
}

// ReplaceWithToken replaces the current selection. Unlike SetContent, this is
// an edit, records undo and invokes OnChange. Read-only and disabled fields reject it.
func (v *InputView) ReplaceWithToken(token InputToken) error {
	if v.readOnly || v.disabled {
		return errors.New("input is not editable")
	}
	if v.document == nil {
		if err := v.SetContent(v.Content()); err != nil {
			return err
		}
	}
	if err := v.document.ReplaceWithToken(token); err != nil {
		return err
	}
	v.value = v.document.Content().Text()
	if v.onChange != nil {
		v.onChange(v.value)
	}
	return nil
}

func (v *InputView) OnTokenActivate(fn func(InputToken)) *InputView { v.onTokenActivate = fn; return v }

// ActivateToken opens the selected reference through the application's handler.
// Applications can bind this command to their preferred keyboard shortcut.
func (v *InputView) ActivateToken() bool {
	if v.disabled || v.document == nil || v.onTokenActivate == nil {
		return false
	}
	token, ok := v.document.SelectedToken()
	if ok {
		v.onTokenActivate(token)
	}
	return ok
}

// TokenRenderer customizes passive inline reference contents. See el.InputTokenRenderer.
func (v *InputView) TokenRenderer(fn el.InputTokenRenderer) *InputView {
	v.tokenRenderer = fn
	return v
}
