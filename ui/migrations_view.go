package ui

import (
	"fmt"

	"github.com/jroimartin/gocui"

	"dumper/config"
	"dumper/migrations"
)

const (
	confirmDialogView = "confirm-dialog"
	confirmButtonView = "confirm-button"
	cancelButtonView  = "cancel-button"
)

type MigrationsView struct {
	gui        *gocui.Gui
	currentEnv *config.Environment
	migrations []migrations.MigrationStatus
	needUpdate bool
	onLog      func(string, ...interface{}) // для логирования
}

func NewMigrationsView(g *gocui.Gui, onLog func(string, ...interface{})) *MigrationsView {
	return &MigrationsView{
		gui:        g,
		needUpdate: true,
		onLog:      onLog,
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
	if err := m.gui.SetKeybinding(migrationsView, gocui.KeyEnter, gocui.ModNone, m.showConfirmDialog); err != nil {
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
	m.migrations, err = migrations.GetMigrationStatus(m.currentEnv.DbDsn, m.currentEnv.MigrationsDir, m.onLog)
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

func (m *MigrationsView) showConfirmDialog(g *gocui.Gui, v *gocui.View) error {
	if len(m.migrations) == 0 {
		return nil
	}

	_, cy := v.Cursor()
	if cy == 0 { // Заголовок
		return nil
	}

	selectedMigration := m.migrations[cy-1]
	maxX, maxY := g.Size()
	width := 60
	height := 8
	x1 := (maxX - width) / 2
	y1 := (maxY - height) / 2
	x2 := x1 + width
	y2 := y1 + height

	// Основной диалог
	if v, err := g.SetView(confirmDialogView, x1, y1, x2, y2); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = " Подтверждение "
		v.Wrap = true
		v.Frame = true

		targetVersion := selectedMigration.ID
		currentVersion := int64(0)

		// Находим текущую версию (последняя примененная миграция)
		for _, m := range m.migrations {
			if m.Applied && m.ID > currentVersion {
				currentVersion = m.ID
			}
		}

		var action string
		if targetVersion > currentVersion {
			action = "применить миграции"
		} else {
			action = "откатить миграции"
		}

		fmt.Fprintf(v, "\n %s до версии %d?\n", action, targetVersion)
		fmt.Fprintf(v, " Текущая версия: %d\n", currentVersion)
		fmt.Fprintln(v, "\n Выберите действие:")

		// Кнопка подтверждения
		buttonX := (width - 25) / 3
		if confirmBtn, err := g.SetView(confirmButtonView, x1+buttonX, y2-2, x1+buttonX+10, y2); err != nil {
			if err != gocui.ErrUnknownView {
				return err
			}
			confirmBtn.Frame = true
			confirmBtn.FgColor = gocui.ColorGreen
			fmt.Fprint(confirmBtn, " Да (Y)")
		}

		// Кнопка отмены
		if cancelBtn, err := g.SetView(cancelButtonView, x2-buttonX-10, y2-2, x2-buttonX, y2); err != nil {
			if err != gocui.ErrUnknownView {
				return err
			}
			cancelBtn.Frame = true
			cancelBtn.FgColor = gocui.ColorRed
			fmt.Fprint(cancelBtn, " Нет (N)")
		}

		// Добавляем обработчики клавиш
		if err := g.SetKeybinding("", 'y', gocui.ModNone,
			func(g *gocui.Gui, v *gocui.View) error {
				return m.confirmMigration(targetVersion)
			}); err != nil {
			return err
		}

		if err := g.SetKeybinding("", 'n', gocui.ModNone, m.closeConfirmDialog); err != nil {
			return err
		}

		// Делаем кнопки кликабельными
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

		g.SetCurrentView(confirmDialogView)
	}

	return nil
}

func (m *MigrationsView) confirmMigration(targetVersion int64) error {
	if err := m.closeConfirmDialog(m.gui, nil); err != nil {
		return err
	}

	m.onLog("Начинаю миграцию на версию %d...", targetVersion)
	if err := migrations.MigrateTo(m.currentEnv.DbDsn, m.currentEnv.MigrationsDir, targetVersion, m.onLog); err != nil {
		m.onLog("Ошибка миграции: %v", err)
		return nil
	}

	m.onLog("Миграция успешно выполнена!")
	m.needUpdate = true // Обновим список после миграции
	m.gui.Update(func(g *gocui.Gui) error {
		return nil // Принудительно обновим UI
	})
	return nil
}

func (m *MigrationsView) closeConfirmDialog(g *gocui.Gui, v *gocui.View) error {
	g.DeleteKeybinding("", 'y', gocui.ModNone)
	g.DeleteKeybinding("", 'n', gocui.ModNone)
	g.DeleteView(confirmButtonView)
	g.DeleteView(cancelButtonView)
	g.DeleteView(confirmDialogView)
	g.SetCurrentView(migrationsView)
	return nil
}
