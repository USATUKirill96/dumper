package ui

import (
	"fmt"
	"log"

	"github.com/jroimartin/gocui"

	"dumper/config"
)

const (
	// Названия view
	envListView    = "environments"
	commandsView   = "commands"
	connectionView = "connection"
	statusView     = "status"
	migrationsView = "migrations"
)

type UI struct {
	gui              *gocui.Gui
	cfg              *config.Config
	localDb          *config.DbConnection
	onDump           func() error
	onLoad           func() error
	logsView         *LogsView
	connectionView   *ConnectionView
	migrationsView   *MigrationsView
	environmentsView *EnvironmentsView
}

func New(cfg *config.Config, localDb *config.DbConnection, onDump, onLoad func() error) (*UI, error) {
	gui, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания GUI: %w", err)
	}

	ui := &UI{
		gui:     gui,
		cfg:     cfg,
		localDb: localDb,
		onDump:  onDump,
		onLoad:  onLoad,
	}

	// Инициализируем компоненты
	ui.logsView = NewLogsView(gui)
	ui.connectionView = NewConnectionView(gui, localDb)
	ui.migrationsView = NewMigrationsView(gui)
	ui.environmentsView = NewEnvironmentsView(gui, cfg, ui.onEnvironmentSelected)

	// Выбираем первое окружение по умолчанию
	if len(cfg.Environments) > 0 {
		ui.onEnvironmentSelected(&cfg.Environments[0])
	}

	gui.SetManager(ui)
	gui.Cursor = false
	gui.Mouse = false

	// Глобальные кейбинды
	if err := gui.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, ui.quit); err != nil {
		return nil, err
	}
	if err := gui.SetKeybinding("", 'd', gocui.ModNone, ui.dump); err != nil {
		return nil, err
	}
	if err := gui.SetKeybinding("", 'l', gocui.ModNone, ui.load); err != nil {
		return nil, err
	}
	if err := gui.SetKeybinding("", 'q', gocui.ModNone, ui.quit); err != nil {
		return nil, err
	}
	if err := gui.SetKeybinding("", gocui.KeySpace, gocui.ModNone, ui.showEnvironments); err != nil {
		return nil, err
	}

	return ui, nil
}

func (ui *UI) Run() error {
	defer ui.gui.Close()
	return ui.gui.MainLoop()
}

func (ui *UI) Layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()

	// Отрисовываем компоненты
	if err := ui.connectionView.Layout(maxX, maxY); err != nil {
		return err
	}

	if err := ui.migrationsView.Layout(maxX, maxY); err != nil {
		return err
	}

	if err := ui.logsView.Layout(maxX, maxY, maxX*2/3); err != nil {
		return err
	}

	// Команды (одна строка внизу)
	if v, err := g.SetView(commandsView, -1, maxY-2, maxX, maxY); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Frame = false
	}

	// Обновляем команды
	if v, err := g.View(commandsView); err == nil {
		v.Clear()
		fmt.Fprintf(v, " d - сохранить копию базы | l - восстановить из копии | пробел - выбор окружения | q - выход")
	}

	// Отрисовываем список окружений поверх всего, если он открыт
	if err := ui.environmentsView.Layout(maxX, maxY); err != nil {
		return err
	}

	return nil
}

func (ui *UI) onEnvironmentSelected(env *config.Environment) {
	ui.localDb.Database = env.Name
	ui.connectionView.SetCurrentEnvironment(env)
	ui.migrationsView.SetCurrentEnvironment(env)
	ui.logsView.AddLog("Выбрано окружение: %s", env.Name)
}

func (ui *UI) showEnvironments(g *gocui.Gui, v *gocui.View) error {
	ui.environmentsView.Show()
	return nil
}

func (ui *UI) dump(g *gocui.Gui, v *gocui.View) error {
	if ui.environmentsView.GetCurrentEnvironment() == nil {
		ui.logsView.AddLog("Сначала выберите окружение!")
		return nil
	}

	ui.logsView.AddLog("Создаю дамп...")
	if err := ui.onDump(); err != nil {
		ui.logsView.AddLog("Ошибка: %v", err)
		return nil
	}
	ui.logsView.AddLog("Дамп успешно создан!")
	return nil
}

func (ui *UI) load(g *gocui.Gui, v *gocui.View) error {
	if ui.environmentsView.GetCurrentEnvironment() == nil {
		ui.logsView.AddLog("Сначала выберите окружение!")
		return nil
	}

	ui.logsView.AddLog("Загружаю дамп в локальную базу...")
	if err := ui.onLoad(); err != nil {
		ui.logsView.AddLog("Ошибка: %v", err)
		return nil
	}
	ui.logsView.AddLog("Дамп успешно загружен!")
	return nil
}

func (ui *UI) quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}

// GetCurrentEnvironment возвращает текущее выбранное окружение
func (ui *UI) GetCurrentEnvironment() *config.Environment {
	return ui.environmentsView.GetCurrentEnvironment()
}

// SelectEnvironment - для обратной совместимости
func SelectEnvironment(cfg *config.Config) (*config.Environment, error) {
	log.Fatal("Этот метод больше не поддерживается")
	return nil, nil
}

// ShowMenu - для обратной совместимости
func ShowMenu(env *config.Environment, dbConn *config.DbConnection) string {
	log.Fatal("Этот метод больше не поддерживается")
	return ""
}
