package components

import (
	"fmt"

	"github.com/jroimartin/gocui"

	"dumper/ui/theme"
	"dumper/ui/views"
)

// LogsView represents the logs display component
type LogsView struct {
	gui        *gocui.Gui
	logs       []string
	needUpdate bool
}

// NewLogsView creates a new logs view component
func NewLogsView(g *gocui.Gui) *LogsView {
	return &LogsView{
		gui:        g,
		logs:       make([]string, 0),
		needUpdate: false,
	}
}

// Layout implements the views.Component interface
func (l *LogsView) Layout(maxX, maxY int) error {
	if v, err := l.gui.SetView(views.StatusView, maxX*2/3, 0, maxX-1, maxY-theme.Dimensions.CommandHeight-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = " Logs "
		v.Wrap = true
		v.Autoscroll = false // Disable autoscroll since we show newest first
		v.Frame = true
		l.needUpdate = true
	}

	// Update logs content only when needed
	if l.needUpdate {
		l.updateContent()
		l.needUpdate = false
	}
	return nil
}

// AddLog adds a new log message to the list
func (l *LogsView) AddLog(format string, a ...interface{}) {
	log := fmt.Sprintf(format, a...)
	l.logs = append(l.logs, log)

	// Обновляем содержимое сразу и помечаем, что обновление не требуется
	l.updateContent()
	l.needUpdate = false
}

// updateContent updates the view content without using gui.Update
func (l *LogsView) updateContent() {
	v, err := l.gui.View(views.StatusView)
	if err != nil {
		return
	}

	v.Clear()
	// Show logs in reverse order (newest first)
	for i := len(l.logs) - 1; i >= 0; i-- {
		if i == len(l.logs)-1 {
			// Highlight the latest log with underline
			fmt.Fprintf(v, "%s\n", l.logs[i])
			fmt.Fprintf(v, "- - - - - - - - - - - - - - - - - - - - - - - - - - - -\n")
		} else {
			fmt.Fprintf(v, "%s\n", l.logs[i])
		}
	}
}
