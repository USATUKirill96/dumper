package layout

import (
	"fmt"

	"github.com/jroimartin/gocui"

	"dumper/ui/theme"
	"dumper/ui/views"
)

// MainLayout manages the main application layout
type MainLayout struct {
	gui        *gocui.Gui
	components []views.Component
	commands   string
}

// NewMainLayout creates a new main layout manager
func NewMainLayout(gui *gocui.Gui) *MainLayout {
	return &MainLayout{
		gui:        gui,
		components: make([]views.Component, 0),
		commands:   " Space - Select Environment | d - Create Dump | l - Load Dump | q/Ctrl+C - Exit",
	}
}

// AddComponent adds a component to the layout
func (l *MainLayout) AddComponent(c views.Component) {
	l.components = append(l.components, c)
}

// Layout implements the gocui.Manager interface
func (l *MainLayout) Layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()

	// Layout all components
	for _, c := range l.components {
		if err := c.Layout(maxX, maxY); err != nil {
			return fmt.Errorf("error laying out component: %w", err)
		}
	}

	commandsY := maxY - theme.Dimensions.CommandHeight

	// Commands bar (always at the bottom)
	if v, err := g.SetView(views.CommandsView, 0, commandsY, maxX-1, maxY-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Frame = true
		v.Title = " Commands "
		v.Wrap = true
		v.BgColor = theme.Colors.DefaultBg
		v.FgColor = theme.Colors.DefaultFg
	}

	// Always update commands text
	if v, err := g.View(views.CommandsView); err == nil {
		v.Clear()
		fmt.Fprintln(v, l.commands)
	}

	return nil
}

// UpdateCommandsBar updates the commands bar text
func (l *MainLayout) UpdateCommandsBar(text string) error {
	l.commands = text
	return nil
}
