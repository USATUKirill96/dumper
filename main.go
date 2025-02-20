package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"dumper/config"
	"dumper/database"
	"dumper/ui"
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

func main() {
	// Парсим флаги
	debug := flag.Bool("debug", false, "Включить отладочный вывод")
	flag.Parse()

	// Загружаем конфигурацию
	cfg, err := config.LoadConfig("config.yaml", dumpsDir)
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

	// Выбираем начальное окружение
	env, err := ui.SelectEnvironment(cfg)
	if err != nil {
		fmt.Printf("Ошибка выбора окружения: %v\n", err)
		os.Exit(1)
	}

	// Создаем локальное подключение
	localDb := &config.DbConnection{
		Host:     "localhost",
		Port:     "5432",
		User:     "postgres",
		Password: postgresPassword,
		Database: env.Name,
		SslMode:  "disable",
	}

	for {
		choice := ui.ShowMenu(env, localDb)
		switch choice {
		case "1":
			dumpFile := filepath.Join(dumpsDir, fmt.Sprintf("%s.sql", env.Name))
			if err := database.DumpDatabase(env.DbDsn, dumpFile, *debug); err != nil {
				fmt.Printf("Ошибка при дампе: %v\n", err)
			} else {
				fmt.Printf("Дамп успешно создан в файле %s\n", dumpFile)
			}
		case "2":
			dumpFile := filepath.Join(dumpsDir, fmt.Sprintf("%s.sql", env.Name))
			if err := database.LoadDump(pgConfig, env.Name, dumpFile); err != nil {
				fmt.Printf("Ошибка при загрузке дампа: %v\n", err)
			} else {
				fmt.Println("Данные успешно загружены в локальную базу!")
			}
		case "3":
			if env, err = ui.SelectEnvironment(cfg); err != nil {
				fmt.Printf("Ошибка выбора окружения: %v\n", err)
			}
			localDb.Database = env.Name
		case "4":
			fmt.Println("До свидания!")
			return
		default:
			fmt.Println("Неизвестная команда")
		}
	}
}
