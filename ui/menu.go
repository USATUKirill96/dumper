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
)

type UI struct {
	gui         *gocui.Gui
	cfg         *config.Config
	currentEnv  *config.Environment
	localDb     *config.DbConnection
	onDump      func() error
	onLoad      func() error
	logs        []string
	showEnvList bool
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
		logs:    make([]string, 0),
	}

	// Выбираем первое окружение по умолчанию
	if len(cfg.Environments) > 0 {
		ui.currentEnv = &cfg.Environments[0]
		ui.localDb.Database = ui.currentEnv.Name
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

	// Кейбинды для списка окружений
	if err := gui.SetKeybinding(envListView, gocui.KeyArrowUp, gocui.ModNone, ui.envUp); err != nil {
		return nil, err
	}
	if err := gui.SetKeybinding(envListView, gocui.KeyArrowDown, gocui.ModNone, ui.envDown); err != nil {
		return nil, err
	}
	if err := gui.SetKeybinding(envListView, gocui.KeyEnter, gocui.ModNone, ui.selectEnv); err != nil {
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

	// Основная информация (левая часть экрана)
	connectionWidth := maxX * 2 / 3
	if v, err := g.SetView(connectionView, -1, -1, connectionWidth, maxY-2); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = fmt.Sprintf(" Окружение: %s ", ui.getCurrentEnvName())
		v.Wrap = true
	}

	// Обновляем информацию о подключении
	if v, err := g.View(connectionView); err == nil {
		ui.updateConnectionInfo(v)
	}

	// Логи (правая часть экрана)
	if v, err := g.SetView(statusView, connectionWidth, -1, maxX, maxY-2); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = " Лог операций "
		v.Wrap = true
		v.Autoscroll = false
	}

	// Обновляем логи
	if v, err := g.View(statusView); err == nil {
		v.Clear()
		for _, log := range ui.logs {
			fmt.Fprintf(v, " %s\n", log)
		}
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

	// Список окружений (поверх основного экрана, если открыт)
	if ui.showEnvList {
		envWidth := 40
		envHeight := len(ui.cfg.Environments) + 4
		x1 := (maxX - envWidth) / 2
		y1 := (maxY - envHeight) / 2
		x2 := x1 + envWidth
		y2 := y1 + envHeight

		if v, err := g.SetView(envListView, x1, y1, x2, y2); err != nil {
			if err != gocui.ErrUnknownView {
				return err
			}
			v.Frame = true
			v.Title = ""
			v.Highlight = true
			v.SelBgColor = gocui.ColorGreen
			v.SelFgColor = gocui.ColorBlack

			// Заполняем список окружений (без отступа сверху)
			for i, env := range ui.cfg.Environments {
				fmt.Fprintf(v, " %d. %s\n", i+1, env.Name)
			}
			fmt.Fprintln(v, "\n Enter - выбрать")

			g.Cursor = true
			g.SetCurrentView(envListView)

			// Устанавливаем курсор на текущее окружение
			if ui.currentEnv != nil {
				for i, env := range ui.cfg.Environments {
					if env.Name == ui.currentEnv.Name {
						v.SetCursor(0, i) // теперь i, а не i+1, так как нет отступа
						break
					}
				}
			} else {
				v.SetCursor(0, 0) // первое окружение
			}
		}
	}

	return nil
}

func (ui *UI) getCurrentEnvName() string {
	if ui.currentEnv != nil {
		return ui.currentEnv.Name
	}
	return "не выбрано"
}

func (ui *UI) showEnvironments(g *gocui.Gui, v *gocui.View) error {
	ui.showEnvList = true
	return nil
}

func (ui *UI) hideEnvironments(g *gocui.Gui, v *gocui.View) error {
	ui.showEnvList = false
	g.DeleteView(envListView)
	g.Cursor = false
	g.SetCurrentView(connectionView)
	return nil
}

func (ui *UI) updateConnectionInfo(v *gocui.View) {
	v.Clear()

	if ui.currentEnv == nil {
		fmt.Fprintln(v, " Выберите окружение...")
		return
	}

	fmt.Fprintln(v, " Локальная база:")
	fmt.Fprintf(v, "   Хост:         %s\n", ui.localDb.Host)
	fmt.Fprintf(v, "   Порт:         %s\n", ui.localDb.Port)
	fmt.Fprintf(v, "   База:         %s\n", ui.localDb.Database)
	fmt.Fprintf(v, "   Пользователь: %s\n", ui.localDb.User)
	fmt.Fprintf(v, "   Пароль:       %s\n", ui.localDb.Password)
	fmt.Fprintf(v, "   DSN:          %s\n", ui.localDb.GetDSN())

	fmt.Fprintln(v, "\n Удаленная база:")
	remote, err := config.ParseDSN(ui.currentEnv.DbDsn)
	if err != nil {
		fmt.Fprintf(v, " Ошибка парсинга DSN: %v\n", err)
		return
	}

	fmt.Fprintf(v, "   Хост:         %s\n", remote.Host)
	fmt.Fprintf(v, "   Порт:         %s\n", remote.Port)
	fmt.Fprintf(v, "   База:         %s\n", remote.Database)
	fmt.Fprintf(v, "   Пользователь: %s\n", remote.User)
}

func (ui *UI) envUp(g *gocui.Gui, v *gocui.View) error {
	if v != nil {
		_, cy := v.Cursor()
		if cy > 0 { // теперь можно подняться до 0
			v.SetCursor(0, cy-1)
		}
	}
	return nil
}

func (ui *UI) envDown(g *gocui.Gui, v *gocui.View) error {
	if v != nil {
		_, cy := v.Cursor()
		if cy < len(ui.cfg.Environments)-1 { // -1 потому что индексы с 0
			v.SetCursor(0, cy+1)
		}
	}
	return nil
}

func (ui *UI) selectEnv(g *gocui.Gui, v *gocui.View) error {
	_, cy := v.Cursor()

	// Теперь cy это прямой индекс в массиве
	if cy >= 0 && cy < len(ui.cfg.Environments) {
		ui.currentEnv = &ui.cfg.Environments[cy]
		ui.localDb.Database = ui.currentEnv.Name
		ui.addLog("Выбрано окружение: %s", ui.currentEnv.Name)
	}

	// Закрываем окно в любом случае
	ui.showEnvList = false
	g.DeleteView(envListView)
	g.Cursor = false
	g.SetCurrentView(connectionView)
	return nil
}

func (ui *UI) dump(g *gocui.Gui, v *gocui.View) error {
	if ui.currentEnv == nil {
		ui.addLog("Сначала выберите окружение!")
		return nil
	}

	ui.addLog("Создаю дамп...")
	if err := ui.onDump(); err != nil {
		ui.addLog("Ошибка: %v", err)
		return nil
	}
	ui.addLog("Дамп успешно создан!")
	return nil
}

func (ui *UI) load(g *gocui.Gui, v *gocui.View) error {
	if ui.currentEnv == nil {
		ui.addLog("Сначала выберите окружение!")
		return nil
	}

	ui.addLog("Загружаю дамп в локальную базу...")
	if err := ui.onLoad(); err != nil {
		ui.addLog("Ошибка: %v", err)
		return nil
	}
	ui.addLog("Дамп успешно загружен!")
	return nil
}

func (ui *UI) quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}

// GetCurrentEnvironment возвращает текущее выбранное окружение
func (ui *UI) GetCurrentEnvironment() *config.Environment {
	return ui.currentEnv
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

func (ui *UI) addLog(format string, a ...interface{}) {
	log := fmt.Sprintf(format, a...)
	ui.logs = append([]string{log}, ui.logs...) // Добавляем новые логи в начало слайса
	ui.gui.Update(func(g *gocui.Gui) error {
		if v, err := g.View(statusView); err == nil {
			v.Clear()
			for i, l := range ui.logs {
				if i == 0 {
					// Последний лог выделяем пунктиром
					fmt.Fprintf(v, " %s\n --------------------------------\n", l)
				} else {
					fmt.Fprintf(v, " %s\n", l)
				}
			}
		}
		return nil
	})
}
