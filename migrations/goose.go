package migrations

import (
	"database/sql"
	"fmt"
	"path/filepath"
	"sort"

	_ "github.com/lib/pq" // PostgreSQL driver
	"github.com/pressly/goose/v3"
)

type MigrationStatus struct {
	ID        int64
	Name      string // full file name with path
	ShortName string // file name without path
	Applied   bool
	Timestamp int64
}

// Logger for intercepting goose output
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

// GetMigrationStatus returns status of all migrations
func GetMigrationStatus(dbDsn string, migrationsDir string, onLog func(string, ...interface{})) ([]MigrationStatus, error) {
	if migrationsDir == "" {
		return nil, nil
	}

	// Set up logger
	goose.SetLogger(&Logger{onLog: onLog})

	// Check directory exists
	absPath, err := filepath.Abs(migrationsDir)
	if err != nil {
		return nil, fmt.Errorf("error getting absolute path: %w", err)
	}

	// Connect to database
	db, err := sql.Open("postgres", dbDsn)
	if err != nil {
		return nil, fmt.Errorf("error connecting to database: %w", err)
	}
	defer db.Close()

	// Get all migrations
	migrations, err := goose.CollectMigrations(absPath, 0, goose.MaxVersion)
	if err != nil {
		return nil, fmt.Errorf("error reading migrations: %w", err)
	}

	// Get applied migrations
	current, err := goose.GetDBVersion(db)
	if err != nil {
		return nil, fmt.Errorf("error getting DB version: %w", err)
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

	// Sort by timestamp
	sort.Slice(result, func(i, j int) bool {
		return result[i].Timestamp < result[j].Timestamp
	})

	return result, nil
}

// MigrateTo migrates database to specified version
func MigrateTo(dbDsn string, migrationsDir string, targetVersion int64, onLog func(string, ...interface{})) error {
	// Set up logger
	goose.SetLogger(&Logger{onLog: onLog})

	// Check directory exists
	absPath, err := filepath.Abs(migrationsDir)
	if err != nil {
		return fmt.Errorf("error getting absolute path: %w", err)
	}

	// Connect to database
	db, err := sql.Open("postgres", dbDsn)
	if err != nil {
		return fmt.Errorf("error connecting to database: %w", err)
	}
	defer db.Close()

	// Get current version
	currentVersion, err := goose.GetDBVersion(db)
	if err != nil {
		return fmt.Errorf("error getting current version: %w", err)
	}

	// Choose migration direction
	if targetVersion > currentVersion {
		// Migrate up
		if err := goose.UpTo(db, absPath, targetVersion); err != nil {
			return fmt.Errorf("error migrating up: %w", err)
		}
	} else if targetVersion < currentVersion {
		// Migrate down
		if err := goose.DownTo(db, absPath, targetVersion); err != nil {
			return fmt.Errorf("error migrating down: %w", err)
		}
	}

	return nil
}
