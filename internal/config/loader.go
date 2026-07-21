package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Loader handles configuration loading
type Loader struct {
	configPath string
}

// NewLoader creates a new configuration loader
func NewLoader() *Loader {
	return &Loader{}
}

// SetConfigPath sets a specific config file path
func (l *Loader) SetConfigPath(path string) {
	l.configPath = path
}

// Load loads the configuration
func (l *Loader) Load() (*Config, error) {
	v := viper.New()

	// Set defaults
	v.SetDefault("output.dir", "./docs")
	v.SetDefault("output.format", "html")
	v.SetDefault("output.theme", "default")
	v.SetDefault("server.port", 8080)
	v.SetDefault("server.livereload", true)
	v.SetDefault("features.search", true)
	v.SetDefault("features.darkMode", true)
	v.SetDefault("features.interactive", false)

	// If specific config path is provided, use it
	if l.configPath != "" {
		v.SetConfigFile(l.configPath)
	} else {
		// Try to find config in priority order
		configPath, err := l.findConfig()
		if err != nil {
			// No config found, return default config
			return DefaultConfig(), nil
		}
		v.SetConfigFile(configPath)
	}

	// Read config
	if err := v.ReadInConfig(); err != nil {
		// If config file not found, return default
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			return DefaultConfig(), nil
		}
		return nil, fmt.Errorf("error reading config: %w", err)
	}

	// Unmarshal into Config struct
	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	// Validate config
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &config, nil
}

// findConfig searches for config files in priority order
func (l *Loader) findConfig() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	// Priority order from PRD
	candidates := []string{
		filepath.Join(cwd, "unifidoc.yaml"),
		filepath.Join(cwd, ".unifidoc.yaml"),
		filepath.Join(cwd, "docs", "unifidoc.yaml"),
		filepath.Join(cwd, ".unifidoc", "config.yaml"),
		filepath.Join(cwd, "unifidoc.yml"),
		filepath.Join(cwd, ".unifidoc.yml"),
	}

	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
	}

	return "", fmt.Errorf("no config file found")
}

// LoadFromFile loads configuration from a specific file
func LoadFromFile(path string) (*Config, error) {
	loader := NewLoader()
	loader.SetConfigPath(path)
	return loader.Load()
}

// LoadDefault loads configuration with auto-detection
func LoadDefault() (*Config, error) {
	loader := NewLoader()
	return loader.Load()
}
