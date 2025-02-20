package keybindings

import (
	"github.com/jroimartin/gocui"
)

// GlobalKeybindings contains all global key bindings
type GlobalKeybindings struct {
	gui     *gocui.Gui
	onQuit  func() error
	onDump  func() error
	onLoad  func() error
	onSpace func() error
}

// NewGlobalKeybindings creates a new global keybindings handler
func NewGlobalKeybindings(
	gui *gocui.Gui,
	onQuit func() error,
	onDump func() error,
	onLoad func() error,
	onSpace func() error,
) *GlobalKeybindings {
	return &GlobalKeybindings{
		gui:     gui,
		onQuit:  onQuit,
		onDump:  onDump,
		onLoad:  onLoad,
		onSpace: onSpace,
	}
}

// Setup sets up all global key bindings
func (k *GlobalKeybindings) Setup() error {
	// Quit - available in all views
	views := []string{"", "migrations", "environments", "connection", "status", "commands"}
	for _, view := range views {
		if err := k.gui.SetKeybinding(view, gocui.KeyCtrlC, gocui.ModNone, k.quit); err != nil {
			return err
		}
		if err := k.gui.SetKeybinding(view, 'q', gocui.ModNone, k.quit); err != nil {
			return err
		}
	}

	// Global commands - only for empty view
	if err := k.gui.SetKeybinding("", 'd', gocui.ModNone, k.dump); err != nil {
		return err
	}

	if err := k.gui.SetKeybinding("", 'l', gocui.ModNone, k.load); err != nil {
		return err
	}

	if err := k.gui.SetKeybinding("", gocui.KeySpace, gocui.ModNone, k.showEnvironments); err != nil {
		return err
	}

	return nil
}

func (k *GlobalKeybindings) quit(g *gocui.Gui, v *gocui.View) error {
	// Handle quit
	return k.onQuit()
}

func (k *GlobalKeybindings) dump(g *gocui.Gui, v *gocui.View) error {
	// Handle dump
	return k.onDump()
}

func (k *GlobalKeybindings) load(g *gocui.Gui, v *gocui.View) error {
	// Handle load
	return k.onLoad()
}

func (k *GlobalKeybindings) showEnvironments(g *gocui.Gui, v *gocui.View) error {
	// Handle show environments
	return k.onSpace()
}
