package ui

import (
	"fmt"

	"github.com/jroimartin/gocui"

	"dumper/config/app"
	"dumper/config/db"
	"dumper/config/env"
	"dumper/ui/components"
	"dumper/ui/keybindings"
	"dumper/ui/layout"
)

// UI represents the main application UI
type UI struct {
	gui              *gocui.Gui
	mainLayout       *layout.MainLayout
	keybindings      *keybindings.GlobalKeybindings
	logsView         *components.LogsView
	connectionView   *components.ConnectionView
	migrationsView   *components.MigrationsView
	environmentsView *components.EnvironmentsView
	cfg              *app.Config
	localDb          *db.Connection
	onDump           func() error
	onLoad           func() error
}

// New creates a new UI instance
func New(cfg *app.Config, localDb *db.Connection, onDump, onLoad func() error) (*UI, error) {
	gui, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		return nil, fmt.Errorf("failed to create GUI: %w", err)
	}

	ui := &UI{
		gui:     gui,
		cfg:     cfg,
		localDb: localDb,
		onDump:  onDump,
		onLoad:  onLoad,
	}

	// Initialize layout and components FIRST
	ui.mainLayout = layout.NewMainLayout(gui)
	ui.logsView = components.NewLogsView(gui)
	ui.connectionView = components.NewConnectionView(gui, localDb)
	ui.migrationsView = components.NewMigrationsView(gui, ui.logsView.AddLog)
	ui.environmentsView = components.NewEnvironmentsView(gui, cfg, ui.onEnvironmentSelected)

	// Add components to layout
	ui.mainLayout.AddComponent(ui.connectionView)
	ui.mainLayout.AddComponent(ui.migrationsView)
	ui.mainLayout.AddComponent(ui.logsView)
	ui.mainLayout.AddComponent(ui.environmentsView)

	// Set up GUI manager AFTER components are initialized
	gui.SetManager(ui.mainLayout)
	gui.Cursor = true
	gui.Mouse = true

	// Set up global keybindings
	ui.keybindings = keybindings.NewGlobalKeybindings(
		gui,
		func() error { return gocui.ErrQuit },
		func() error { return ui.handleDump() },
		func() error { return ui.handleLoad() },
		func() error { return ui.handleShowEnvironments() },
	)

	if err := ui.keybindings.Setup(); err != nil {
		return nil, fmt.Errorf("error setting up keybindings: %w", err)
	}

	// Update commands bar
	ui.mainLayout.UpdateCommandsBar(" Space - Select Environment | d - Dump Database | l - Load Database | q/Ctrl+C - Quit")

	// Select first environment by default
	environments := cfg.GetEnvironments()
	if len(environments) > 0 {
		ui.onEnvironmentSelected(&environments[0])
	}

	return ui, nil
}

// Run starts the UI main loop
func (ui *UI) Run() error {
	defer ui.gui.Close()
	return ui.gui.MainLoop()
}

// Event handlers
func (ui *UI) onEnvironmentSelected(env *env.Environment) {
	ui.localDb.Database = env.Name
	ui.connectionView.SetCurrentEnvironment(env)
	ui.migrationsView.SetCurrentEnvironment(env)
	ui.logsView.AddLog("Selected environment: %s", env.Name)

	// Update commands bar
	ui.mainLayout.UpdateCommandsBar(" Space - Select Environment | d - Dump Database | l - Load Database | q/Ctrl+C - Quit")
}

func (ui *UI) GetCurrentEnvironment() *env.Environment {
	return ui.environmentsView.GetCurrentEnvironment()
}

func (ui *UI) handleShowEnvironments() error {
	ui.environmentsView.Show()
	return nil
}

func (ui *UI) handleDump() error {
	env := ui.GetCurrentEnvironment()
	if env == nil {
		ui.logsView.AddLog("No environment selected")
		return nil
	}

	if err := ui.onDump(); err != nil {
		ui.logsView.AddLog("Error: %v", err)
		return err
	}
	ui.logsView.AddLog("Database dump completed successfully!")

	return nil
}

func (ui *UI) handleLoad() error {
	env := ui.GetCurrentEnvironment()
	if env == nil {
		ui.logsView.AddLog("No environment selected")
		return nil
	}

	ui.logsView.AddLog("Loading dump into local database...")
	if err := ui.onLoad(); err != nil {
		ui.logsView.AddLog("Error: %v", err)
		return nil
	}
	ui.logsView.AddLog("Dump loaded successfully!")
	return nil
}

// Update forces an immediate UI update and waits for it to complete
func (ui *UI) Update() {
	done := make(chan struct{})
	ui.gui.Update(func(g *gocui.Gui) error {
		defer close(done)
		return nil
	})
	<-done
}

// ForceRedraw forces an immediate UI redraw
func (ui *UI) ForceRedraw() {
	ui.gui.Update(func(g *gocui.Gui) error {
		return ui.mainLayout.Layout(g)
	})
}

// AddLog adds a log message to the logs view
func (ui *UI) AddLog(format string, a ...interface{}) {
	ui.logsView.AddLog(format, a...)
}
