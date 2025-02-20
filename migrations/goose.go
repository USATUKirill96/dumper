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

// GetMigrationStatus возвращает статус всех миграций
func GetMigrationStatus(dbDsn string, migrationsDir string) ([]MigrationStatus, error) {
	if migrationsDir == "" {
		return nil, nil
	}

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
