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

// debugPrintf prints debug information only if debug mode is enabled
func debugPrintf(debug bool, format string, a ...interface{}) {
	if debug {
		fmt.Printf(format, a...)
	}
}

// debugCmd executes command with stdout/stderr output only in debug mode
func debugCmd(cmd *exec.Cmd, debug bool) error {
	if debug {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	return cmd.Run()
}

// DumpDatabase creates a database dump
func DumpDatabase(dsn string, dumpFile string, debug bool) error {
	// First check if there are tables in source database
	checkCmd := exec.Command("bash", "-c",
		fmt.Sprintf("psql '%s' -t -c \"SELECT COUNT(*) FROM pg_tables WHERE schemaname = 'public';\"",
			dsn))
	output, err := checkCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error checking tables in source database: %w\nOutput: %s", err, string(output))
	}

	// Create dump with additional options
	cmdString := fmt.Sprintf(
		"pg_dump "+
			"--format=plain "+ // Plain text format
			"--no-owner "+ // No owners
			"--no-privileges "+ // No privileges
			"--clean "+ // Clean before restore
			"--if-exists "+ // Add IF EXISTS to DROP
			"--schema=public "+ // Only public schema
			"--disable-triggers "+ // Disable triggers during restore
			"'%s' > %s",
		dsn, dumpFile)

	cmd := exec.Command("bash", "-c", cmdString)
	if debug {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error creating dump: %w", err)
	}

	return nil
}

// LoadDump loads dump into local database
func LoadDump(cfg PostgresConfig, dbName string, dumpFile string) error {
	// Check if container is running
	checkCmd := exec.Command("docker", "ps", "-q", "-f", "name="+cfg.ContainerName)
	output, err := checkCmd.Output()
	if err != nil || len(output) == 0 {
		// Container is not running, start it
		debugPrintf(cfg.Debug, "Container is not running, starting...\n")
		if err := startContainer(cfg, dbName); err != nil {
			return fmt.Errorf("failed to start container: %w", err)
		}

		// Wait for PostgreSQL to start in container
		if err := waitForPostgres(cfg); err != nil {
			return fmt.Errorf("PostgreSQL failed to start: %w", err)
		}
	}

	// Check if dump file exists
	if _, err := os.Stat(dumpFile); os.IsNotExist(err) {
		return fmt.Errorf("dump file not found: %s", dumpFile)
	}

	debugPrintf(cfg.Debug, "Loading dump into database %s...\n", dbName)

	// Show dump size only in debug mode
	if cfg.Debug {
		statCmd := exec.Command("ls", "-l", dumpFile)
		statCmd.Stdout = os.Stdout
		debugCmd(statCmd, cfg.Debug)
	}

	// Drop current database if it exists
	dropCmd := exec.Command("bash", "-c",
		fmt.Sprintf("docker exec %s dropdb -U postgres --if-exists %s",
			cfg.ContainerName, dbName))
	if err := debugCmd(dropCmd, cfg.Debug); err != nil {
		debugPrintf(cfg.Debug, "Warning: failed to drop old database: %v\n", err)
	}

	// Create new database
	createCmd := exec.Command("bash", "-c",
		fmt.Sprintf("docker exec %s createdb -U postgres %s",
			cfg.ContainerName, dbName))
	if err := debugCmd(createCmd, cfg.Debug); err != nil {
		return fmt.Errorf("error creating database: %w", err)
	}

	// Restore data
	restoreCmd := exec.Command("bash", "-c",
		fmt.Sprintf("docker exec -i %s psql -U postgres -d %s < %s",
			cfg.ContainerName, dbName, dumpFile))
	if cfg.Debug {
		restoreCmd.Stdout = os.Stdout
		restoreCmd.Stderr = os.Stderr
	}

	if err := restoreCmd.Run(); err != nil {
		return fmt.Errorf("error restoring SQL: %w", err)
	}

	// Check number of tables
	statsCmd := exec.Command("bash", "-c",
		fmt.Sprintf("docker exec %s psql -U postgres -d %s -t -c \""+
			"SELECT COUNT(*) FROM pg_tables WHERE schemaname = 'public';\"",
			cfg.ContainerName, dbName))
	output, err = statsCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error counting tables: %w\nOutput: %s", err, string(output))
	}

	count := strings.TrimSpace(string(output))
	debugPrintf(cfg.Debug, "\nNumber of tables in database: %s\n", count)

	if count == "0" {
		return fmt.Errorf("no tables in database after restore")
	}

	return nil
}

// startContainer starts a new PostgreSQL container
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

	debugPrintf(cfg.Debug, "Starting container %s...\n", cfg.ContainerName)
	return cmd.Run()
}

// waitForPostgres waits until PostgreSQL in container becomes available
func waitForPostgres(cfg PostgresConfig) error {
	debugPrintf(cfg.Debug, "Waiting for PostgreSQL to accept connections (pg_isready)...\n")
	for i := 0; i < cfg.MaxWaitSeconds; i++ {
		checkCmd := exec.Command("docker", "exec", cfg.ContainerName, "pg_isready", "-U", "postgres")
		if err := checkCmd.Run(); err == nil {
			debugPrintf(cfg.Debug, "PostgreSQL is ready!\n")
			return nil
		}
		time.Sleep(time.Second)
	}
	return fmt.Errorf("timeout after %d seconds", cfg.MaxWaitSeconds)
}
