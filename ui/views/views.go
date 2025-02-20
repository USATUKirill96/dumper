package views

// View names
const (
	// Main views
	EnvironmentsView = "environments"
	MigrationsView   = "migrations"
	ConnectionView   = "connection"
	StatusView       = "status"
	CommandsView     = "commands"

	// Dialog views
	ConfirmDialogView = "confirm-dialog"
	ConfirmButtonView = "confirm-button"
	CancelButtonView  = "cancel-button"
)

// Component represents a UI component that can be laid out
type Component interface {
	// Layout renders the component
	Layout(maxX, maxY int) error
}

// Focusable represents a component that can receive focus
type Focusable interface {
	// SetFocus sets the focus to this component
	SetFocus() error
}

// Updatable represents a component that can be updated
type Updatable interface {
	// Update updates the component's state
	Update() error
}

// KeybindingHandler represents a component that handles key bindings
type KeybindingHandler interface {
	// SetupKeybindings sets up the component's key bindings
	SetupKeybindings() error
}
