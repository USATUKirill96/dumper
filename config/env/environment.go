package env

// Environment represents a database environment configuration
type Environment struct {
	Name          string `yaml:"name"`
	DbDsn         string `yaml:"db_dsn"`
	MigrationsDir string `yaml:"migrations_dir"`
}

// Config represents a list of environments
type Config struct {
	Environments []Environment `yaml:"environments"`
}

// GetEnvironmentByName returns an environment by its name
func (c *Config) GetEnvironmentByName(name string) *Environment {
	for i := range c.Environments {
		if c.Environments[i].Name == name {
			return &c.Environments[i]
		}
	}
	return nil
}
