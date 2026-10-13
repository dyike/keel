package sys

// TrayOptions is an immutable snapshot. Event 0 activates the icon; menu IDs
// are positive and never reused by the public controller.
type TrayOptions struct {
	PNG      []byte
	Template bool
	Tooltip  string
	Menu     []TrayMenuItem
	Event    func(uint32)
}

type TrayMenuItem struct {
	ID                           uint32
	Title                        string
	Disabled, Checked, Separator bool
	Children                     []TrayMenuItem
}

type TrayBackend interface {
	Update(TrayOptions) error
	Close() error
}
