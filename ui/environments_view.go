package ui

import (
	"fmt"

	"github.com/jroimartin/gocui"

	"dumper/config"
)

type EnvironmentsView struct {
	gui         *gocui.Gui
	cfg         *config.Config
	currentEnv  *config.Environment
	showEnvList bool
	onSelect    func(*config.Environment)
}

func NewEnvironmentsView(g *gocui.Gui, cfg *config.Config, onSelect func(*config.Environment)) *EnvironmentsView {
	return &EnvironmentsView{
		gui:      g,
		cfg:      cfg,
		onSelect: onSelect,
	}
}

func (e *EnvironmentsView) Layout(maxX, maxY int) error {
	if !e.showEnvList {
		return nil
	}

	envWidth := 40
	envHeight := len(e.cfg.Environments) + 4
	x1 := (maxX - envWidth) / 2
	y1 := (maxY - envHeight) / 2
	x2 := x1 + envWidth
	y2 := y1 + envHeight

	if v, err := e.gui.SetView(envListView, x1, y1, x2, y2); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Frame = true
		v.Title = ""
		v.Highlight = true
		v.SelBgColor = gocui.ColorGreen
		v.SelFgColor = gocui.ColorBlack

		if err := e.setupKeybindings(); err != nil {
			return err
		}

		// Заполняем список окружений
		for i, env := range e.cfg.Environments {
			fmt.Fprintf(v, " %d. %s\n", i+1, env.Name)
		}
		fmt.Fprintln(v, "\n Enter - выбрать")

		e.gui.Cursor = true
		e.gui.SetCurrentView(envListView)

		// Устанавливаем курсор на текущее окружение
		if e.currentEnv != nil {
			for i, env := range e.cfg.Environments {
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
	if err := e.gui.SetKeybinding(envListView, gocui.KeyArrowUp, gocui.ModNone, e.up); err != nil {
		return err
	}
	if err := e.gui.SetKeybinding(envListView, gocui.KeyArrowDown, gocui.ModNone, e.down); err != nil {
		return err
	}
	if err := e.gui.SetKeybinding(envListView, gocui.KeyEnter, gocui.ModNone, e.select_); err != nil {
		return err
	}
	return nil
}

func (e *EnvironmentsView) Show() {
	e.showEnvList = true
}

func (e *EnvironmentsView) Hide() {
	e.showEnvList = false
	e.gui.DeleteView(envListView)
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
		if cy < len(e.cfg.Environments)-1 {
			v.SetCursor(0, cy+1)
		}
	}
	return nil
}

func (e *EnvironmentsView) select_(g *gocui.Gui, v *gocui.View) error {
	_, cy := v.Cursor()

	if cy >= 0 && cy < len(e.cfg.Environments) {
		e.currentEnv = &e.cfg.Environments[cy]
		if e.onSelect != nil {
			e.onSelect(e.currentEnv)
		}
	}

	e.Hide()
	return nil
}

func (e *EnvironmentsView) GetCurrentEnvironment() *config.Environment {
	return e.currentEnv
}
