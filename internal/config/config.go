package config

// Config represents the complete Unifidoc configuration
type Config struct {
	Project   ProjectConfig   `yaml:"project" json:"project"`
	Output    OutputConfig    `yaml:"output" json:"output"`
	Protocols []ProtocolConfig `yaml:"protocols" json:"protocols"`
	Features  FeaturesConfig  `yaml:"features" json:"features"`
	Server    ServerConfig    `yaml:"server" json:"server"`
}

// ProjectConfig represents project-level configuration
type ProjectConfig struct {
	Name        string `yaml:"name" json:"name"`
	Version     string `yaml:"version" json:"version"`
	Description string `yaml:"description" json:"description"`
	Logo        string `yaml:"logo,omitempty" json:"logo,omitempty"`
}

// OutputConfig represents output configuration
type OutputConfig struct {
	Dir    string `yaml:"dir" json:"dir"`
	Format string `yaml:"format" json:"format"` // html (only supported for now)
	Theme  string `yaml:"theme" json:"theme"`   // default, dark, or path to custom
}

// ProtocolConfig represents a protocol specification configuration
type ProtocolConfig struct {
	Plugin   string `yaml:"plugin" json:"plugin"`
	Spec     string `yaml:"spec" json:"spec"`
	Template string `yaml:"template,omitempty" json:"template,omitempty"`
	Enabled  bool   `yaml:"enabled" json:"enabled"`
}

// FeaturesConfig represents feature flags
type FeaturesConfig struct {
	Search       bool              `yaml:"search" json:"search"`
	Interactive  bool              `yaml:"interactive" json:"interactive"`
	DarkMode     bool              `yaml:"darkMode" json:"darkMode"`
	CodeSnippets CodeSnippetsConfig `yaml:"codeSnippets" json:"codeSnippets"`
}

// CodeSnippetsConfig represents code snippet configuration
type CodeSnippetsConfig struct {
	Languages []string `yaml:"languages" json:"languages"`
}

// ServerConfig represents dev server configuration
type ServerConfig struct {
	Port       int  `yaml:"port" json:"port"`
	LiveReload bool `yaml:"livereload" json:"livereload"`
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return &Config{
		Project: ProjectConfig{
			Name:        "API Documentation",
			Version:     "1.0.0",
			Description: "Unified API documentation",
		},
		Output: OutputConfig{
			Dir:    "./docs",
			Format: "html",
			Theme:  "default",
		},
		Protocols: []ProtocolConfig{},
		Features: FeaturesConfig{
			Search:      true,
			Interactive: false,
			DarkMode:    true,
			CodeSnippets: CodeSnippetsConfig{
				Languages: []string{"curl", "javascript", "python", "go"},
			},
		},
		Server: ServerConfig{
			Port:       8080,
			LiveReload: true,
		},
	}
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Project.Name == "" {
		c.Project.Name = "API Documentation"
	}

	if c.Output.Dir == "" {
		c.Output.Dir = "./docs"
	}

	if c.Output.Format == "" {
		c.Output.Format = "html"
	}

	if c.Output.Theme == "" {
		c.Output.Theme = "default"
	}

	if c.Server.Port == 0 {
		c.Server.Port = 8080
	}

	return nil
}
