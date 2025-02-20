package ui

import (
	"fmt"

	"github.com/jroimartin/gocui"
)

type LogsView struct {
	logs []string
	gui  *gocui.Gui
}

func NewLogsView(g *gocui.Gui) *LogsView {
	return &LogsView{
		logs: make([]string, 0),
		gui:  g,
	}
}

func (l *LogsView) Layout(maxX, maxY, startX int) error {
	if v, err := l.gui.SetView(statusView, startX, -1, maxX, maxY-2); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = " Лог операций "
		v.Wrap = true
		v.Autoscroll = false
	}

	// Обновляем логи
	if v, err := l.gui.View(statusView); err == nil {
		v.Clear()
		for i, log := range l.logs {
			if i == 0 {
				fmt.Fprintf(v, " %s\n --------------------------------\n", log)
			} else {
				fmt.Fprintf(v, " %s\n", log)
			}
		}
	}

	return nil
}

func (l *LogsView) AddLog(format string, a ...interface{}) {
	log := fmt.Sprintf(format, a...)
	l.logs = append([]string{log}, l.logs...) // Добавляем новые логи в начало слайса

	l.gui.Update(func(g *gocui.Gui) error {
		if v, err := g.View(statusView); err == nil {
			v.Clear()
			for i, l := range l.logs {
				if i == 0 {
					fmt.Fprintf(v, " %s\n --------------------------------\n", l)
				} else {
					fmt.Fprintf(v, " %s\n", l)
				}
			}
		}
		return nil
	})
}
