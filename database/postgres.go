package database

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

type PostgresConfig struct {
	ContainerName  string
	Image          string
	Password       string
	MaxWaitSeconds int
	Debug          bool
}

// debugPrintf выводит отладочную информацию только если включен режим debug
func debugPrintf(debug bool, format string, a ...interface{}) {
	if debug {
		fmt.Printf(format, a...)
	}
}

// debugCmd выполняет команду с выводом в stdout/stderr только в режиме debug
func debugCmd(cmd *exec.Cmd, debug bool) error {
	if debug {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	return cmd.Run()
}

// DumpDatabase делает дамп базы данных
func DumpDatabase(dsn string, dumpFile string, debug bool) error {
	// Сначала проверим, есть ли таблицы в исходной базе
	checkCmd := exec.Command("bash", "-c",
		fmt.Sprintf("psql '%s' -t -c \"SELECT COUNT(*) FROM pg_tables WHERE schemaname = 'public';\"",
			dsn))
	output, err := checkCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ошибка проверки таблиц в исходной базе: %w\nВывод: %s", err, string(output))
	}
	debugPrintf(debug, "Количество таблиц в исходной базе: %s", string(output))

	// Делаем дамп с дополнительными опциями
	cmdString := fmt.Sprintf(
		"pg_dump "+
			"--format=plain "+ // Текстовый формат
			"--no-owner "+ // Без владельцев
			"--no-privileges "+ // Без привилегий
			"--clean "+ // Очистка перед восстановлением
			"--if-exists "+ // Добавляем IF EXISTS в DROP
			"--schema=public "+ // Только схема public
			"--disable-triggers "+ // Отключаем триггеры при восстановлении
			"'%s' > %s",
		dsn, dumpFile)

	cmd := exec.Command("bash", "-c", cmdString)
	if debug {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	debugPrintf(debug, "Выполняю команду дампа...\n")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ошибка при создании дампа: %w", err)
	}

	return nil
}

// LoadDump загружает дамп в локальную базу данных
func LoadDump(cfg PostgresConfig, dbName string, dumpFile string) error {
	// Проверяем, запущен ли контейнер
	checkCmd := exec.Command("docker", "ps", "-q", "-f", "name="+cfg.ContainerName)
	output, err := checkCmd.Output()
	if err != nil || len(output) == 0 {
		// Контейнер не запущен, стартуем его
		debugPrintf(cfg.Debug, "Контейнер не запущен, запускаю...\n")
		if err := startContainer(cfg, dbName); err != nil {
			return fmt.Errorf("не удалось запустить контейнер: %w", err)
		}

		// Ждём, пока PostgreSQL в контейнере поднимется
		if err := waitForPostgres(cfg); err != nil {
			return fmt.Errorf("PostgreSQL не поднялся: %w", err)
		}
	}

	// Проверяем существование файла дампа
	if _, err := os.Stat(dumpFile); os.IsNotExist(err) {
		return fmt.Errorf("файл дампа не найден: %s", dumpFile)
	}

	debugPrintf(cfg.Debug, "Загружаю дамп в базу %s...\n", dbName)

	// Выводим размер дампа только в режиме debug
	if cfg.Debug {
		statCmd := exec.Command("ls", "-l", dumpFile)
		statCmd.Stdout = os.Stdout
		debugCmd(statCmd, cfg.Debug)
	}

	// Удаляем только текущую базу данных, если она существует
	dropCmd := exec.Command("bash", "-c",
		fmt.Sprintf("docker exec %s dropdb -U postgres --if-exists %s",
			cfg.ContainerName, dbName))
	if err := debugCmd(dropCmd, cfg.Debug); err != nil {
		debugPrintf(cfg.Debug, "Предупреждение: не удалось удалить старую базу: %v\n", err)
	}

	// Создаем новую базу
	createCmd := exec.Command("bash", "-c",
		fmt.Sprintf("docker exec %s createdb -U postgres %s",
			cfg.ContainerName, dbName))
	if err := debugCmd(createCmd, cfg.Debug); err != nil {
		return fmt.Errorf("ошибка создания базы данных: %w", err)
	}

	// Восстанавливаем данные
	restoreCmd := exec.Command("bash", "-c",
		fmt.Sprintf("docker exec -i %s psql -U postgres -d %s < %s",
			cfg.ContainerName, dbName, dumpFile))
	if cfg.Debug {
		restoreCmd.Stdout = os.Stdout
		restoreCmd.Stderr = os.Stderr
	}

	if err := restoreCmd.Run(); err != nil {
		return fmt.Errorf("ошибка при восстановлении SQL: %w", err)
	}

	// Проверяем количество таблиц
	statsCmd := exec.Command("bash", "-c",
		fmt.Sprintf("docker exec %s psql -U postgres -d %s -t -c \""+
			"SELECT COUNT(*) FROM pg_tables WHERE schemaname = 'public';\"",
			cfg.ContainerName, dbName))
	output, err = statsCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ошибка при подсчете таблиц: %w\nВывод: %s", err, string(output))
	}

	count := strings.TrimSpace(string(output))
	debugPrintf(cfg.Debug, "\nКоличество таблиц в базе данных: %s\n", count)

	if count == "0" {
		return fmt.Errorf("после восстановления в базе нет таблиц")
	}

	return nil
}

// startContainer поднимает новый контейнер PostgreSQL
func startContainer(cfg PostgresConfig, dbName string) error {
	cmd := exec.Command("docker", "run", "-d",
		"--name", cfg.ContainerName,
		"-e", "POSTGRES_PASSWORD="+cfg.Password,
		"-e", fmt.Sprintf("POSTGRES_DB=%s", dbName),
		"-p", "5432:5432",
		cfg.Image,
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	debugPrintf(cfg.Debug, "Запускаю контейнер %s...\n", cfg.ContainerName)
	return cmd.Run()
}

// waitForPostgres ждёт, пока PostgreSQL внутри контейнера станет доступен
func waitForPostgres(cfg PostgresConfig) error {
	debugPrintf(cfg.Debug, "Ожидаю, пока PostgreSQL примет соединения (pg_isready)...\n")
	for i := 0; i < cfg.MaxWaitSeconds; i++ {
		checkCmd := exec.Command("docker", "exec", cfg.ContainerName, "pg_isready", "-U", "postgres")
		if err := checkCmd.Run(); err == nil {
			debugPrintf(cfg.Debug, "PostgreSQL готов!\n")
			return nil
		}
		time.Sleep(time.Second)
	}
	return fmt.Errorf("превышено %d секунд ожидания", cfg.MaxWaitSeconds)
}
