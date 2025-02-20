package migrations

import (
	"database/sql"
	"fmt"
	"path/filepath"
	"sort"

	_ "github.com/lib/pq" // драйвер PostgreSQL
	"github.com/pressly/goose/v3"
)

type MigrationStatus struct {
	ID        int64
	Name      string // полное имя файла с путём
	ShortName string // только имя файла без пути
	Applied   bool
	Timestamp int64
}

// Logger для перехвата вывода goose
type Logger struct {
	onLog func(string, ...interface{})
}

func (l *Logger) Fatalf(format string, v ...interface{}) {
	if l.onLog != nil {
		l.onLog(format, v...)
	}
}

func (l *Logger) Printf(format string, v ...interface{}) {
	if l.onLog != nil {
		l.onLog(format, v...)
	}
}

// GetMigrationStatus возвращает статус всех миграций
func GetMigrationStatus(dbDsn string, migrationsDir string, onLog func(string, ...interface{})) ([]MigrationStatus, error) {
	if migrationsDir == "" {
		return nil, nil
	}

	// Настраиваем логгер
	goose.SetLogger(&Logger{onLog: onLog})

	// Проверяем существование директории
	absPath, err := filepath.Abs(migrationsDir)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения абсолютного пути: %w", err)
	}

	// Подключаемся к базе
	db, err := sql.Open("postgres", dbDsn)
	if err != nil {
		return nil, fmt.Errorf("ошибка подключения к базе: %w", err)
	}
	defer db.Close()

	// Получаем список всех миграций
	migrations, err := goose.CollectMigrations(absPath, 0, goose.MaxVersion)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения миграций: %w", err)
	}

	// Получаем примененные миграции
	current, err := goose.GetDBVersion(db)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения версии БД: %w", err)
	}

	var result []MigrationStatus
	for _, m := range migrations {
		result = append(result, MigrationStatus{
			ID:        m.Version,
			Name:      m.Source,
			ShortName: filepath.Base(m.Source),
			Applied:   m.Version <= current,
			Timestamp: m.Version,
		})
	}

	// Сортируем по timestamp
	sort.Slice(result, func(i, j int) bool {
		return result[i].Timestamp < result[j].Timestamp
	})

	return result, nil
}

// MigrateTo выполняет миграцию базы данных до указанной версии
func MigrateTo(dbDsn string, migrationsDir string, targetVersion int64, onLog func(string, ...interface{})) error {
	// Настраиваем логгер
	goose.SetLogger(&Logger{onLog: onLog})

	// Проверяем существование директории
	absPath, err := filepath.Abs(migrationsDir)
	if err != nil {
		return fmt.Errorf("ошибка получения абсолютного пути: %w", err)
	}

	// Подключаемся к базе
	db, err := sql.Open("postgres", dbDsn)
	if err != nil {
		return fmt.Errorf("ошибка подключения к базе: %w", err)
	}
	defer db.Close()

	// Получаем текущую версию
	currentVersion, err := goose.GetDBVersion(db)
	if err != nil {
		return fmt.Errorf("ошибка получения текущей версии: %w", err)
	}

	// Выбираем направление миграции
	if targetVersion > currentVersion {
		// Миграция вперед
		if err := goose.UpTo(db, absPath, targetVersion); err != nil {
			return fmt.Errorf("ошибка выполнения миграции вперед: %w", err)
		}
	} else if targetVersion < currentVersion {
		// Откат назад
		if err := goose.DownTo(db, absPath, targetVersion); err != nil {
			return fmt.Errorf("ошибка отката миграции: %w", err)
		}
	}

	return nil
}
