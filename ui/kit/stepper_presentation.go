package kit

// StepperNavigation controls which enabled steps the user may select.
type StepperNavigation uint8

const (
	StepperNavigationNone StepperNavigation = iota
	StepperNavigationCompleted
	StepperNavigationAll
)

// Navigation chooses read-only, completed-only, or unrestricted navigation.
// Programmatic SetValue is unaffected. Clicking the current step is a no-op.
func (v *StepperView) Navigation(mode StepperNavigation) *StepperView {
	if mode <= StepperNavigationAll {
		v.navigation = mode
	}
	return v
}

// TextCenter centers text and rich content. In horizontal layout the content
// sits below the marker, with connectors between neighboring marker centers.
// In vertical layout the marker stays beside the content.
func (v *StepperView) TextCenter(on bool) *StepperView { v.textCenter = on; return v }

// Horizontal restores left-to-right layout after Vertical.
func (v *StepperView) Horizontal() *StepperView { v.vertical = false; return v }

func (v *StepperView) canNavigate(index int) bool {
	if v.disabled || index < 0 || index >= len(v.steps) || v.steps[index].Disabled {
		return false
	}
	return v.navigation == StepperNavigationAll || v.navigation == StepperNavigationCompleted && index < v.current
}
