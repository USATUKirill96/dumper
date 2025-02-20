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
	// Имя контейнера
	containerName = "local_postgres"

	// Образ PostgreSQL
	postgresImage = "postgres:16"

	// Пароль для пользователя postgres в контейнере
	postgresPassword = "pass"

	// Сколько ждём, пока контейнер поднимет PostgreSQL
	maxWaitSeconds = 30

	// Директория для дампов
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
	// Парсим флаги
	debug := flag.Bool("debug", false, "Включить отладочный вывод")
	flag.Parse()

	// Загружаем конфигурацию
	cfg, err := app.LoadConfig("config.yaml", dumpsDir)
	if err != nil {
		fmt.Printf("Ошибка загрузки конфига: %v\n", err)
		os.Exit(1)
	}

	// Создаем конфигурацию PostgreSQL
	pgConfig := database.PostgresConfig{
		ContainerName:  containerName,
		Image:          postgresImage,
		Password:       postgresPassword,
		MaxWaitSeconds: maxWaitSeconds,
		Debug:          *debug,
	}

	// Создаем локальное подключение с временной базой
	localDb := &db.Connection{
		Host:     "localhost",
		Port:     "5432",
		User:     "postgres",
		Password: postgresPassword,
		Database: "postgres", // Будет изменено при выборе окружения
		SslMode:  "disable",
	}

	app := &application{
		cfg:      cfg,
		localDb:  localDb,
		pgConfig: pgConfig,
		debug:    *debug,
	}

	// Создаем UI
	ui, err := ui.New(cfg, localDb,
		// Функция для создания дампа
		app.dump,
		// Функция для загрузки дампа
		app.load,
	)
	if err != nil {
		fmt.Printf("Ошибка создания UI: %v\n", err)
		os.Exit(1)
	}

	app.ui = ui

	// Запускаем UI
	if err := ui.Run(); err != nil && err != gocui.ErrQuit {
		fmt.Printf("Ошибка в UI: %v\n", err)
		os.Exit(1)
	}
}

func (a *application) dump() error {
	currentEnv := a.ui.GetCurrentEnvironment()
	if currentEnv == nil {
		return fmt.Errorf("environment not selected")
	}

	dumpFile := filepath.Join(dumpsDir, fmt.Sprintf("%s.sql", a.localDb.Database))

	// Выполняем операцию дампа
	if err := database.DumpDatabase(currentEnv.DbDsn, dumpFile, a.debug); err != nil {
		return fmt.Errorf("failed to create dump: %w", err)
	}

	return nil
}

func (a *application) load() error {
	dumpFile := filepath.Join(dumpsDir, fmt.Sprintf("%s.sql", a.localDb.Database))
	return database.LoadDump(a.pgConfig, a.localDb.Database, dumpFile)
}
