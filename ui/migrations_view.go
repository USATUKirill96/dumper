package ui

import (
	"fmt"

	"github.com/jroimartin/gocui"

	"dumper/config"
	"dumper/migrations"
)

type MigrationsView struct {
	gui        *gocui.Gui
	currentEnv *config.Environment
	migrations []migrations.MigrationStatus
	needUpdate bool
}

func NewMigrationsView(g *gocui.Gui) *MigrationsView {
	return &MigrationsView{
		gui:        g,
		needUpdate: true,
	}
}

func (m *MigrationsView) Layout(maxX, maxY int) error {
	connectionWidth := maxX * 2 / 3
	migrationY := 12

	if v, err := m.gui.SetView(migrationsView, 1, migrationY, connectionWidth-2, maxY-3); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = ""
		v.Wrap = false
		v.Highlight = true
		v.SelBgColor = gocui.ColorGreen
		v.SelFgColor = gocui.ColorBlack

		if err := m.setupKeybindings(); err != nil {
			return err
		}

		if !m.isEnvironmentListVisible() {
			m.gui.SetCurrentView(migrationsView)
		}
	}

	// Обновляем список миграций только если нужно
	if v, err := m.gui.View(migrationsView); err == nil && m.needUpdate {
		m.updateMigrationsList(v)
		m.needUpdate = false
	}

	return nil
}

func (m *MigrationsView) SetCurrentEnvironment(env *config.Environment) {
	m.currentEnv = env
	m.needUpdate = true

	// После обновления списка возвращаем фокус на view с миграциями
	m.gui.Update(func(g *gocui.Gui) error {
		if !m.isEnvironmentListVisible() {
			g.SetCurrentView(migrationsView)
		}
		return nil
	})
}

func (m *MigrationsView) setupKeybindings() error {
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
	return nil
}

func (m *MigrationsView) isEnvironmentListVisible() bool {
	_, err := m.gui.View(envListView)
	return err == nil
}

func (m *MigrationsView) updateMigrationsList(v *gocui.View) {
	v.Clear()

	if m.currentEnv == nil || m.currentEnv.MigrationsDir == "" {
		fmt.Fprintln(v, " Путь к миграциям не указан")
		return
	}

	var err error
	m.migrations, err = migrations.GetMigrationStatus(m.currentEnv.DbDsn, m.currentEnv.MigrationsDir)
	if err != nil {
		fmt.Fprintf(v, " Ошибка получения статуса миграций: %v\n", err)
		return
	}

	if len(m.migrations) == 0 {
		fmt.Fprintln(v, " Миграции не найдены")
		return
	}

	var applied int
	for _, m := range m.migrations {
		if m.Applied {
			applied++
		}
	}

	fmt.Fprintf(v, " Применено: %d/%d\n", applied, len(m.migrations))

	for _, m := range m.migrations {
		status := "✓"
		if !m.Applied {
			status = " "
		}
		fmt.Fprintf(v, " [%s] %s\n", status, m.ShortName)
	}

	v.SetOrigin(0, 0)
	v.SetCursor(0, 1)

	// Убеждаемся, что view с миграциями активен
	if !m.isEnvironmentListVisible() {
		m.gui.SetCurrentView(migrationsView)
	}
}

func (m *MigrationsView) up(g *gocui.Gui, v *gocui.View) error {
	if len(m.migrations) == 0 {
		return nil
	}

	ox, oy := v.Origin()
	_, cy := v.Cursor()

	if cy > 1 {
		// Пробуем сначала просто передвинуть курсор
		if err := v.SetCursor(0, cy-1); err != nil {
			// Если не получилось, значит надо скроллить
			if oy > 0 {
				if err := v.SetOrigin(ox, oy-1); err != nil {
					return err
				}
				return v.SetCursor(0, cy-1)
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
	maxY := len(m.migrations) + 1 // +1 для учета заголовка

	if cy < maxY-1 {
		// Пробуем сначала просто передвинуть курсор
		if err := v.SetCursor(0, cy+1); err != nil {
			// Если не получилось, значит надо скроллить
			if oy < maxY-height {
				if err := v.SetOrigin(ox, oy+1); err != nil {
					return err
				}
				return v.SetCursor(0, cy+1)
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
		return err
	}
	return v.SetCursor(0, 1)
}

func (m *MigrationsView) end(g *gocui.Gui, v *gocui.View) error {
	if len(m.migrations) == 0 {
		return nil
	}

	maxY := len(m.migrations) + 1 // +1 для учета заголовка
	_, height := v.Size()

	// Вычисляем позицию origin для последней страницы
	originY := 0
	if maxY > height {
		originY = maxY - height
	}

	if err := v.SetOrigin(0, originY); err != nil {
		return err
	}

	// Устанавливаем курсор на последнюю строку
	lastVisibleLine := maxY - originY - 1
	if lastVisibleLine >= height {
		lastVisibleLine = height - 1
	}
	return v.SetCursor(0, lastVisibleLine)
}
