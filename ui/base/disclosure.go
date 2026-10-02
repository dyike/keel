package base

// Disclosure is open/closed state that can be disabled, for popovers, menus,
// collapsibles and dialogs. OnChange runs only for changes the user makes
// through Change, not for Set.
type Disclosure struct {
	open, disabled bool
	OnChange       func(open bool)
}

func (d *Disclosure) Open() bool     { return d.open }
func (d *Disclosure) Disabled() bool { return d.disabled }

// Set opens or closes without calling OnChange. A disabled disclosure stays
// closed.
func (d *Disclosure) Set(open bool) { d.open = open && !d.disabled }

// SetDisabled disables it, which also closes it.
func (d *Disclosure) SetDisabled(on bool) {
	d.disabled = on
	if on {
		d.open = false
	}
}

// Change opens or closes as the user asked and calls OnChange if that
// changed anything. It reports whether it did.
func (d *Disclosure) Change(open bool) bool {
	open = open && !d.disabled
	if open == d.open {
		return false
	}
	d.open = open
	if d.OnChange != nil {
		d.OnChange(open)
	}
	return true
}

// Toggle is Change(!Open()).
func (d *Disclosure) Toggle() bool { return d.Change(!d.open) }
