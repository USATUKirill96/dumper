package components

import (
	"fmt"

	"github.com/jroimartin/gocui"

	"dumper/config/db"
	"dumper/config/env"
	"dumper/migrations"
	"dumper/ui/theme"
	"dumper/ui/views"
)

const (
	migrationsView    = views.MigrationsView
	confirmDialogView = views.ConfirmDialogView
	confirmButtonView = views.ConfirmButtonView
	cancelButtonView  = views.CancelButtonView
)

// MigrationsView represents the migrations list and management UI component
type MigrationsView struct {
	gui         *gocui.Gui
	currentEnv  *env.Environment
	migrations  []migrations.MigrationStatus
	needUpdate  bool
	onLog       func(string, ...interface{}) // logging function
	localDb     *db.Connection
	isMigrating bool
}

// NewMigrationsView creates a new migrations view component
func NewMigrationsView(g *gocui.Gui, localDb *db.Connection, onLog func(string, ...interface{})) *MigrationsView {
	return &MigrationsView{
		gui:        g,
		needUpdate: true,
		onLog:      onLog,
		localDb:    localDb,
	}
}

// Layout implements the gocui.Manager interface
func (m *MigrationsView) Layout(maxX, maxY int) error {
	connectionWidth := maxX * 2 / 3
	migrationY := theme.Dimensions.DialogMinHeight + theme.Dimensions.ElementSpacing

	if v, err := m.gui.SetView(migrationsView,
		theme.Dimensions.PaddingX,
		migrationY,
		connectionWidth-2,
		maxY-theme.Dimensions.CommandHeight-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = " Migrations "
		v.Wrap = false
		v.Highlight = true
		v.SelBgColor = theme.Colors.SelectionBg
		v.SelFgColor = theme.Colors.SelectionFg

		if err := m.setupKeybindings(); err != nil {
			return err
		}

		if !m.isEnvironmentListVisible() {
			m.gui.SetCurrentView(migrationsView)
		}
	}

	// Update migrations list if needed
	if v, err := m.gui.View(migrationsView); err == nil && m.needUpdate {
		m.updateMigrationsList(v)
		m.needUpdate = false
	}

	return nil
}

// SetCurrentEnvironment updates the current environment and refreshes the migrations list
func (m *MigrationsView) SetCurrentEnvironment(env *env.Environment) {
	m.currentEnv = env
	m.needUpdate = true

	// Return focus to migrations view after update
	m.gui.Update(func(g *gocui.Gui) error {
		if !m.isEnvironmentListVisible() {
			g.SetCurrentView(migrationsView)
		}
		return nil
	})
}

func (m *MigrationsView) setupKeybindings() error {
	// Только навигация
	if err := m.gui.SetKeybinding(migrationsView, gocui.KeyArrowUp, gocui.ModNone, m.up); err != nil {
		return err
	}
	if err := m.gui.SetKeybinding(migrationsView, gocui.KeyArrowDown, gocui.ModNone, m.down); err != nil {
		return err
	}
	if err := m.gui.SetKeybinding(migrationsView, gocui.KeyArrowLeft, gocui.ModNone, m.start); err != nil {
		return err
	}
	if err := m.gui.SetKeybinding(migrationsView, gocui.KeyArrowRight, gocui.ModNone, m.end); err != nil {
		return err
	}
	if err := m.gui.SetKeybinding(migrationsView, gocui.KeyEnter, gocui.ModNone, m.showConfirmDialog); err != nil {
		return err
	}

	return nil
}

func (m *MigrationsView) isEnvironmentListVisible() bool {
	_, err := m.gui.View("environments")
	return err == nil
}

func (m *MigrationsView) updateMigrationsList(v *gocui.View) error {
	v.Clear()

	if m.currentEnv == nil || m.currentEnv.MigrationsDir == "" {
		fmt.Fprintln(v, " Migrations directory not specified")
		return nil
	}

	if m.isMigrating {
		fmt.Fprintln(v, " Migration in progress...")
		return nil
	}

	var err error
	m.migrations, err = migrations.GetMigrationStatus(m.localDb.GetDSN(), m.currentEnv.MigrationsDir, m.onLog)
	if err != nil {
		fmt.Fprintf(v, " Error getting migrations status: %v\n", err)
		return err
	}

	if len(m.migrations) == 0 {
		fmt.Fprintln(v, " No migrations found")
		return nil
	}

	var applied int
	for _, m := range m.migrations {
		if m.Applied {
			applied++
		}
	}

	fmt.Fprintf(v, " Applied: %d/%d\n", applied, len(m.migrations))

	for _, m := range m.migrations {
		status := "✓"
		if !m.Applied {
			status = " "
		}
		fmt.Fprintf(v, " [%s] %s\n", status, m.ShortName)
	}

	if err := v.SetOrigin(0, 0); err != nil {
		m.onLog("Error setting origin: %v", err)
		return err
	}

	if err := v.SetCursor(0, 1); err != nil {
		m.onLog("Error setting cursor: %v", err)
		return err
	}

	// Ensure migrations view is active
	if !m.isEnvironmentListVisible() {
		if _, err := m.gui.SetCurrentView(migrationsView); err != nil {
			m.onLog("Error setting current view: %v", err)
			return err
		}
	}

	return nil
}

// Navigation methods
func (m *MigrationsView) up(g *gocui.Gui, v *gocui.View) error {
	if len(m.migrations) == 0 {
		return nil
	}

	ox, oy := v.Origin()
	_, cy := v.Cursor()

	if cy > 1 {
		if err := v.SetCursor(0, cy-1); err != nil {
			if oy > 0 {
				if err := v.SetOrigin(ox, oy-1); err != nil {
					m.onLog("Error setting origin: %v", err)
					return err
				}
				if err := v.SetCursor(0, cy-1); err != nil {
					m.onLog("Error setting cursor: %v", err)
					return err
				}
			} else {
				m.onLog("Error setting cursor: %v", err)
				return err
			}
		}
	}
	return nil
}

func (m *MigrationsView) down(g *gocui.Gui, v *gocui.View) error {
	if len(m.migrations) == 0 {
		return nil
	}

	ox, oy := v.Origin()
	_, cy := v.Cursor()
	_, height := v.Size()
	maxY := len(m.migrations) + 1 // +1 for header

	if cy < maxY-1 {
		if err := v.SetCursor(0, cy+1); err != nil {
			if oy < maxY-height {
				if err := v.SetOrigin(ox, oy+1); err != nil {
					m.onLog("Error setting origin: %v", err)
					return err
				}
				if err := v.SetCursor(0, cy+1); err != nil {
					m.onLog("Error setting cursor: %v", err)
					return err
				}
			} else {
				m.onLog("Error setting cursor: %v", err)
				return err
			}
		}
	}
	return nil
}

func (m *MigrationsView) start(g *gocui.Gui, v *gocui.View) error {
	if len(m.migrations) == 0 {
		return nil
	}

	if err := v.SetOrigin(0, 0); err != nil {
		m.onLog("Error setting origin: %v", err)
		return err
	}
	if err := v.SetCursor(0, 1); err != nil {
		m.onLog("Error setting cursor: %v", err)
		return err
	}
	return nil
}

func (m *MigrationsView) end(g *gocui.Gui, v *gocui.View) error {
	if len(m.migrations) == 0 {
		return nil
	}

	maxY := len(m.migrations) + 1 // +1 for header
	_, height := v.Size()

	originY := 0
	if maxY > height {
		originY = maxY - height
	}

	if err := v.SetOrigin(0, originY); err != nil {
		m.onLog("Error setting origin: %v", err)
		return err
	}

	lastVisibleLine := maxY - originY - 1
	if lastVisibleLine >= height {
		lastVisibleLine = height - 1
	}
	if err := v.SetCursor(0, lastVisibleLine); err != nil {
		m.onLog("Error setting cursor: %v", err)
		return err
	}
	return nil
}

// Dialog methods
func (m *MigrationsView) showConfirmDialog(g *gocui.Gui, v *gocui.View) error {
	if len(m.migrations) == 0 {
		return nil
	}

	_, cy := v.Cursor()
	if cy == 0 { // Header
		return nil
	}

	if cy-1 >= len(m.migrations) {
		m.onLog("Error: invalid cursor position")
		return fmt.Errorf("invalid cursor position")
	}

	selectedMigration := m.migrations[cy-1]
	maxX, maxY := g.Size()
	width := theme.Dimensions.DialogMinWidth
	height := theme.Dimensions.DialogMinHeight
	x1 := (maxX - width) / 2
	y1 := (maxY - height) / 2
	x2 := x1 + width
	y2 := y1 + height

	// Main dialog
	if v, err := g.SetView(confirmDialogView, x1, y1, x2, y2); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = " Confirm Migration "
		v.Wrap = true
		v.Frame = true

		targetVersion := selectedMigration.ID
		currentVersion := int64(0)

		// Find current version (last applied migration)
		for _, m := range m.migrations {
			if m.Applied && m.ID > currentVersion {
				currentVersion = m.ID
			}
		}

		var action string
		if targetVersion > currentVersion {
			action = "apply migrations"
		} else {
			action = "rollback migrations"
		}

		fmt.Fprintf(v, "\n Do you want to %s to version %d?\n", action, targetVersion)
		fmt.Fprintf(v, " Current version: %d\n", currentVersion)
		fmt.Fprintln(v, "\n Choose action:")

		// Confirm button
		buttonX := (width - 25) / 3
		if confirmBtn, err := g.SetView(confirmButtonView,
			x1+buttonX, y2-2,
			x1+buttonX+theme.Dimensions.ButtonWidth, y2); err != nil {
			if err != gocui.ErrUnknownView {
				return err
			}
			confirmBtn.Frame = true
			confirmBtn.FgColor = theme.Colors.ButtonConfirmFg
			fmt.Fprint(confirmBtn, " Yes (Y)")
		}

		// Cancel button
		if cancelBtn, err := g.SetView(cancelButtonView,
			x2-buttonX-theme.Dimensions.ButtonWidth, y2-2,
			x2-buttonX, y2); err != nil {
			if err != gocui.ErrUnknownView {
				return err
			}
			cancelBtn.Frame = true
			cancelBtn.FgColor = theme.Colors.ButtonCancelFg
			fmt.Fprint(cancelBtn, " No (N)")
		}

		// Add key handlers
		if err := g.SetKeybinding("", 'y', gocui.ModNone,
			func(g *gocui.Gui, v *gocui.View) error {
				return m.confirmMigration(targetVersion)
			}); err != nil {
			return err
		}

		if err := g.SetKeybinding("", 'n', gocui.ModNone, m.closeConfirmDialog); err != nil {
			return err
		}

		// Make buttons clickable
		if err := g.SetKeybinding(confirmButtonView, gocui.MouseLeft, gocui.ModNone,
			func(g *gocui.Gui, v *gocui.View) error {
				return m.confirmMigration(targetVersion)
			}); err != nil {
			return err
		}

		if err := g.SetKeybinding(cancelButtonView, gocui.MouseLeft, gocui.ModNone,
			m.closeConfirmDialog); err != nil {
			return err
		}

		if _, err := g.SetCurrentView(confirmDialogView); err != nil {
			m.onLog("Error setting current view: %v", err)
			return err
		}
	}

	return nil
}

func (m *MigrationsView) confirmMigration(targetVersion int64) error {
	if err := m.closeConfirmDialog(m.gui, nil); err != nil {
		m.onLog("Error closing dialog: %v", err)
		return err
	}

	m.onLog("Starting migration to version %d...", targetVersion)

	if m.currentEnv == nil {
		m.onLog("Error: environment not selected")
		return fmt.Errorf("environment not selected")
	}

	if m.currentEnv.MigrationsDir == "" {
		m.onLog("Error: migrations directory not specified")
		return fmt.Errorf("migrations directory not specified")
	}

	if m.localDb == nil {
		m.onLog("Error: local database not initialized")
		return fmt.Errorf("local database not initialized")
	}

	// Устанавливаем флаг миграции и обновляем UI
	m.isMigrating = true
	m.needUpdate = true
	m.gui.Update(func(g *gocui.Gui) error {
		if v, err := g.View(migrationsView); err == nil {
			return m.updateMigrationsList(v)
		}
		return nil
	})

	// Запускаем миграцию в отдельной горутине
	go func() {
		if err := migrations.MigrateTo(m.localDb.GetDSN(), m.currentEnv.MigrationsDir, targetVersion, m.onLog); err != nil {
			m.onLog("Migration error: %v", err)
		} else {
			m.onLog("Migration completed successfully!")
		}

		// Сбрасываем флаг миграции и обновляем UI
		m.isMigrating = false
		m.needUpdate = true
		m.gui.Update(func(g *gocui.Gui) error {
			if v, err := g.View(migrationsView); err == nil {
				return m.updateMigrationsList(v)
			}
			return nil
		})
	}()

	return nil
}

func (m *MigrationsView) closeConfirmDialog(g *gocui.Gui, v *gocui.View) error {
	// Delete keybindings first
	g.DeleteKeybinding("", 'y', gocui.ModNone)
	g.DeleteKeybinding("", 'n', gocui.ModNone)

	// Delete views if they exist
	if v := g.CurrentView(); v != nil && v.Name() == confirmDialogView {
		if _, err := g.SetCurrentView(migrationsView); err != nil {
			m.onLog("Error setting current view: %v", err)
			return err
		}
	}

	views := []string{confirmButtonView, cancelButtonView, confirmDialogView}
	for _, name := range views {
		if v, err := g.View(name); err == nil && v != nil {
			if err := g.DeleteView(name); err != nil {
				m.onLog("Error deleting view %s: %v", name, err)
				return err
			}
		}
	}

	return nil
}
