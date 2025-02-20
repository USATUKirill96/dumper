package config

import (
	"fmt"
	"net/url"
	"os"
	"strings"

	"gopkg.in/yaml.v2"
)

type Environment struct {
	Name          string `yaml:"name"`
	DbDsn         string `yaml:"db_dsn"`
	MigrationsDir string `yaml:"migrations_dir"`
}

type DbConnection struct {
	Host     string
	Port     string
	User     string
	Password string
	Database string
	SslMode  string
}

type Config struct {
	Environments []Environment `yaml:"environments"`
}

// LoadConfig загружает конфигурацию из файла
func LoadConfig(configPath string, dumpsDir string) (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения конфига: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("ошибка парсинга конфига: %w", err)
	}

	if len(config.Environments) == 0 {
		return nil, fmt.Errorf("не найдено ни одного окружения в конфиге")
	}

	// Создаем директорию для дампов, если её нет
	if err := os.MkdirAll(dumpsDir, 0755); err != nil {
		return nil, fmt.Errorf("ошибка создания директории для дампов: %w", err)
	}

	return &config, nil
}

// ParseDSN парсит строку подключения к базе данных
func ParseDSN(dsn string) (*DbConnection, error) {
	u, err := url.Parse(dsn)
	if err != nil {
		return nil, fmt.Errorf("ошибка парсинга DSN: %w", err)
	}

	password, _ := u.User.Password()
	query := u.Query()

	return &DbConnection{
		Host:     u.Hostname(),
		Port:     u.Port(),
		User:     u.User.Username(),
		Password: password,
		Database: strings.TrimPrefix(u.Path, "/"),
		SslMode:  query.Get("sslmode"),
	}, nil
}

// GetDSN формирует строку подключения из параметров
func (c *DbConnection) GetDSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		c.User, c.Password, c.Host, c.Port, c.Database, c.SslMode)
}
