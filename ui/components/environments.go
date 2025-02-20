package components

import (
	"fmt"

	"github.com/jroimartin/gocui"

	"dumper/config/app"
	"dumper/config/env"
	"dumper/ui/theme"
	"dumper/ui/views"
)

// EnvironmentsView represents the environments list component
type EnvironmentsView struct {
	gui         *gocui.Gui
	cfg         *app.Config
	currentEnv  *env.Environment
	showEnvList bool
	onSelect    func(*env.Environment)
}

// NewEnvironmentsView creates a new environments view component
func NewEnvironmentsView(g *gocui.Gui, cfg *app.Config, onSelect func(*env.Environment)) *EnvironmentsView {
	env := &EnvironmentsView{
		gui:      g,
		cfg:      cfg,
		onSelect: onSelect,
	}

	// Set first environment as default without calling onSelect
	environments := cfg.GetEnvironments()
	if len(environments) > 0 {
		env.currentEnv = &environments[0]
	}

	return env
}

// Layout implements the views.Component interface
func (e *EnvironmentsView) Layout(maxX, maxY int) error {
	if !e.showEnvList {
		return nil
	}

	envWidth := theme.Dimensions.DialogMinWidth
	envHeight := len(e.cfg.GetEnvironments()) + 4
	x1 := (maxX - envWidth) / 2
	y1 := (maxY - envHeight) / 2
	x2 := x1 + envWidth
	y2 := y1 + envHeight

	if v, err := e.gui.SetView(views.EnvironmentsView, x1, y1, x2, y2); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Frame = true
		v.Title = " Select Environment "
		v.Highlight = true
		v.SelBgColor = theme.Colors.SelectionBg
		v.SelFgColor = theme.Colors.SelectionFg

		if err := e.setupKeybindings(); err != nil {
			return err
		}

		// Fill environments list
		environments := e.cfg.GetEnvironments()
		for i, env := range environments {
			fmt.Fprintf(v, " %d. %s\n", i+1, env.Name)
		}
		fmt.Fprintln(v, "\n Enter - select")

		e.gui.Cursor = true
		e.gui.SetCurrentView(views.EnvironmentsView)

		// Set cursor to current environment
		if e.currentEnv != nil {
			for i, env := range environments {
				if env.Name == e.currentEnv.Name {
					v.SetCursor(0, i)
					break
				}
			}
		} else {
			v.SetCursor(0, 0)
		}
	}

	return nil
}

func (e *EnvironmentsView) setupKeybindings() error {
	if err := e.gui.SetKeybinding(views.EnvironmentsView, gocui.KeyArrowUp, gocui.ModNone, e.up); err != nil {
		return err
	}
	if err := e.gui.SetKeybinding(views.EnvironmentsView, gocui.KeyArrowDown, gocui.ModNone, e.down); err != nil {
		return err
	}
	if err := e.gui.SetKeybinding(views.EnvironmentsView, gocui.KeyEnter, gocui.ModNone, e.select_); err != nil {
		return err
	}
	return nil
}

// Show displays the environments list
func (e *EnvironmentsView) Show() {
	e.showEnvList = true
}

// Hide hides the environments list
func (e *EnvironmentsView) Hide() {
	e.showEnvList = false
	// Delete keybindings first
	e.gui.DeleteKeybinding(views.EnvironmentsView, gocui.KeyArrowUp, gocui.ModNone)
	e.gui.DeleteKeybinding(views.EnvironmentsView, gocui.KeyArrowDown, gocui.ModNone)
	e.gui.DeleteKeybinding(views.EnvironmentsView, gocui.KeyEnter, gocui.ModNone)
	e.gui.DeleteView(views.EnvironmentsView)
	e.gui.Cursor = false
}

func (e *EnvironmentsView) up(g *gocui.Gui, v *gocui.View) error {
	if v != nil {
		_, cy := v.Cursor()
		if cy > 0 {
			v.SetCursor(0, cy-1)
		}
	}
	return nil
}

func (e *EnvironmentsView) down(g *gocui.Gui, v *gocui.View) error {
	if v != nil {
		_, cy := v.Cursor()
		environments := e.cfg.GetEnvironments()
		if cy < len(environments)-1 {
			v.SetCursor(0, cy+1)
		}
	}
	return nil
}

func (e *EnvironmentsView) select_(g *gocui.Gui, v *gocui.View) error {
	_, cy := v.Cursor()
	environments := e.cfg.GetEnvironments()

	if cy >= 0 && cy < len(environments) {
		e.currentEnv = &environments[cy]
		if e.onSelect != nil {
			e.onSelect(e.currentEnv)
		}
	}

	e.Hide()
	return nil
}

// GetCurrentEnvironment returns the currently selected environment
func (e *EnvironmentsView) GetCurrentEnvironment() *env.Environment {
	return e.currentEnv
}
