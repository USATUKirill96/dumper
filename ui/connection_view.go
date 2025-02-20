package ui

import (
	"fmt"

	"github.com/jroimartin/gocui"

	"dumper/config"
)

type ConnectionView struct {
	gui        *gocui.Gui
	currentEnv *config.Environment
	localDb    *config.DbConnection
}

func NewConnectionView(g *gocui.Gui, localDb *config.DbConnection) *ConnectionView {
	return &ConnectionView{
		gui:     g,
		localDb: localDb,
	}
}

func (c *ConnectionView) Layout(maxX, maxY int) error {
	connectionWidth := maxX * 2 / 3
	if v, err := c.gui.SetView(connectionView, -1, -1, connectionWidth, maxY-2); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = fmt.Sprintf(" Окружение: %s ", c.getCurrentEnvName())
		v.Wrap = true
	}

	// Обновляем информацию о подключении
	if v, err := c.gui.View(connectionView); err == nil {
		c.updateConnectionInfo(v)
	}

	return nil
}

func (c *ConnectionView) SetCurrentEnvironment(env *config.Environment) {
	c.currentEnv = env
}

func (c *ConnectionView) getCurrentEnvName() string {
	if c.currentEnv != nil {
		return c.currentEnv.Name
	}
	return "не выбрано"
}

func (c *ConnectionView) updateConnectionInfo(v *gocui.View) {
	v.Clear()

	if c.currentEnv == nil {
		fmt.Fprintln(v, " Выберите окружение...")
		return
	}

	fmt.Fprintln(v, " Локальная база:")
	fmt.Fprintf(v, "   Хост:         %s\n", c.localDb.Host)
	fmt.Fprintf(v, "   Порт:         %s\n", c.localDb.Port)
	fmt.Fprintf(v, "   База:         %s\n", c.localDb.Database)
	fmt.Fprintf(v, "   Пользователь: %s\n", c.localDb.User)
	fmt.Fprintf(v, "   Пароль:       %s\n", c.localDb.Password)
	fmt.Fprintf(v, "   DSN:          %s\n", c.localDb.GetDSN())

	fmt.Fprintln(v, "\n Удаленная база:")
	remote, err := config.ParseDSN(c.currentEnv.DbDsn)
	if err != nil {
		fmt.Fprintf(v, " Ошибка парсинга DSN: %v\n", err)
		return
	}

	fmt.Fprintf(v, "   Хост:         %s\n", remote.Host)
	fmt.Fprintf(v, "   База:         %s\n", remote.Database)
}
