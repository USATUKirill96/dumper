package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"dumper/config/app"
	"dumper/config/db"
	"dumper/database"
	"dumper/ui"

	"github.com/jroimartin/gocui"
)

const (
	// Container name
	containerName = "local_postgres"

	// PostgreSQL image
	postgresImage = "postgres:16"

	// Password for postgres user in container
	postgresPassword = "pass"

	// How long to wait for PostgreSQL to start in container
	maxWaitSeconds = 30

	// Directory for dumps
	dumpsDir = "dumps"
)

type application struct {
	ui       *ui.UI
	cfg      *app.Config
	localDb  *db.Connection
	pgConfig database.PostgresConfig
	debug    bool
}

func main() {
	// Parse flags
	debug := flag.Bool("debug", false, "Enable debug output")
	flag.Parse()

	// Load configuration
	cfg, err := app.LoadConfig("config.yaml", dumpsDir)
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Create PostgreSQL configuration
	pgConfig := database.PostgresConfig{
		ContainerName:  containerName,
		Image:          postgresImage,
		Password:       postgresPassword,
		MaxWaitSeconds: maxWaitSeconds,
		Debug:          *debug,
	}

	// Create local connection with temporary database
	localDb := &db.Connection{
		Host:     "localhost",
		Port:     "5432",
		User:     "postgres",
		Password: postgresPassword,
		Database: "postgres", // Will be changed when environment is selected
		SslMode:  "disable",
	}

	app := &application{
		cfg:      cfg,
		localDb:  localDb,
		pgConfig: pgConfig,
		debug:    *debug,
	}

	// Create UI
	ui, err := ui.New(cfg, localDb,
		// Function to create dump
		app.dump,
		// Function to load dump
		app.load,
	)
	if err != nil {
		fmt.Printf("Error creating UI: %v\n", err)
		os.Exit(1)
	}

	app.ui = ui

	// Run UI
	if err := ui.Run(); err != nil && err != gocui.ErrQuit {
		fmt.Printf("UI error: %v\n", err)
		os.Exit(1)
	}
}

func (a *application) dump() error {
	currentEnv := a.ui.GetCurrentEnvironment()
	if currentEnv == nil {
		return fmt.Errorf("environment not selected")
	}

	dumpFile := filepath.Join(dumpsDir, fmt.Sprintf("%s.sql", a.localDb.Database))

	// Execute dump operation
	if err := database.DumpDatabase(currentEnv.DbDsn, dumpFile, a.debug); err != nil {
		return fmt.Errorf("failed to create dump: %w", err)
	}

	return nil
}

func (a *application) load() error {
	dumpFile := filepath.Join(dumpsDir, fmt.Sprintf("%s.sql", a.localDb.Database))
	return database.LoadDump(a.pgConfig, a.localDb.Database, dumpFile)
}
