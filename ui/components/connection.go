package components

import (
	"fmt"

	"github.com/jroimartin/gocui"

	"dumper/config/db"
	"dumper/config/env"
	"dumper/ui/theme"
	"dumper/ui/views"
)

// ConnectionView represents the connection information display component
type ConnectionView struct {
	gui        *gocui.Gui
	currentEnv *env.Environment
	localDb    *db.Connection
}

// NewConnectionView creates a new connection view component
func NewConnectionView(g *gocui.Gui, localDb *db.Connection) *ConnectionView {
	return &ConnectionView{
		gui:     g,
		localDb: localDb,
	}
}

// Layout implements the views.Component interface
func (c *ConnectionView) Layout(maxX, maxY int) error {
	connectionWidth := maxX * 2 / 3
	if v, err := c.gui.SetView(views.ConnectionView, 0, 0, connectionWidth, theme.Dimensions.DialogMinHeight); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = fmt.Sprintf(" Environment: %s ", c.getCurrentEnvName())
		v.Wrap = true
		v.Frame = true
	}

	// Update connection info
	if v, err := c.gui.View(views.ConnectionView); err == nil {
		c.updateConnectionInfo(v)
	}

	return nil
}

// SetCurrentEnvironment updates the current environment
func (c *ConnectionView) SetCurrentEnvironment(env *env.Environment) {
	c.currentEnv = env
}

func (c *ConnectionView) getCurrentEnvName() string {
	if c.currentEnv != nil {
		return c.currentEnv.Name
	}
	return "not selected"
}

func (c *ConnectionView) updateConnectionInfo(v *gocui.View) {
	v.Clear()

	if c.currentEnv == nil {
		fmt.Fprintln(v, " Please select an environment...")
		return
	}

	fmt.Fprintln(v, " Local database:")
	fmt.Fprintf(v, "   Host:     %s\n", c.localDb.Host)
	fmt.Fprintf(v, "   Port:     %s\n", c.localDb.Port)
	fmt.Fprintf(v, "   Database: %s\n", c.localDb.Database)
	fmt.Fprintf(v, "   User:     %s\n", c.localDb.User)
	fmt.Fprintf(v, "   Password: %s\n", c.localDb.Password)

	fmt.Fprintln(v, "\n Remote database:")
	remote, err := db.ParseDSN(c.currentEnv.DbDsn)
	if err != nil {
		fmt.Fprintf(v, " Error parsing DSN: %v\n", err)
		return
	}

	fmt.Fprintf(v, "   Host:     %s\n", remote.Host)
	fmt.Fprintf(v, "   Database: %s\n", remote.Database)
}
