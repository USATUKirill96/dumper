package db

import (
	"fmt"
	"net/url"
	"strings"
)

// Connection represents database connection parameters
type Connection struct {
	Host     string
	Port     string
	User     string
	Password string
	Database string
	SslMode  string
}

// ParseDSN parses a database connection string into Connection struct
func ParseDSN(dsn string) (*Connection, error) {
	u, err := url.Parse(dsn)
	if err != nil {
		return nil, fmt.Errorf("error parsing DSN: %w", err)
	}

	password, _ := u.User.Password()
	query := u.Query()

	return &Connection{
		Host:     u.Hostname(),
		Port:     u.Port(),
		User:     u.User.Username(),
		Password: password,
		Database: strings.TrimPrefix(u.Path, "/"),
		SslMode:  query.Get("sslmode"),
	}, nil
}

// GetDSN returns a formatted connection string
func (c *Connection) GetDSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		c.User, c.Password, c.Host, c.Port, c.Database, c.SslMode)
}

// Clone creates a copy of the connection with optional parameter overrides
func (c *Connection) Clone(overrides map[string]string) *Connection {
	conn := &Connection{
		Host:     c.Host,
		Port:     c.Port,
		User:     c.User,
		Password: c.Password,
		Database: c.Database,
		SslMode:  c.SslMode,
	}

	if host, ok := overrides["host"]; ok {
		conn.Host = host
	}
	if port, ok := overrides["port"]; ok {
		conn.Port = port
	}
	if user, ok := overrides["user"]; ok {
		conn.User = user
	}
	if password, ok := overrides["password"]; ok {
		conn.Password = password
	}
	if database, ok := overrides["database"]; ok {
		conn.Database = database
	}
	if sslMode, ok := overrides["sslmode"]; ok {
		conn.SslMode = sslMode
	}

	return conn
}
